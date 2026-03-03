// Package ui provides terminal output styling and simple interactive prompts.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Adaptive color pairs: (light terminal bg, dark terminal bg).
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5C4EE5", Dark: "#9D8EFF"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#0E9F6E", Dark: "#31C48D"}
	colorWarning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
	colorDanger  = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}

	// Exported styles for use in command output formatting.
	Primary = lipgloss.NewStyle().Foreground(colorPrimary)
	Success = lipgloss.NewStyle().Foreground(colorSuccess)
	Warning = lipgloss.NewStyle().Foreground(colorWarning)
	Danger  = lipgloss.NewStyle().Foreground(colorDanger)
	Muted   = lipgloss.NewStyle().Foreground(colorMuted)
	Bold    = lipgloss.NewStyle().Bold(true)

	// ActiveMarker is the bullet shown next to the current profile.
	ActiveMarker = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("●")
)

// Successf prints a styled success message to stdout.
func Successf(format string, a ...any) {
	fmt.Println(Success.Render("✓ " + fmt.Sprintf(format, a...)))
}

// Infof prints a styled info message to stdout.
func Infof(format string, a ...any) {
	fmt.Println(Primary.Render("→ " + fmt.Sprintf(format, a...)))
}

// Warningf prints a styled warning message to stdout.
func Warningf(format string, a ...any) {
	fmt.Println(Warning.Render("⚠ " + fmt.Sprintf(format, a...)))
}

// Errorf prints a styled error message to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, Danger.Render("✗ "+fmt.Sprintf(format, a...)))
}

// Input prompts for text input. Returns defaultVal unchanged if the user presses Enter.
func Input(prompt, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("%s %s %s: ", Primary.Render("?"), Bold.Render(prompt), Muted.Render("("+defaultVal+")"))
	} else {
		fmt.Printf("%s %s: ", Primary.Render("?"), Bold.Render(prompt))
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

// Confirm asks a yes/no question. Returns true only for "y" or "yes".
func Confirm(prompt string) (bool, error) {
	fmt.Printf("%s %s %s: ", Primary.Render("?"), Bold.Render(prompt), Muted.Render("[y/N]"))
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes", nil
}

// Select shows a numbered list and returns the zero-based index of the chosen item.
func Select(title string, options []string) (int, error) {
	fmt.Printf("\n%s %s\n\n", Primary.Render("?"), Bold.Render(title))
	for i, opt := range options {
		fmt.Printf("  %s  %s\n", Muted.Render(fmt.Sprintf("[%d]", i+1)), opt)
	}
	fmt.Printf("\n%s", Bold.Render("Enter number: "))

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}
	line = strings.TrimSpace(line)

	var choice int
	if _, err := fmt.Sscanf(line, "%d", &choice); err != nil || choice < 1 || choice > len(options) {
		return -1, fmt.Errorf("invalid selection %q", line)
	}
	return choice - 1, nil
}

// IsTTY reports whether stdin is connected to a terminal.
// Always returns false during `go test` runs so that commands never block
// waiting for interactive input in the test environment.
func IsTTY() bool {
	if testing.Testing() {
		return false
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
