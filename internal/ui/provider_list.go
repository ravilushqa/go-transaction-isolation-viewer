package ui

import (
	"fmt"
	"strings"

	"txdemo/internal/provider"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProviderListModel represents the provider selection view
type ProviderListModel struct {
	providers    *provider.Registry
	cursor       int
	loading      bool
	loadingFrame int
}

// NewProviderListModel creates a new provider list model
func NewProviderListModel(providers *provider.Registry) *ProviderListModel {
	return &ProviderListModel{
		providers: providers,
		cursor:    0,
	}
}

// Update handles provider list input
func (m *ProviderListModel) Update(msg tea.Msg) (*ProviderListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			providers := m.providers.GetAll()
			if m.cursor < len(providers)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// Selected returns the currently selected provider
func (m *ProviderListModel) Selected() provider.Provider {
	providers := m.providers.GetAll()
	if m.cursor >= 0 && m.cursor < len(providers) {
		return providers[m.cursor]
	}
	return nil
}

// View renders the provider list
func (m *ProviderListModel) View() string {
	var b strings.Builder

	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("🗄️ Select Database Provider")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		MarginBottom(2).
		Render("Choose a database to explore its isolation levels")

	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	providers := m.providers.GetAll()

	if len(providers) == 0 {
		b.WriteString(WarningStyle.Render("  No providers registered"))
		return b.String()
	}

	// Provider items
	for i, p := range providers {
		cursor := "  "
		nameStyle := NormalStyle
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).MarginLeft(4)

		if i == m.cursor {
			cursor = "▸ "
			nameStyle = SelectedStyle
		}

		// Provider icon based on name
		icon := "📦"
		switch p.Name() {
		case "MongoDB":
			icon = "🍃"
		case "PostgreSQL":
			icon = "🐘"
		case "MySQL":
			icon = "🐬"
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n",
			CursorStyle.Render(cursor),
			icon,
			nameStyle.Render(p.Name())))
		b.WriteString(descStyle.Render(p.Description()))
		b.WriteString("\n\n")
	}

	// Note about container
	note := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Italic(true).
		Render("⚠️  This will start a Docker container using testcontainers")

	b.WriteString(note)
	b.WriteString("\n\n")

	// Help
	b.WriteString(HelpStyle.Render("↑/↓ navigate • enter select • esc/q back"))

	return b.String()
}
