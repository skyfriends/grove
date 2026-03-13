package main

import (
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	cfg = loadConfig()

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "", "list", "ls", "l":
		cmdList()
	case "status", "st":
		cmdStatus()
	case "stale":
		cmdStale()
	case "sync", "s":
		cmdSync()
	case "prune", "p":
		cmdPrune()
	case "clean", "c":
		cmdClean()
	case "nuke":
		cmdNuke()
	case "reset", "r":
		cmdReset()
	case "wt":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "  usage: grove wt <repo> <branch>")
			os.Exit(1)
		}
		cmdWt(os.Args[2], os.Args[3])
	case "clone":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "  usage: grove clone <url|org/repo>")
			os.Exit(1)
		}
		cmdClone(os.Args[2])
	case "open", "o":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "  usage: grove open <repo>")
			os.Exit(1)
		}
		cmdOpen(os.Args[2])
	case "init":
		cmdInit()
	case "version", "-v", "--version":
		fmt.Printf("  grove %s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "  unknown command: %s (try: grove help)\n", cmd)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println()
	fmt.Println("  " + white.Render("grove") + "  " + dim.Render("manage local repositories and worktrees"))
	fmt.Println()
	fmt.Println("  " + dim.Render("browse"))
	fmt.Println("    " + blue.Render("grove") + "              list repos, branches, and worktrees")
	fmt.Println("    " + blue.Render("grove status") + "       show ahead/behind counts for all repos")
	fmt.Println("    " + blue.Render("grove stale") + "        find and interactively delete stale branches")
	fmt.Println()
	fmt.Println("  " + dim.Render("maintain"))
	fmt.Println("    " + blue.Render("grove sync") + "         pull latest on repos at their default branch")
	fmt.Println("    " + blue.Render("grove prune") + "        remove stale remote-tracking references")
	fmt.Println("    " + blue.Render("grove clean") + "        interactively remove worktrees")
	fmt.Println("    " + blue.Render("grove nuke") + "         remove all clean worktrees for selected repos")
	fmt.Println("    " + blue.Render("grove reset") + "        switch repos back to their default branch")
	fmt.Println()
	fmt.Println("  " + dim.Render("create"))
	fmt.Println("    " + blue.Render("grove wt") + " <repo> <branch>    create a new worktree")
	fmt.Println("    " + blue.Render("grove clone") + " <repo>           clone into root directory")
	fmt.Println("    " + blue.Render("grove open") + " <repo>            open repo in editor")
	fmt.Println()
	fmt.Println("  " + dim.Render("setup"))
	fmt.Println("    " + blue.Render("grove init") + "         create ~/.grove.toml")
	fmt.Println("    " + blue.Render("grove version") + "      print version")
	fmt.Println()
	fmt.Println("  " + dim.Render("config: ~/.grove.toml"))
	fmt.Println()
}
