package scenario

import (
	"context"
	"testing"
)

// MockScenario is a mock implementation of the Scenario interface
type MockScenario struct {
	name string
}

func (m *MockScenario) Name() string {
	return m.name
}

func (m *MockScenario) Description() string {
	return "Mock Description"
}

func (m *MockScenario) IsolationLevel() string {
	return "Mock Level"
}

func (m *MockScenario) Setup(ctx context.Context) error {
	return nil
}

func (m *MockScenario) Run(ctx context.Context, output chan<- StepResult) error {
	return nil
}

func (m *MockScenario) Cleanup(ctx context.Context) error {
	return nil
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()

	// Register some scenarios
	r.Register(&MockScenario{name: "Scenario 1"})
	r.Register(&MockScenario{name: "Scenario 2"})

	// Verify count
	if len(r.GetAll()) != 2 {
		t.Fatalf("Expected 2 scenarios, got %d", len(r.GetAll()))
	}

	// Clear registry
	r.Clear()

	// Verify count is 0
	if len(r.GetAll()) != 0 {
		t.Fatalf("Expected 0 scenarios after Clear(), got %d", len(r.GetAll()))
	}

	// Register again to ensure it still works
	r.Register(&MockScenario{name: "Scenario 3"})
	if len(r.GetAll()) != 1 {
		t.Fatalf("Expected 1 scenario, got %d", len(r.GetAll()))
	}
}
