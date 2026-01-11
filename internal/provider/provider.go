package provider

import (
	"context"

	"txdemo/internal/scenario"
)

// Provider defines the interface for database providers
type Provider interface {
	// Name returns the name of the database (e.g., "MongoDB", "PostgreSQL")
	Name() string

	// Description returns a description of the provider
	Description() string

	// Start initializes and starts the database container
	Start(ctx context.Context) error

	// Stop stops and cleans up the database container
	Stop(ctx context.Context) error

	// IsRunning returns whether the database is currently running
	IsRunning() bool

	// GetScenarios returns the registry of scenarios for this provider
	GetScenarios() *scenario.Registry

	// ConnectionInfo returns connection details for display purposes
	ConnectionInfo() string
}

// Registry holds all registered providers
type Registry struct {
	providers []Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make([]Provider, 0),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) {
	r.providers = append(r.providers, p)
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []Provider {
	return r.providers
}

// GetByName returns a provider by name
func (r *Registry) GetByName(name string) Provider {
	for _, p := range r.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
