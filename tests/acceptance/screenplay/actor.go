package screenplay

import (
	"fmt"
	"testing"
)

// Actor represents a user or external system interacting with the application.
type Actor struct {
	Name      string
	Abilities map[string]Ability
	T         *testing.T // Optional, for assertions if needed directly in tasks
}

// NewActor creates a new actor with a name.
func NewActor(name string, t *testing.T) *Actor {
	return &Actor{
		Name:      name,
		Abilities: make(map[string]Ability),
		T:         t,
	}
}

// WhoCan gives the actor an ability.
func (a *Actor) WhoCan(ability Ability) *Actor {
	a.Abilities[ability.Name()] = ability
	return a
}

// Ability returns an ability by name, or nil if not possessed.
func (a *Actor) Ability(name string) Ability {
	return a.Abilities[name]
}

// AttemptsTo performs a sequence of tasks or interactions.
func (a *Actor) AttemptsTo(tasks ...Task) error {
	for _, task := range tasks {
		if err := task.PerformAs(a); err != nil {
			return fmt.Errorf("%s failed to %s: %w", a.Name, task.Description(), err)
		}
	}
	return nil
}

// Task represents a high-level action an actor can perform.
type Task interface {
	PerformAs(actor *Actor) error
	Description() string
}

// Ability represents a capability an actor possesses (e.g., BrowseTheWeb).
type Ability interface {
	Name() string
}

