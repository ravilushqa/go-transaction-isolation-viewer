package main

import (
	"fmt"
	"os"

	"github.com/ravilushqa/go-transaction-isolation-viewer/internal/provider"
	"github.com/ravilushqa/go-transaction-isolation-viewer/internal/provider/mongodb"
	"github.com/ravilushqa/go-transaction-isolation-viewer/internal/ui"

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
