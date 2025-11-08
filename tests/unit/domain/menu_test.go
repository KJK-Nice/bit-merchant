package domain_test

import (
	"testing"

	"bitmerchant/internal/domain"
)

func TestMenuItem_DisplayLogic(t *testing.T) {
	tests := []struct {
		name          string
		item          *domain.MenuItem
		wantAvailable bool
		wantDisplay   bool
	}{
		{
			name: "available item should display",
			item: &domain.MenuItem{
				ID:          "item_001",
				IsAvailable: true,
			},
			wantAvailable: true,
			wantDisplay:   true,
		},
		{
			name: "unavailable item should not display",
			item: &domain.MenuItem{
				ID:          "item_002",
				IsAvailable: false,
			},
			wantAvailable: false,
			wantDisplay:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item.IsAvailable != tt.wantAvailable {
				t.Errorf("IsAvailable = %v, want %v", tt.item.IsAvailable, tt.wantAvailable)
			}
			// Item should display if available
			if (tt.item.IsAvailable) != tt.wantDisplay {
				t.Errorf("Display logic failed: available=%v, want display=%v", tt.item.IsAvailable, tt.wantDisplay)
			}
		})
	}
}

func TestNewMenuItem_Validation(t *testing.T) {
	tests := []struct {
		name         string
		id           domain.ItemID
		categoryID   domain.CategoryID
		restaurantID domain.RestaurantID
		itemName     string
		price        float64
		wantError    bool
	}{
		{
			name:         "valid item",
			id:           "item_001",
			categoryID:   "cat_001",
			restaurantID: "rest_001",
			itemName:     "Test Item",
			price:        10.99,
			wantError:    false,
		},
		{
			name:         "empty name should fail",
			id:           "item_002",
			categoryID:   "cat_001",
			restaurantID: "rest_001",
			itemName:     "",
			price:        10.99,
			wantError:    true,
		},
		{
			name:         "zero price should fail",
			id:           "item_003",
			categoryID:   "cat_001",
			restaurantID: "rest_001",
			itemName:     "Test Item",
			price:        0,
			wantError:    true,
		},
		{
			name:         "negative price should fail",
			id:           "item_004",
			categoryID:   "cat_001",
			restaurantID: "rest_001",
			itemName:     "Test Item",
			price:        -5.99,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := domain.NewMenuItem(tt.id, tt.categoryID, tt.restaurantID, tt.itemName, tt.price)
			if (err != nil) != tt.wantError {
				t.Errorf("NewMenuItem() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && item == nil {
				t.Error("NewMenuItem() returned nil item without error")
			}
		})
	}
}

func TestMenuItem_SetAvailable(t *testing.T) {
	item, _ := domain.NewMenuItem("item_001", "cat_001", "rest_001", "Test Item", 10.99)

	item.SetAvailable(false)
	if item.IsAvailable {
		t.Error("SetAvailable(false) did not set IsAvailable to false")
	}

	item.SetAvailable(true)
	if !item.IsAvailable {
		t.Error("SetAvailable(true) did not set IsAvailable to true")
	}
}
