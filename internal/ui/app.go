package ui

import (
	"context"
	"fmt"

	"txdemo/internal/provider"
	"txdemo/internal/scenario"

	tea "github.com/charmbracelet/bubbletea"
)

// View represents the current view in the application
type View int

const (
	ViewMenu View = iota
	ViewProviderSelect
	ViewLoading
	ViewScenarioList
	ViewRunner
	ViewHelp
)

// App is the main application model
type App struct {
	providers    *provider.Registry
	currentView  View
	menu         *MenuModel
	providerList *ProviderListModel
	loading      *LoadingModel
	scenarioList *ScenarioListModel
	runner       *RunnerModel
	help         *HelpModel

	selectedProvider provider.Provider
	width            int
	height           int
	err              error
	quitting         bool
}

// NewApp creates a new application
func NewApp(providers *provider.Registry) *App {
	app := &App{
		providers:   providers,
		currentView: ViewMenu,
		width:       80,
		height:      24,
	}

	app.menu = NewMenuModel()
	app.help = NewHelpModel()
	app.providerList = NewProviderListModel(providers)

	return app
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			return a, a.cleanup()
		case "q":
			if a.currentView == ViewMenu {
				a.quitting = true
				return a, a.cleanup()
			}
			// Go back
			return a, a.goBack()
		case "esc":
			return a, a.goBack()
		}

	case ProviderStartedMsg:
		if msg.Err != nil {
			a.err = msg.Err
			a.currentView = ViewProviderSelect
			return a, nil
		}
		a.selectedProvider = msg.Provider
		a.scenarioList = NewScenarioListModel(msg.Provider)
		a.currentView = ViewScenarioList
		return a, nil

	case loadingTickMsg:
		if a.loading != nil {
			var cmd tea.Cmd
			a.loading, cmd = a.loading.Update(msg)
			return a, cmd
		}
		return a, nil

	case ProviderStoppedMsg:
		a.selectedProvider = nil
		if a.quitting {
			return a, tea.Quit
		}
		return a, nil

	case ScenarioSelectedMsg:
		a.runner = NewRunnerModel(msg.Scenario)
		a.currentView = ViewRunner
		return a, a.runner.Start()

	case RunnerDoneMsg:
		// Stay on runner view to show results
		return a, nil
	}

	// Delegate to current view
	var cmd tea.Cmd
	switch a.currentView {
	case ViewMenu:
		cmd = a.updateMenu(msg)
	case ViewProviderSelect:
		cmd = a.updateProviderList(msg)
	case ViewLoading:
		// Loading view handles its own updates via loadingTickMsg
	case ViewScenarioList:
		cmd = a.updateScenarioList(msg)
	case ViewRunner:
		cmd = a.updateRunner(msg)
	case ViewHelp:
		cmd = a.updateHelp(msg)
	}

	return a, cmd
}

func (a *App) updateMenu(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch a.menu.Selected() {
			case 0: // Select Database
				a.currentView = ViewProviderSelect
			case 1: // Help
				a.currentView = ViewHelp
			case 2: // Quit
				a.quitting = true
				return a.cleanup()
			}
		}
	}

	var cmd tea.Cmd
	a.menu, cmd = a.menu.Update(msg)
	return cmd
}

func (a *App) updateProviderList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := a.providerList.Selected()
			if selected != nil {
				return a.startProvider(selected)
			}
		}
	}

	var cmd tea.Cmd
	a.providerList, cmd = a.providerList.Update(msg)
	return cmd
}

func (a *App) updateScenarioList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			scenario := a.scenarioList.Selected()
			if scenario != nil {
				return func() tea.Msg {
					return ScenarioSelectedMsg{Scenario: scenario}
				}
			}
		}
	}

	var cmd tea.Cmd
	a.scenarioList, cmd = a.scenarioList.Update(msg)
	return cmd
}

func (a *App) updateRunner(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	a.runner, cmd = a.runner.Update(msg)
	return cmd
}

func (a *App) updateHelp(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	a.help, cmd = a.help.Update(msg)
	return cmd
}

// View implements tea.Model
func (a *App) View() string {
	if a.quitting {
		return "\n  Cleaning up containers...\n\n"
	}

	if a.err != nil {
		return fmt.Sprintf("\n  %s\n\n  Press esc to go back.\n",
			ErrorStyle.Render(fmt.Sprintf("Error: %v", a.err)))
	}

	switch a.currentView {
	case ViewMenu:
		return a.menu.View()
	case ViewProviderSelect:
		return a.providerList.View()
	case ViewLoading:
		if a.loading != nil {
			return a.loading.View()
		}
	case ViewScenarioList:
		return a.scenarioList.View()
	case ViewRunner:
		return a.runner.View()
	case ViewHelp:
		return a.help.View()
	}

	return ""
}

func (a *App) goBack() tea.Cmd {
	// Clear any error when going back
	a.err = nil

	switch a.currentView {
	case ViewProviderSelect:
		a.currentView = ViewMenu
	case ViewLoading:
		// Can't go back while loading, but clear loading state
		a.loading = nil
		a.currentView = ViewProviderSelect
	case ViewScenarioList:
		a.currentView = ViewProviderSelect
		// Stop the provider
		if a.selectedProvider != nil {
			return a.stopProvider()
		}
	case ViewRunner:
		a.currentView = ViewScenarioList
	case ViewHelp:
		a.currentView = ViewMenu
	}
	return nil
}

func (a *App) startProvider(p provider.Provider) tea.Cmd {
	// Create loading view
	a.loading = NewLoadingModel(fmt.Sprintf("Starting %s...", p.Name()))
	a.loading.AddMessage("Initializing container...")
	a.currentView = ViewLoading

	// Return batch command: start ticker and start provider
	return tea.Batch(
		a.loading.Tick(),
		func() tea.Msg {
			ctx := context.Background()
			err := p.Start(ctx)
			return ProviderStartedMsg{Provider: p, Err: err}
		},
	)
}

func (a *App) stopProvider() tea.Cmd {
	p := a.selectedProvider
	return func() tea.Msg {
		if p != nil {
			ctx := context.Background()
			_ = p.Stop(ctx)
		}
		return ProviderStoppedMsg{}
	}
}

func (a *App) cleanup() tea.Cmd {
	p := a.selectedProvider
	return func() tea.Msg {
		if p != nil {
			ctx := context.Background()
			_ = p.Stop(ctx)
		}
		return tea.Quit()
	}
}

// Message types
type ProviderStartedMsg struct {
	Provider provider.Provider
	Err      error
}

type ProviderStoppedMsg struct{}

type ScenarioSelectedMsg struct {
	Scenario scenario.Scenario
}

type RunnerDoneMsg struct{}
