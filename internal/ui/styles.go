package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	textColor      = lipgloss.Color("#F9FAFB") // Light
)

// Session colors for differentiating concurrent operations
var (
	sessionAColor = lipgloss.Color("#3B82F6") // Blue
	sessionBColor = lipgloss.Color("#EC4899") // Pink
	setupColor    = lipgloss.Color("#8B5CF6") // Purple
	resultColor   = lipgloss.Color("#10B981") // Green
)

// Base styles
var (
	// Title style for main headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginBottom(1)

	// Box style for content areas
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Selected item in list
	SelectedStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Bold(true).
			Padding(0, 1)

	// Normal item in list
	NormalStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	// Cursor indicator
	CursorStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Help text at bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Header style for scenario sections
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginTop(1).
			MarginBottom(1)

	// Query/code style
	QueryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Italic(true)

	// Result style
	ResultStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Description text
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(textColor)
)

// SessionStyle returns a style for a specific session
func SessionStyle(session string) lipgloss.Style {
	var color lipgloss.Color
	switch session {
	case "Session A":
		color = sessionAColor
	case "Session B":
		color = sessionBColor
	case "Setup":
		color = setupColor
	case "Result":
		color = resultColor
	default:
		color = mutedColor
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
}

// Badge creates a badge-style element
func Badge(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(color).
		Padding(0, 1).
		Bold(true).
		Render(text)
}

// Spinner frames for loading animation
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
