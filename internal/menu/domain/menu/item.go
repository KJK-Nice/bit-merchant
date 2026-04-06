package menu

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// MenuItem represents a food/drink item.
type MenuItem struct {
	ID               common.ItemID
	CategoryID       common.CategoryID
	RestaurantID     common.RestaurantID
	Name             string
	Description      string
	Price            float64
	PhotoURL         string
	PhotoOriginalURL string
	IsAvailable      bool
	DisplayOrder     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewMenuItem(id common.ItemID, categoryID common.CategoryID, restaurantID common.RestaurantID, name string, price float64) (*MenuItem, error) {
	if err := ValidateItemName(name); err != nil {
		return nil, err
	}
	if err := ValidatePrice(price); err != nil {
		return nil, err
	}

	now := time.Now()
	return &MenuItem{
		ID:           id,
		CategoryID:   categoryID,
		RestaurantID: restaurantID,
		Name:         name,
		Price:        price,
		IsAvailable:  true,
		DisplayOrder: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ValidateDisplayOrder(order int) error {
	if order < 0 {
		return errors.New("display order must be >= 0")
	}
	return nil
}

// SetDisplayOrder updates sort position within the category.
func (m *MenuItem) SetDisplayOrder(order int) error {
	if err := ValidateDisplayOrder(order); err != nil {
		return err
	}
	m.DisplayOrder = order
	m.UpdatedAt = time.Now()
	return nil
}

func ValidateItemName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("item name must be between 1 and 100 characters")
	}
	return nil
}

func ValidatePrice(price float64) error {
	if price <= 0 {
		return errors.New("price must be greater than 0")
	}
	if price > 100_000_000 {
		return errors.New("price must be less than 100,000,000")
	}
	return nil
}

func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description must be <= 500 characters")
	}
	return nil
}

func (m *MenuItem) SetDescription(description string) error {
	if err := ValidateDescription(description); err != nil {
		return err
	}
	m.Description = description
	m.UpdatedAt = time.Now()
	return nil
}

func (m *MenuItem) SetPhotoURLs(photoURL, photoOriginalURL string) {
	m.PhotoURL = photoURL
	m.PhotoOriginalURL = photoOriginalURL
	m.UpdatedAt = time.Now()
}

// MakeAvailable marks item as available.
func (m *MenuItem) MakeAvailable() {
	m.IsAvailable = true
	m.UpdatedAt = time.Now()
}

// MakeUnavailable marks item as unavailable.
func (m *MenuItem) MakeUnavailable() {
	m.IsAvailable = false
	m.UpdatedAt = time.Now()
}

// SetAvailable updates item availability.
func (m *MenuItem) SetAvailable(isAvailable bool) {
	m.IsAvailable = isAvailable
	m.UpdatedAt = time.Now()
}
