package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingModel displays a loading spinner with status messages
type LoadingModel struct {
	title    string
	messages []string
	frame    int
	done     bool
}

// NewLoadingModel creates a new loading model
func NewLoadingModel(title string) *LoadingModel {
	return &LoadingModel{
		title:    title,
		messages: []string{},
		frame:    0,
	}
}

// AddMessage adds a status message
func (l *LoadingModel) AddMessage(msg string) {
	l.messages = append(l.messages, msg)
}

// SetDone marks loading as complete
func (l *LoadingModel) SetDone() {
	l.done = true
}

type loadingTickMsg struct{}

// Tick returns a command that ticks the spinner
func (l *LoadingModel) Tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return loadingTickMsg{}
	})
}

// Update handles loading model updates
func (l *LoadingModel) Update(msg tea.Msg) (*LoadingModel, tea.Cmd) {
	switch msg.(type) {
	case loadingTickMsg:
		l.frame++
		if !l.done {
			return l, l.Tick()
		}
	}
	return l, nil
}

// View renders the loading screen
func (l *LoadingModel) View() string {
	var b strings.Builder

	// Title with spinner
	spinner := SpinnerFrames[l.frame%len(SpinnerFrames)]

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))

	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B"))

	b.WriteString("\n")
	b.WriteString(spinnerStyle.Render(spinner))
	b.WriteString(" ")
	b.WriteString(titleStyle.Render(l.title))
	b.WriteString("\n\n")

	// Status messages
	checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))

	for i, msg := range l.messages {
		if i < len(l.messages)-1 || l.done {
			// Completed step
			b.WriteString(checkStyle.Render("  ✓ "))
		} else {
			// Current step
			b.WriteString(spinnerStyle.Render(fmt.Sprintf("  %s ", SpinnerFrames[l.frame%len(SpinnerFrames)])))
		}
		b.WriteString(msgStyle.Render(msg))
		b.WriteString("\n")
	}

	// Add some padding
	b.WriteString("\n")

	// Tips
	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	tips := []string{
		"💡 MongoDB requires a replica set for multi-document transactions",
		"💡 First container pull may take a minute or two",
		"💡 Subsequent runs will be much faster",
	}

	tipIndex := (l.frame / 30) % len(tips)
	b.WriteString(tipStyle.Render(tips[tipIndex]))
	b.WriteString("\n")

	return b.String()
}

// Predefined loading messages for container startup
var ContainerStartupMessages = []string{
	"Pulling MongoDB 7.0 image...",
	"Starting container...",
	"Initializing replica set...",
	"Waiting for MongoDB to be ready...",
	"Connecting to database...",
}
