package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/huh"
)

// ── list ────────────────────────────────────────────────────────────────────

func cmdList() {
	repos := findRepos()

	var offDefault, withWt, dirtyCount, totalWt int
	for _, r := range repos {
		if r.Branch != r.Default {
			offDefault++
		}
		if len(r.Worktrees) > 0 {
			withWt++
			totalWt += len(r.Worktrees)
		}
		if r.Dirty {
			dirtyCount++
		}
	}

	fmt.Println()
	fmt.Printf("  %s  %s  %s  %s  %s\n",
		blue.Render(cfg.Root),
		white.Render(fmt.Sprintf("%d repos", len(repos))),
		yellow.Render(fmt.Sprintf("%d off default", offDefault)),
		muted.Render(fmt.Sprintf("%d worktrees", totalWt)),
		red.Render(fmt.Sprintf("%d dirty", dirtyCount)),
	)
	fmt.Println("  " + divider)
	fmt.Println()

	for _, r := range repos {
		branch := r.Branch
		if branch == "" {
			branch = "detached"
		}
		nameStr := white.Render(pad(r.Name, 26))
		bs := green
		if branch != r.Default {
			bs = yellow
		}
		extras := ""
		if r.Dirty {
			extras += "  " + red.Render("dirty")
		}
		if len(r.Worktrees) > 0 {
			extras += "  " + dim.Render(fmt.Sprintf("%d wt", len(r.Worktrees)))
		}
		fmt.Printf("  %s%s%s\n", nameStr, bs.Render(branch), extras)

		for i, wt := range r.Worktrees {
			connector := "├─"
			if i == len(r.Worktrees)-1 {
				connector = "└─"
			}
			wtDirty := ""
			if wt.Dirty {
				wtDirty = " " + red.Render("dirty")
			}
			fmt.Printf("     %s %s  %s%s\n",
				dimmer.Render(connector),
				muted.Render(wt.Branch),
				dim.Render(wt.Dir),
				wtDirty,
			)
		}
	}
	fmt.Println()
}

// ── status ──────────────────────────────────────────────────────────────────

func cmdStatus() {
	repos := findRepos()

	// fetch tracking info in parallel
	tracking := make([]string, len(repos))
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for i, r := range repos {
		go func(idx int, repo Repo) {
			defer wg.Done()
			tracking[idx] = git(repo.Path, "for-each-ref",
				"--format=%(upstream:track)", "refs/heads/"+repo.Branch)
		}(i, r)
	}
	wg.Wait()

	fmt.Println()
	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("%d repos", len(repos))))
	fmt.Println("  " + divider)
	fmt.Println()

	for i, r := range repos {
		nameStr := white.Render(pad(r.Name, 26))
		bs := green
		if r.Branch != r.Default {
			bs = yellow
		}
		extras := ""
		if r.Dirty {
			extras += "  " + red.Render("dirty")
		}
		if tracking[i] != "" {
			extras += "  " + blue.Render(tracking[i])
		}
		fmt.Printf("  %s%s%s\n", nameStr, bs.Render(r.Branch), extras)
	}
	fmt.Println()
}

// ── stale ───────────────────────────────────────────────────────────────────

