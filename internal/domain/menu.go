package domain

import (
	"errors"
	"time"
)

// CategoryID represents a unique menu category identifier
type CategoryID string

// ItemID represents a unique menu item identifier
type ItemID string

// MenuCategory represents a logical grouping of menu items
type MenuCategory struct {
	ID           CategoryID
	RestaurantID RestaurantID
	Name         string
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewMenuCategory creates a new MenuCategory with validation
func NewMenuCategory(id CategoryID, restaurantID RestaurantID, name string, displayOrder int) (*MenuCategory, error) {
	if err := ValidateCategoryName(name); err != nil {
		return nil, err
	}
	if displayOrder < 0 {
		return nil, errors.New("display order must be >= 0")
	}

	now := time.Now()
	return &MenuCategory{
		ID:           id,
		RestaurantID: restaurantID,
		Name:         name,
		DisplayOrder: displayOrder,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidateCategoryName validates category name
func ValidateCategoryName(name string) error {
	if len(name) == 0 || len(name) > 50 {
		return errors.New("category name must be between 1 and 50 characters")
	}
	return nil
}

// SetActive updates category active status
func (c *MenuCategory) SetActive(isActive bool) {
	c.IsActive = isActive
	c.UpdatedAt = time.Now()
}

// MenuItem represents a food/drink item
type MenuItem struct {
	ID               ItemID
	CategoryID       CategoryID
	RestaurantID     RestaurantID
	Name             string
	Description      string
	Price            float64
	PhotoURL         string
	PhotoOriginalURL string
	IsAvailable      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewMenuItem creates a new MenuItem with validation
func NewMenuItem(id ItemID, categoryID CategoryID, restaurantID RestaurantID, name string, price float64) (*MenuItem, error) {
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
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidateItemName validates item name
func ValidateItemName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("item name must be between 1 and 100 characters")
	}
	return nil
}

// ValidatePrice validates item price
func ValidatePrice(price float64) error {
	if price <= 0 {
		return errors.New("price must be greater than 0")
	}
	if price > 100_000_000 {
		return errors.New("price must be less than 100,000,000")
	}
	return nil
}

// ValidateDescription validates item description
func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description must be <= 500 characters")
	}
	return nil
}

// SetDescription updates item description
func (m *MenuItem) SetDescription(description string) error {
	if err := ValidateDescription(description); err != nil {
		return err
	}
	m.Description = description
	m.UpdatedAt = time.Now()
	return nil
}

// SetPhotoURLs updates photo URLs
func (m *MenuItem) SetPhotoURLs(photoURL, photoOriginalURL string) {
	m.PhotoURL = photoURL
	m.PhotoOriginalURL = photoOriginalURL
	m.UpdatedAt = time.Now()
}

// SetAvailable updates item availability
func (m *MenuItem) SetAvailable(isAvailable bool) {
	m.IsAvailable = isAvailable
	m.UpdatedAt = time.Now()
}
