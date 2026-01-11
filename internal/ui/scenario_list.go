package ui

import (
	"fmt"
	"strings"

	"txdemo/internal/provider"
	"txdemo/internal/scenario"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScenarioListModel represents the scenario selection view
type ScenarioListModel struct {
	provider  provider.Provider
	scenarios []scenario.Scenario
	cursor    int
}

// NewScenarioListModel creates a new scenario list model
func NewScenarioListModel(p provider.Provider) *ScenarioListModel {
	return &ScenarioListModel{
		provider:  p,
		scenarios: p.GetScenarios().GetAll(),
		cursor:    0,
	}
}

// Update handles scenario list input
func (m *ScenarioListModel) Update(msg tea.Msg) (*ScenarioListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.scenarios)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// Selected returns the currently selected scenario
func (m *ScenarioListModel) Selected() scenario.Scenario {
	if m.cursor >= 0 && m.cursor < len(m.scenarios) {
		return m.scenarios[m.cursor]
	}
	return nil
}

// View renders the scenario list
func (m *ScenarioListModel) View() string {
	var b strings.Builder

	// Header
	providerBadge := Badge(m.provider.Name(), lipgloss.Color("#10B981"))

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("📚 Select Demonstration Scenario")

	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("  ")
	b.WriteString(providerBadge)
	b.WriteString("\n\n")

	// Connection info
	connInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		Render(fmt.Sprintf("Connected: %s", m.provider.ConnectionInfo()))
	b.WriteString(connInfo)
	b.WriteString("\n\n")

	if len(m.scenarios) == 0 {
		b.WriteString(WarningStyle.Render("  No scenarios available"))
		return b.String()
	}

	// Scenario items
	for i, s := range m.scenarios {
		cursor := "  "
		nameStyle := NormalStyle

		if i == m.cursor {
			cursor = "▸ "
			nameStyle = SelectedStyle
		}

		// Isolation level badge
		levelBadge := Badge(s.IsolationLevel(), lipgloss.Color("#7C3AED"))

		b.WriteString(fmt.Sprintf("%s%s  %s\n",
			CursorStyle.Render(cursor),
			nameStyle.Render(s.Name()),
			levelBadge))

		// Show description for selected item
		if i == m.cursor {
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				MarginLeft(4).
				Width(70)

			// First few lines of description
			desc := s.Description()
			lines := strings.Split(desc, "\n")
			if len(lines) > 3 {
				lines = lines[:3]
			}
			b.WriteString(descStyle.Render(strings.Join(lines, "\n")))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString(HelpStyle.Render("↑/↓ navigate • enter run scenario • esc/q back"))

	return b.String()
}