func cmdStale() {
	repos := findRepos()
	threshold := time.Now().AddDate(0, 0, -cfg.StaleDays)

	type staleBranch struct {
		branch string
		days   int
	}
	type repoStale struct {
		name     string
		branches []staleBranch
	}

	results := make([]repoStale, len(repos))
	var wg sync.WaitGroup
	wg.Add(len(repos))
	for i, r := range repos {
		go func(idx int, repo Repo) {
			defer wg.Done()
			out := git(repo.Path, "for-each-ref", "--sort=committerdate",
				"--format=%(refname:short)\t%(committerdate:unix)", "refs/heads/")
			if out == "" {
				return
			}
			var stale []staleBranch
			for _, line := range strings.Split(out, "\n") {
				parts := strings.SplitN(line, "\t", 2)
				if len(parts) != 2 {
					continue
				}
				branch := parts[0]
				if branch == repo.Default {
					continue
				}
				unix, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					continue
				}
				commitTime := time.Unix(unix, 0)
				if commitTime.Before(threshold) {
					days := int(time.Since(commitTime).Hours() / 24)
					stale = append(stale, staleBranch{branch, days})
				}
			}
			results[idx] = repoStale{repo.Name, stale}
		}(i, r)
	}
	wg.Wait()

	fmt.Println()
	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("branches with no activity in %d+ days", cfg.StaleDays)))
	fmt.Println("  " + divider)
	fmt.Println()

	found := false
	for _, r := range results {
		if len(r.branches) == 0 {
			continue
		}
		found = true
		fmt.Printf("  %s\n", white.Render(r.name))
		for _, b := range r.branches {
			fmt.Printf("    %s  %s\n",
				yellow.Render(pad(b.branch, 36)),
				dim.Render(fmt.Sprintf("%dd", b.days)))
		}
		fmt.Println()
	}

	if !found {
		fmt.Println("  " + dim.Render("no stale branches"))
		fmt.Println()
	}
}

// ── sync ────────────────────────────────────────────────────────────────────

func cmdSync() {
	repos := findRepos()

	var candidates []Repo
	for _, r := range repos {
		if r.Branch == r.Default && !r.Dirty {
			candidates = append(candidates, r)
		}
	}

	if len(candidates) == 0 {
		fmt.Println()
		fmt.Println("  " + dim.Render("no clean repos on their default branch"))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("syncing %d repos", len(candidates))))
	fmt.Println("  " + divider)
	fmt.Println()

	type syncResult struct {
		name, branch, status string
	}

	results := make([]syncResult, len(candidates))
	var wg sync.WaitGroup
	wg.Add(len(candidates))
	for i, r := range candidates {
		go func(idx int, repo Repo) {
			defer wg.Done()
			out, err := gitRun(repo.Path, "pull", "--ff-only")
			res := syncResult{name: repo.Name, branch: repo.Branch}
			if err != nil {
				res.status = "failed"
			} else if strings.Contains(out, "Already up to date") {
				res.status = "current"
			} else {
				res.status = "pulled"
			}
			results[idx] = res
		}(i, r)
	}
	wg.Wait()

	var pulled, failed int
	for _, r := range results {
		label := pad(r.name, 24) + muted.Render(r.branch)
		switch r.status {
		case "pulled":
			fmt.Printf("  %s  %s\n", green.Render("↓"), label)
			pulled++
		case "current":
			fmt.Printf("  %s  %s\n", dim.Render("-"), label)
		case "failed":
			fmt.Printf("  %s  %s\n", red.Render("✗"), label)
			failed++
		}
	}

	fmt.Println()
	if pulled == 0 && failed == 0 {
		fmt.Println("  " + dim.Render("everything current"))
	}
	if pulled > 0 {
		fmt.Printf("  %s\n", green.Render(fmt.Sprintf("  %d pulled", pulled)))
	}
	if failed > 0 {
		fmt.Printf("  %s\n", red.Render(fmt.Sprintf("  %d failed", failed)))
	}
	fmt.Println()
}

// ── clean ───────────────────────────────────────────────────────────────────

