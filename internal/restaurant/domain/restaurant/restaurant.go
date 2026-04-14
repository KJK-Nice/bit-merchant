package restaurant

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

const (
	MinTableCount = 1
	MaxTableCount = 200
)

var ErrInvalidTableCount = errors.New("invalid table count")

// Restaurant represents a single restaurant tenant.
type Restaurant struct {
	ID             common.RestaurantID
	Name           string
	TableCount     int
	IsOpen         bool
	ClosedMessage  string
	ReopeningHours string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewRestaurant creates a new Restaurant with validation.
func NewRestaurant(id common.RestaurantID, name string) (*Restaurant, error) {
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

func ValidateRestaurantName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("restaurant name must be between 1 and 100 characters")
	}
	return nil
}

func ValidateTableCount(n int) error {
	if n < MinTableCount || n > MaxTableCount {
		return ErrInvalidTableCount
	}
	return nil
}

func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description must be <= 500 characters")
	}
	return nil
}

// Open marks the restaurant as open, clearing closed messaging.
func (r *Restaurant) Open() {
	r.IsOpen = true
	r.ClosedMessage = ""
	r.ReopeningHours = ""
	r.UpdatedAt = time.Now()
}

// Close marks the restaurant as closed with optional messaging.
func (r *Restaurant) Close(closedMessage, reopeningHours string) {
	r.IsOpen = false
	r.ClosedMessage = closedMessage
	r.ReopeningHours = reopeningHours
	r.UpdatedAt = time.Now()
}

// UpdateStatus updates restaurant open/closed status (kept for backward compatibility).
func (r *Restaurant) UpdateStatus(isOpen bool, closedMessage, reopeningHours string) {
	r.IsOpen = isOpen
	r.ClosedMessage = closedMessage
	r.ReopeningHours = reopeningHours
	r.UpdatedAt = time.Now()
}
