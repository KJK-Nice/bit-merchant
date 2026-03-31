package menu

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// MenuCategory represents a logical grouping of menu items.
type MenuCategory struct {
	ID           common.CategoryID
	RestaurantID common.RestaurantID
	Name         string
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewMenuCategory(id common.CategoryID, restaurantID common.RestaurantID, name string, displayOrder int) (*MenuCategory, error) {
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

func ValidateCategoryName(name string) error {
	if len(name) == 0 || len(name) > 50 {
		return errors.New("category name must be between 1 and 50 characters")
	}
	return nil
}

// Activate marks the category as active.
func (c *MenuCategory) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// Deactivate marks the category as inactive.
func (c *MenuCategory) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// SetActive updates category active status.
func (c *MenuCategory) SetActive(isActive bool) {
	c.IsActive = isActive
	c.UpdatedAt = time.Now()
}