func cmdClean() {
	repos := findRepos()

	var allWts []Worktree
	for _, r := range repos {
		allWts = append(allWts, r.Worktrees...)
	}

	if len(allWts) == 0 {
		fmt.Println()
		fmt.Println("  " + dim.Render("no worktrees found"))
		fmt.Println()
		return
	}

	repoSet := map[string]bool{}
	for _, wt := range allWts {
		repoSet[wt.RepoName] = true
	}

	fmt.Println()
	fmt.Printf("  %s across %s\n",
		white.Render(fmt.Sprintf("%d worktrees", len(allWts))),
		white.Render(fmt.Sprintf("%d repos", len(repoSet))),
	)
	fmt.Println("  " + divider)
	fmt.Println()

	var opts []huh.Option[int]
	for i, wt := range allWts {
		label := pad(wt.RepoName, 22) + wt.Branch
		if wt.Dirty {
			label += "  (dirty)"
		}
		opts = append(opts, huh.NewOption(label, i))
	}

	var selected []int
	height := len(opts) + 6
	if height > 28 {
		height = 28
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Remove worktrees").
				Description("space toggle  ·  a select all  ·  enter confirm  ·  esc cancel").
				Options(opts...).
				Value(&selected).
				Height(height),
		),
	).WithTheme(customTheme())

	if err := form.Run(); err != nil {
		return
	}

	if len(selected) == 0 {
		fmt.Println("  " + dim.Render("cancelled"))
		fmt.Println()
		return
	}

	fmt.Println()
	for _, idx := range selected {
		wt := allWts[idx]
		label := pad(wt.RepoName, 22) + muted.Render(wt.Branch)
		if wt.Dirty {
			fmt.Printf("  %s  %s\n", skipBadge.Render("SKIP"), label)
			continue
		}
		if gitOk(wt.RepoPath, "worktree", "remove", wt.Path) {
			exec.Command("git", "-C", wt.RepoPath, "branch", "-d", wt.Branch).Run()
			fmt.Printf("  %s    %s\n", okTag.Render("OK"), label)
		} else {
			fmt.Printf("  %s  %s\n", failBadge.Render("FAIL"), label)
		}
	}
	fmt.Println()
}

// ── nuke ────────────────────────────────────────────────────────────────────

func cmdNuke() {
	repos := findRepos()

	var withWt []Repo
	for _, r := range repos {
		if len(r.Worktrees) > 0 {
			withWt = append(withWt, r)
		}
	}

	if len(withWt) == 0 {
		fmt.Println()
		fmt.Println("  " + dim.Render("no repos with worktrees"))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("%d repos with worktrees", len(withWt))))
	fmt.Println("  " + divider)
	fmt.Println()

	for _, r := range withWt {
		dirtyCount := 0
		for _, wt := range r.Worktrees {
			if wt.Dirty {
				dirtyCount++
			}
		}
		cleanCount := len(r.Worktrees) - dirtyCount
		detail := green.Render(fmt.Sprintf("%d clean", cleanCount))
		if dirtyCount > 0 {
			detail += "  " + yellow.Render(fmt.Sprintf("%d dirty (kept)", dirtyCount))
		}
		fmt.Printf("  %-24s%s  %s\n", white.Render(r.Name), dim.Render(fmt.Sprintf("%d wt", len(r.Worktrees))), detail)
	}
	fmt.Println()

	opts := make([]huh.Option[int], len(withWt))
	for i, r := range withWt {
		dirtyCount := 0
		for _, wt := range r.Worktrees {
			if wt.Dirty {
				dirtyCount++
			}
		}
		label := pad(r.Name, 22) + fmt.Sprintf("%d worktrees", len(r.Worktrees))
		if dirtyCount > 0 {
			label += fmt.Sprintf("  (%d dirty, kept)", dirtyCount)
		}
		opts[i] = huh.NewOption(label, i)
	}

	var selected []int
	height := len(opts) + 6
	if height > 28 {
		height = 28
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Nuke worktrees").
				Description("removes all clean worktrees for selected repos  ·  dirty ones are kept").
				Options(opts...).
				Value(&selected).
				Height(height),
		),
	).WithTheme(customTheme())

	if err := form.Run(); err != nil {
		return
	}

	if len(selected) == 0 {
		fmt.Println("  " + dim.Render("cancelled"))
		fmt.Println()
		return
	}

	fmt.Println()
	for _, idx := range selected {
		r := withWt[idx]
		fmt.Printf("  %s\n", white.Render(r.Name))
		for _, wt := range r.Worktrees {
			label := "  " + pad(wt.Branch, 50)
			if wt.Dirty {
				fmt.Printf("  %s  %s\n", skipBadge.Render("KEPT"), label+dim.Render("dirty"))
				continue
			}
			if gitOk(wt.RepoPath, "worktree", "remove", wt.Path) {
				exec.Command("git", "-C", wt.RepoPath, "branch", "-d", wt.Branch).Run()
				fmt.Printf("  %s    %s\n", okTag.Render("OK"), label)
			} else {
				fmt.Printf("  %s  %s\n", failBadge.Render("FAIL"), label)
			}
		}
		fmt.Println()
	}
}

