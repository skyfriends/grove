package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Repo struct {
	Name      string
	Path      string
	Branch    string
	Default   string
	Dirty     bool
	Worktrees []Worktree
}

type Worktree struct {
	RepoName string
	RepoPath string
	Path     string
	Dir      string
	Branch   string
	Dirty    bool
}

// git runs a git command and returns trimmed stdout. Errors are silently ignored.
func git(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

// gitOk runs a git command and returns whether it succeeded.
func gitOk(dir string, args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	return cmd.Run() == nil
}

// gitRun runs a git command and returns combined output + error.
func gitRun(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func defaultBranch(path string) string {
	ref := git(path, "symbolic-ref", "refs/remotes/origin/HEAD")
	if ref != "" {
		return strings.TrimPrefix(ref, "refs/remotes/origin/")
	}
	for _, b := range []string{"dev", "main", "master"} {
		if gitOk(path, "show-ref", "--verify", "--quiet", "refs/heads/"+b) {
			return b
		}
	}
	return "main"
}

func isDirty(path string) bool {
	out := git(path, "status", "--porcelain")
	return out != ""
}

func parseWorktrees(repoName, repoPath string) []Worktree {
	out := git(repoPath, "worktree", "list")
	lines := strings.Split(out, "\n")

	var wts []Worktree
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		wtPath := parts[0]
		branch := ""
		for _, p := range parts {
			if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
				branch = p[1 : len(p)-1]
			}
		}
		wts = append(wts, Worktree{
			RepoName: repoName,
			RepoPath: repoPath,
			Path:     wtPath,
			Dir:      filepath.Base(wtPath),
			Branch:   branch,
		})
	}

	// dirty checks in parallel
	var wg sync.WaitGroup
	wg.Add(len(wts))
	for i := range wts {
		go func(idx int) {
			defer wg.Done()
			wts[idx].Dirty = isDirty(wts[idx].Path)
		}(i)
	}
	wg.Wait()
	return wts
}

func scanRepo(name, path string) Repo {
	repo := Repo{Name: name, Path: path}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() { defer wg.Done(); repo.Branch = git(path, "branch", "--show-current") }()
	go func() { defer wg.Done(); repo.Default = defaultBranch(path) }()
	go func() { defer wg.Done(); repo.Dirty = isDirty(path) }()
	go func() { defer wg.Done(); repo.Worktrees = parseWorktrees(name, path) }()
	wg.Wait()
	return repo
}

func findRepos() []Repo {
	entries, err := os.ReadDir(cfg.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error reading %s: %v\n", cfg.Root, err)
		os.Exit(1)
	}

	type entry struct{ name, path string }
	var toScan []entry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(cfg.Root, e.Name())
		info, err := os.Stat(filepath.Join(path, ".git"))
		if err != nil || !info.IsDir() {
			continue // skip non-repos and worktrees (.git file instead of dir)
		}
		toScan = append(toScan, entry{e.Name(), path})
	}

	repos := make([]Repo, len(toScan))
	var wg sync.WaitGroup
	wg.Add(len(toScan))
	for i, e := range toScan {
		go func(idx int, name, path string) {
			defer wg.Done()
			repos[idx] = scanRepo(name, path)
		}(i, e.name, e.path)
	}
	wg.Wait()
	return repos
}
