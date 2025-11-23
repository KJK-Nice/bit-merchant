package domain

import (
	"errors"
	"time"
)

// RestaurantID represents a unique restaurant identifier
type RestaurantID string

// Restaurant represents a single restaurant tenant
type Restaurant struct {
	ID             RestaurantID
	Name           string
	IsOpen         bool
	ClosedMessage  string
	ReopeningHours string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewRestaurant creates a new Restaurant with validation
func NewRestaurant(id RestaurantID, name string) (*Restaurant, error) {
	if err := ValidateRestaurantName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Restaurant{
		ID:        id,
		Name:      name,
		IsOpen:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ValidateRestaurantName validates restaurant name
func ValidateRestaurantName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("restaurant name must be between 1 and 100 characters")
	}
	return nil
}

// UpdateStatus updates restaurant open/closed status
func (r *Restaurant) UpdateStatus(isOpen bool, closedMessage, reopeningHours string) {
	r.IsOpen = isOpen
	r.ClosedMessage = closedMessage
	r.ReopeningHours = reopeningHours
	r.UpdatedAt = time.Now()
}