// ── reset ───────────────────────────────────────────────────────────────────

func cmdReset() {
	repos := findRepos()

	var offDefault []Repo
	for _, r := range repos {
		if r.Branch != r.Default {
			offDefault = append(offDefault, r)
		}
	}

	if len(offDefault) == 0 {
		fmt.Println()
		fmt.Println("  " + dim.Render("all repos on their default branch"))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("%d repos off default branch", len(offDefault))))
	fmt.Println("  " + divider)
	fmt.Println()

	opts := make([]huh.Option[int], len(offDefault))
	for i, r := range offDefault {
		label := pad(r.Name, 22) + r.Branch + "  →  " + r.Default
		if r.Dirty {
			label += "  (dirty)"
		}
		opts[i] = huh.NewOption(label, i)
	}

	var selected []int
	height := len(opts) + 6
	if height > 28 {
		height = 28
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Checkout default branch").
				Description("space toggle  ·  a select all  ·  enter confirm  ·  esc cancel").
				Options(opts...).
				Value(&selected).
				Height(height),
		),
	).WithTheme(customTheme())

	if err := form.Run(); err != nil {
		return
	}

	if len(selected) == 0 {
		fmt.Println("  " + dim.Render("cancelled"))
		fmt.Println()
		return
	}

	fmt.Println()
	for _, idx := range selected {
		r := offDefault[idx]
		label := pad(r.Name, 22) + muted.Render(r.Branch+" → "+r.Default)
		if r.Dirty {
			fmt.Printf("  %s  %s\n", skipBadge.Render("SKIP"), label)
			continue
		}
		if gitOk(r.Path, "checkout", r.Default) {
			fmt.Printf("  %s    %s\n", okTag.Render("OK"), label)
		} else {
			fmt.Printf("  %s  %s\n", failBadge.Render("FAIL"), label)
		}
	}
	fmt.Println()
}

// ── wt (worktree create) ───────────────────────────────────────────────────

func cmdWt(repoName, branch string) {
	repoPath := filepath.Join(cfg.Root, repoName)
	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "  repo not found: %s\n", repoName)
		os.Exit(1)
	}

	// generate unique worktree path
	sanitized := strings.NewReplacer("/", "-", "\\", "-").Replace(branch)
	var wtPath, wtDir string
	for {
		suffix := randomSuffix(6)
		wtDir = fmt.Sprintf("%s-%s.%s", repoName, sanitized, suffix)
		wtPath = filepath.Join(cfg.Root, wtDir)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			break
		}
	}

	// if branch exists, check it out; otherwise create from default
	var ok bool
	if gitOk(repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branch) {
		ok = gitOk(repoPath, "worktree", "add", wtPath, branch)
	} else {
		def := defaultBranch(repoPath)
		ok = gitOk(repoPath, "worktree", "add", "-b", branch, wtPath, def)
	}

	if !ok {
		fmt.Println()
		fmt.Printf("  %s  failed to create worktree\n", failBadge.Render("FAIL"))
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n", okTag.Render("OK"), white.Render(wtDir))
	fmt.Printf("       %s  %s\n", dim.Render("branch"), muted.Render(branch))
	fmt.Printf("       %s  %s\n", dim.Render("path  "), muted.Render(wtPath))
	fmt.Println()
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}

// ── clone ───────────────────────────────────────────────────────────────────

