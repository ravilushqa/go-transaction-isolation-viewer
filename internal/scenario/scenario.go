package scenario

import (
	"context"
)

// StepResult represents the result of a single step in a scenario
type StepResult struct {
	Session     string // Which session/transaction this step belongs to (e.g., "Session A", "Session B")
	Step        int
	Description string
	Query       string // The operation being performed
	Result      string // The result of the operation
	Success     bool
	IsHeader    bool // Whether this is a section header
}

// Scenario defines the interface for transaction isolation demonstrations
type Scenario interface {
	// Name returns the name of the scenario
	Name() string

	// Description returns a detailed description of what this scenario demonstrates
	Description() string

	// IsolationLevel returns the isolation level being demonstrated
	IsolationLevel() string

	// Setup prepares any necessary data before running the scenario
	Setup(ctx context.Context) error

	// Run executes the scenario and sends step results to the output channel
	Run(ctx context.Context, output chan<- StepResult) error

	// Cleanup removes any data created during the scenario
	Cleanup(ctx context.Context) error
}

// Registry holds all registered scenarios
type Registry struct {
	scenarios []Scenario
}

// NewRegistry creates a new scenario registry
func NewRegistry() *Registry {
	return &Registry{
		scenarios: make([]Scenario, 0),
	}
}

// Clear removes all registered scenarios
func (r *Registry) Clear() {
	r.scenarios = make([]Scenario, 0)
}

// Register adds a scenario to the registry
func (r *Registry) Register(s Scenario) {
	r.scenarios = append(r.scenarios, s)
}

// GetAll returns all registered scenarios
func (r *Registry) GetAll() []Scenario {
	return r.scenarios
}

// GetByName returns a scenario by name
func (r *Registry) GetByName(name string) Scenario {
	for _, s := range r.scenarios {
		if s.Name() == name {
			return s
		}
	}
	return nil
}
