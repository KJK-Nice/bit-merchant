package domain

import (
	"errors"
	"fmt"
	"time"
)

// Table count bounds for dining QR workflow (owner-configurable).
const (
	MinTableCount = 1
	MaxTableCount = 200
)

// RestaurantID represents a unique restaurant identifier
type RestaurantID string

// Restaurant represents a single restaurant tenant
type Restaurant struct {
	ID             RestaurantID
	Name           string
	TableCount     int
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
		ID:         id,
		Name:       name,
		TableCount: MinTableCount,
		IsOpen:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ValidateRestaurantName validates restaurant name
func ValidateRestaurantName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("restaurant name must be between 1 and 100 characters")
	}
	return nil
}

// ValidateTableCount checks owner-configured table count for QR printing.
func ValidateTableCount(n int) error {
	if n < MinTableCount || n > MaxTableCount {
		return fmt.Errorf("table count must be between %d and %d", MinTableCount, MaxTableCount)
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
