package dsl

import (
	"testing"
)

// Action represents either a step or an assertion
type Action interface {
	Execute(t *testing.T, app *TestApplication)
	IsAssertion() bool
}

// StepAction wraps a Step as an Action
type StepAction struct {
	Step Step
}

func (a StepAction) Execute(t *testing.T, app *TestApplication) {
	a.Step.Execute(t, app)
}

func (a StepAction) IsAssertion() bool {
	return false
}

// AssertionAction wraps an Assertion as an Action
type AssertionAction struct {
	Assertion Assertion
}

func (a AssertionAction) Execute(t *testing.T, app *TestApplication) {
	a.Assertion.Verify(t, app)
}

func (a AssertionAction) IsAssertion() bool {
	return true
}

// Scenario represents a test scenario with fluent API
type Scenario struct {
	t       *testing.T
	name    string
	setup   *TestSetup
	actions []Action // Interleaved steps and assertions
	context *TestContext
}

// NewScenario creates a new test scenario
func NewScenario(t *testing.T, name string) *Scenario {
	return &Scenario{
		t:       t,
		name:    name,
		setup:   NewTestSetup(),
		actions: []Action{},
		context: NewTestContext(),
	}
}

// Given sets up initial state
func (s *Scenario) Given(setupFn func(*GivenBuilder)) *Scenario {
	builder := &GivenBuilder{setup: s.setup}
	setupFn(builder)
	return s
}

// When defines actions
func (s *Scenario) When(actionFn func(*WhenBuilder)) *Scenario {
	builder := &WhenBuilder{scenario: s}
	actionFn(builder)
	return s
}

// Then defines assertions
func (s *Scenario) Then(assertFn func(*ThenBuilder)) *Scenario {
	builder := &ThenBuilder{scenario: s}
	assertFn(builder)
	return s
}

// addStep adds a step to the scenario
func (s *Scenario) addStep(step Step) {
	s.actions = append(s.actions, StepAction{Step: step})
}

// addAssertion adds an assertion to the scenario
func (s *Scenario) addAssertion(assertion Assertion) {
	s.actions = append(s.actions, AssertionAction{Assertion: assertion})
}

// Run executes the scenario
func (s *Scenario) Run() {
	s.t.Run(s.name, func(t *testing.T) {
		// Setup
		app := s.setup.Build(t)
		defer app.Cleanup()
		app.context = s.context

		// Execute actions in sequence (steps and assertions interleaved)
		for _, action := range s.actions {
			action.Execute(t, app)
		}
	})
}
