package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"txdemo/internal/scenario"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunnerModel displays the scenario execution
type RunnerModel struct {
	scenario scenario.Scenario
	results  []scenario.StepResult
	running  bool
	done     bool
	err      error
	frame    int
}

// NewRunnerModel creates a new runner model
func NewRunnerModel(s scenario.Scenario) *RunnerModel {
	return &RunnerModel{
		scenario: s,
		results:  make([]scenario.StepResult, 0),
		running:  false,
	}
}

// Start begins the scenario execution
func (r *RunnerModel) Start() tea.Cmd {
	return func() tea.Msg {
		return runnerStartMsg{}
	}
}

type runnerStartMsg struct{}
type runnerStepMsg struct {
	result scenario.StepResult
}
type runnerCompleteMsg struct {
	err error
}
type runnerTickMsg struct{}

// Update handles runner updates
func (r *RunnerModel) Update(msg tea.Msg) (*RunnerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case runnerStartMsg:
		r.running = true
		r.results = nil
		return r, tea.Batch(r.runScenario(), r.tick())

	case runnerStepMsg:
		r.results = append(r.results, msg.result)
		return r, nil

	case runnerCompleteMsg:
		r.running = false
		r.done = true
		r.err = msg.err
		return r, func() tea.Msg { return RunnerDoneMsg{} }

	case runnerTickMsg:
		r.frame++
		if r.running {
			return r, r.tick()
		}
		return r, nil
	}

	return r, nil
}

func (r *RunnerModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return runnerTickMsg{}
	})
}

func (r *RunnerModel) runScenario() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		output := make(chan scenario.StepResult, 100)

		// Setup
		if err := r.scenario.Setup(ctx); err != nil {
			return runnerCompleteMsg{err: err}
		}

		// Run in goroutine
		var runErr error
		go func() {
			runErr = r.scenario.Run(ctx, output)
		}()

		// Collect results
		for result := range output {
			// Send each result as a message
			// Note: This is a simplified approach; in a real app we'd need
			// a proper channel-based message system
			r.results = append(r.results, result)
		}

		// Cleanup
		_ = r.scenario.Cleanup(ctx)

		return runnerCompleteMsg{err: runErr}
	}
}

// View renders the runner
func (r *RunnerModel) View() string {
	var b strings.Builder

	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Render(fmt.Sprintf("🎬 %s", r.scenario.Name()))

	b.WriteString("\n")
	b.WriteString(title)

	// Status indicator
	if r.running {
		spinner := SpinnerFrames[r.frame%len(SpinnerFrames)]
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Render(fmt.Sprintf("  %s Running...", spinner))
		b.WriteString(status)
	} else if r.done {
		if r.err != nil {
			status := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Render("  ❌ Error")
			b.WriteString(status)
		} else {
			status := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Render("  ✓ Complete")
			b.WriteString(status)
		}
	}

	b.WriteString("\n")

	// Isolation level badge
	levelBadge := Badge(r.scenario.IsolationLevel(), lipgloss.Color("#7C3AED"))
	b.WriteString(levelBadge)
	b.WriteString("\n\n")

	// Results
	if len(r.results) == 0 && r.running {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Render("  Preparing scenario..."))
		b.WriteString("\n")
	}

	for _, result := range r.results {
		if result.IsHeader {
			// Section header
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F9FAFB")).
				Background(lipgloss.Color("#374151")).
				Padding(0, 1).
				MarginTop(1).
				MarginBottom(1)
			b.WriteString(headerStyle.Render(result.Description))
			b.WriteString("\n\n")
			continue
		}

		// Step
		sessionStyle := SessionStyle(result.Session)
		stepNum := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render(fmt.Sprintf("[%d]", result.Step))

		b.WriteString(fmt.Sprintf("%s %s  %s\n",
			stepNum,
			sessionStyle.Render(fmt.Sprintf("%-10s", result.Session)),
			DescriptionStyle.Render(result.Description)))

		// Query
		if result.Query != "" {
			queryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				MarginLeft(4).
				Italic(true)
			b.WriteString(queryStyle.Render("→ " + result.Query))
			b.WriteString("\n")
		}

		// Result
		if result.Result != "" {
			resultStyle := lipgloss.NewStyle().
				MarginLeft(4)

			if result.Success {
				resultStyle = resultStyle.Foreground(lipgloss.Color("#10B981"))
			} else {
				resultStyle = resultStyle.Foreground(lipgloss.Color("#EF4444"))
			}

			// Handle multiline results
			lines := strings.Split(result.Result, "\n")
			for _, line := range lines {
				b.WriteString(resultStyle.Render("  " + line))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
	}

	// Error message
	if r.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("\nError: %v", r.err)))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	if r.done {
		b.WriteString(HelpStyle.Render("esc/q back to scenarios"))
	} else {
		b.WriteString(HelpStyle.Render("Please wait for scenario to complete..."))
	}

	return b.String()
}
