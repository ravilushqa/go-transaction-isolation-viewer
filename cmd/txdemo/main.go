package main

import (
	"fmt"
	"os"

	"txdemo/internal/provider"
	"txdemo/internal/provider/mongodb"
	"txdemo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create provider registry
	providers := provider.NewRegistry()

	// Register MongoDB provider
	providers.Register(mongodb.NewProvider())

	// Create the application
	app := ui.NewApp(providers)

	// Run the TUI
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
