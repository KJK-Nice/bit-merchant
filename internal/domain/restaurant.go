package domain

import (
	"errors"
	"time"
)

// RestaurantID represents a unique restaurant identifier
type RestaurantID string

// Restaurant represents a single restaurant tenant
type Restaurant struct {
	ID               RestaurantID
	Name             string
	LightningAddress string
	IsOpen           bool
	ClosedMessage    string
	ReopeningHours   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewRestaurant creates a new Restaurant with validation
func NewRestaurant(id RestaurantID, name, lightningAddress string) (*Restaurant, error) {
	if err := ValidateRestaurantName(name); err != nil {
		return nil, err
	}
	if err := ValidateLightningAddress(lightningAddress); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Restaurant{
		ID:               id,
		Name:             name,
		LightningAddress: lightningAddress,
		IsOpen:           true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// ValidateRestaurantName validates restaurant name
func ValidateRestaurantName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("restaurant name must be between 1 and 100 characters")
	}
	return nil
}

// ValidateLightningAddress validates Lightning address format
func ValidateLightningAddress(address string) error {
	if len(address) == 0 {
		return errors.New("lightning address is required")
	}
	// Basic validation - should contain @ symbol
	// Full validation would check against Lightning address spec
	if len(address) < 3 {
		return errors.New("invalid lightning address format")
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
