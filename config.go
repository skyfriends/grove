package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all grove settings, loaded from ~/.grove.toml.
type Config struct {
	Root      string `toml:"root"`
	Editor    string `toml:"editor"`
	StaleDays int    `toml:"stale_days"`
}

var cfg Config

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".grove.toml")
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func loadConfig() Config {
	c := Config{
		Root:      "~/projects",
		StaleDays: 30,
	}

	data, err := os.ReadFile(configPath())
	if err == nil {
		if _, err := toml.Decode(string(data), &c); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", configPath(), err)
		}
	}

	c.Root = expandHome(c.Root)
	if c.StaleDays <= 0 {
		c.StaleDays = 30
	}
	if c.Editor == "" {
		c.Editor = os.Getenv("EDITOR")
	}
	if c.Editor == "" {
		c.Editor = "code"
	}
	return c
}

const defaultConfigTOML = `# Grove configuration

# Root directory containing your git repositories
root = "~/projects"

# Editor for 'grove open' (defaults to $EDITOR, then "code")
# editor = "code"

# Branches inactive for this many days are flagged by 'grove stale'
stale_days = 30
`

func cmdInit() {
	path := configPath()

	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		fmt.Println()
		fmt.Printf("  %s %s\n", dim.Render("config exists:"), muted.Render(path))
		fmt.Println()
		for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			fmt.Printf("    %s\n", dim.Render(line))
		}
		fmt.Println()
		return
	}

	if err := os.WriteFile(path, []byte(defaultConfigTOML), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n", okTag.Render("OK"), muted.Render(path))
	fmt.Println()
	for _, line := range strings.Split(strings.TrimRight(defaultConfigTOML, "\n"), "\n") {
		fmt.Printf("    %s\n", dim.Render(line))
	}
	fmt.Println()
}