func cmdClone(target string) {
	url := target
	if !strings.Contains(target, "://") && !strings.HasPrefix(target, "git@") {
		url = "https://github.com/" + target + ".git"
	}

	name := strings.TrimSuffix(filepath.Base(url), ".git")
	dest := filepath.Join(cfg.Root, name)

	if _, err := os.Stat(dest); err == nil {
		fmt.Fprintf(os.Stderr, "  already exists: %s\n", name)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", dim.Render("cloning"), muted.Render(url))
	fmt.Println()

	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println()
		fmt.Printf("  %s  %s\n", failBadge.Render("FAIL"), red.Render(name))
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n", okTag.Render("OK"), white.Render(name))
	fmt.Println()
}

// ── doctor ──────────────────────────────────────────────────────────────────

func cmdDoctor() {
	repos := findRepos()

	type issue struct {
		repo   string
		detail string
	}

	missingOriginHead := make([]issue, len(repos))
	detachedHead := make([]issue, len(repos))
	noRemote := make([]issue, len(repos))
	forgottenWip := make([]issue, len(repos))

	var wg sync.WaitGroup
	wg.Add(len(repos))
	for i, r := range repos {
		go func(idx int, repo Repo) {
			defer wg.Done()

			// missing origin/HEAD
			ref := git(repo.Path, "symbolic-ref", "refs/remotes/origin/HEAD")
			hasRemote := git(repo.Path, "remote") != ""
			if hasRemote && ref == "" {
				missingOriginHead[idx] = issue{repo.Name, "origin/HEAD symref not set"}
			}

			// detached HEAD
			if repo.Branch == "" {
				detachedHead[idx] = issue{repo.Name, "HEAD is detached"}
			}

			// no remote configured
			if !hasRemote {
				noRemote[idx] = issue{repo.Name, "no remote configured"}
			}

			// dirty + off default branch = likely forgotten WIP
			if repo.Dirty && repo.Branch != "" && repo.Branch != repo.Default {
				forgottenWip[idx] = issue{repo.Name, repo.Branch}
			}
		}(i, r)
	}
	wg.Wait()

	// collect non-empty issues
	collect := func(raw []issue) []issue {
		var out []issue
		for _, iss := range raw {
			if iss.repo != "" {
				out = append(out, iss)
			}
		}
		return out
	}

	groups := []struct {
		label string
		items []issue
	}{
		{"missing origin/HEAD", collect(missingOriginHead)},
		{"detached HEAD", collect(detachedHead)},
		{"no remote configured", collect(noRemote)},
		{"uncommitted changes on non-default branch", collect(forgottenWip)},
	}

	total := 0
	for _, g := range groups {
		total += len(g.items)
	}

	fmt.Println()
	if total == 0 {
		fmt.Printf("  %s  %s\n", green.Render("✓"), white.Render("all repos healthy"))
		fmt.Printf("  %s\n", dim.Render(fmt.Sprintf("checked %d repos", len(repos))))
		fmt.Println()
		return
	}

	fmt.Printf("  %s\n", white.Render(fmt.Sprintf("%d issues across %d repos", total, len(repos))))
	fmt.Println("  " + divider)

	for _, g := range groups {
		if len(g.items) == 0 {
			continue
		}
		fmt.Println()
		fmt.Printf("  %s  %s\n", yellow.Render("●"), white.Render(g.label))
		for _, iss := range g.items {
			fmt.Printf("    %s  %s\n", muted.Render(iss.repo), dim.Render(iss.detail))
		}
	}
	fmt.Println()
}

// ── open ────────────────────────────────────────────────────────────────────

func cmdOpen(repoName string) {
	repoPath := filepath.Join(cfg.Root, repoName)

	// exact match first, then fuzzy
	if _, err := os.Stat(repoPath); err != nil {
		entries, _ := os.ReadDir(cfg.Root)
		lower := strings.ToLower(repoName)
		var matches []string
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), lower) {
				matches = append(matches, e.Name())
			}
		}
		if len(matches) == 1 {
			repoPath = filepath.Join(cfg.Root, matches[0])
			repoName = matches[0]
		} else if len(matches) > 1 {
			fmt.Println()
			fmt.Printf("  %s\n", yellow.Render("multiple matches:"))
			for _, m := range matches {
				fmt.Printf("    %s\n", muted.Render(m))
			}
			fmt.Println()
			return
		} else {
			fmt.Fprintf(os.Stderr, "  not found: %s\n", repoName)
			os.Exit(1)
		}
	}

	cmd := exec.Command(cfg.Editor, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s  %s %s\n", okTag.Render("OK"), dim.Render(cfg.Editor), white.Render(repoName))
	fmt.Println()
}
