package mongodb

import (
	"context"
	"fmt"

	"txdemo/internal/provider"
	"txdemo/internal/scenario"

	mongoScenarios "txdemo/internal/scenario/mongodb"
)

// Compile-time interface check
var _ provider.Provider = (*Provider)(nil)

// Provider implements the provider.Provider interface for MongoDB
type Provider struct {
	container *Container
	scenarios *scenario.Registry
}

// NewProvider creates a new MongoDB provider
func NewProvider() *Provider {
	p := &Provider{
		container: NewContainer(),
		scenarios: scenario.NewRegistry(),
	}
	return p
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "MongoDB"
}

// Description returns the provider description
func (p *Provider) Description() string {
	return "MongoDB 7.0 with replica set for multi-document transaction support"
}

// Start initializes the MongoDB container and registers scenarios
func (p *Provider) Start(ctx context.Context) error {
	if err := p.container.Start(ctx); err != nil {
		return err
	}

	// Register MongoDB-specific scenarios
	p.scenarios.Clear()
	p.registerScenarios()

	return nil
}

// Stop terminates the MongoDB container
func (p *Provider) Stop(ctx context.Context) error {
	return p.container.Stop(ctx)
}

// IsRunning returns whether the container is running
func (p *Provider) IsRunning() bool {
	return p.container.IsRunning()
}

// GetScenarios returns the scenario registry
func (p *Provider) GetScenarios() *scenario.Registry {
	return p.scenarios
}

// ConnectionInfo returns connection details
func (p *Provider) ConnectionInfo() string {
	connStr := p.container.ConnectionString()
	if connStr == "" {
		return "Not connected"
	}
	return fmt.Sprintf("Connected to MongoDB replica set\n%s", connStr)
}

// GetContainer returns the underlying container for scenario access
func (p *Provider) GetContainer() *Container {
	return p.container
}

// registerScenarios registers all MongoDB-specific scenarios
func (p *Provider) registerScenarios() {
	db := p.container.Database("txdemo")
	client := p.container.Client()

	// Register scenarios
	p.scenarios.Register(mongoScenarios.NewDirtyReadScenario(client, db))
	p.scenarios.Register(mongoScenarios.NewReadCommittedScenario(client, db))
	p.scenarios.Register(mongoScenarios.NewSnapshotIsolationScenario(client, db))
	p.scenarios.Register(mongoScenarios.NewWriteConflictScenario(client, db))
}
