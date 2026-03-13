package main

import (
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	white     = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e4e9")).Bold(true)
	blue      = lipgloss.NewStyle().Foreground(lipgloss.Color("#61afef"))
	green     = lipgloss.NewStyle().Foreground(lipgloss.Color("#98c379"))
	yellow    = lipgloss.NewStyle().Foreground(lipgloss.Color("#e5c07b"))
	red       = lipgloss.NewStyle().Foreground(lipgloss.Color("#e06c75"))
	dim       = lipgloss.NewStyle().Foreground(lipgloss.Color("#5c6370"))
	dimmer    = lipgloss.NewStyle().Foreground(lipgloss.Color("#3e4451"))
	muted     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7a8394"))
	okTag     = lipgloss.NewStyle().Foreground(lipgloss.Color("#282c34")).Background(lipgloss.Color("#98c379")).Bold(true).Padding(0, 1)
	skipBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#282c34")).Background(lipgloss.Color("#e5c07b")).Bold(true).Padding(0, 1)
	failBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#282c34")).Background(lipgloss.Color("#e06c75")).Bold(true).Padding(0, 1)
	divider   = dimmer.Render(strings.Repeat("─", 50))
)

func customTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Base = t.Focused.Base.
		BorderForeground(lipgloss.Color("#3e4451"))
	t.Focused.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e2e4e9")).Bold(true).
		MarginBottom(1)
	t.Focused.Description = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5c6370")).
		MarginBottom(1)

	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		SetString("› ").
		Foreground(lipgloss.Color("#61afef")).Bold(true)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		SetString("◉ ").
		Foreground(lipgloss.Color("#98c379"))
	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98c379"))
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		SetString("○ ").
		Foreground(lipgloss.Color("#3e4451"))
	t.Focused.UnselectedOption = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7a8394"))

	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#61afef"))
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#61afef"))
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3e4451"))
	t.Focused.TextInput.Text = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e2e4e9"))

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")

	return t
}

func pad(s string, width int) string {
	n := width - len(s)
	if n < 2 {
		n = 2
	}
	return s + strings.Repeat(" ", n)
}
