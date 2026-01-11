package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel represents the help and about screen
type HelpModel struct {
	viewportHeight int
	viewportWidth  int
}

// NewHelpModel creates a new help model
func NewHelpModel() *HelpModel {
	return &HelpModel{}
}

// Update handles help input
func (m *HelpModel) Update(msg tea.Msg) (*HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewportWidth = msg.Width
		m.viewportHeight = msg.Height
	}
	// Main app handles navigation back with Esc/q
	return m, nil
}

// View renders the help screen
func (m *HelpModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("❓ Help & About")

	b.WriteString(header + "\n")

	// Content
	content := `
TxDemo is an interactive CLI tool for demonstrating database transaction isolation levels.

It helps developers visualize and understand:
• Dirty Reads
• Non-Repeatable Reads
• Phantom Reads
• Serialization Anomalies

navigation:
• Use ↑/↓ to navigate menus
• Press Enter to select items
• Press Esc to go back
• Press q to quit

Created for educational purposes.
`
	// Simple indentation for content
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			b.WriteString("\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("esc back • q quit"))

	return b.String()
}
