package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuModel represents the main menu
type MenuModel struct {
	items    []string
	cursor   int
	selected int
}

// NewMenuModel creates a new menu model
func NewMenuModel() *MenuModel {
	return &MenuModel{
		items: []string{
			"🗄️  Select Database Provider",
			"❓ Help & About",
			"🚪 Quit",
		},
		cursor: 0,
	}
}

// Update handles menu input
func (m *MenuModel) Update(msg tea.Msg) (*MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// Selected returns the currently selected index
func (m *MenuModel) Selected() int {
	return m.cursor
}

// View renders the menu
func (m *MenuModel) View() string {
	var b strings.Builder

	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("🔄 Transaction Isolation Levels Demo")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		MarginBottom(2).
		Render("Learn how database isolation levels work with live demonstrations")

	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		style := NormalStyle

		if i == m.cursor {
			cursor = "▸ "
			style = SelectedStyle
		}

		b.WriteString(fmt.Sprintf("%s%s\n", CursorStyle.Render(cursor), style.Render(item)))
	}

	// Help
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ navigate • enter select • q quit"))

	return b.String()
}
