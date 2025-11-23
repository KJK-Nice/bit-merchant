package menu_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestGetMenuUseCase(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	uc := menu.NewGetMenuUseCase(catRepo, itemRepo)

	restID := domain.RestaurantID("r1")

	// Setup data
	cat1, _ := domain.NewMenuCategory("c1", restID, "Starters", 1)
	cat2, _ := domain.NewMenuCategory("c2", restID, "Mains", 2)
	catRepo.Save(cat1)
	catRepo.Save(cat2)

	item1, _ := domain.NewMenuItem("i1", "c1", restID, "Salad", 10.0)
	item2, _ := domain.NewMenuItem("i2", "c2", restID, "Steak", 20.0)
	itemRepo.Save(item1)
	itemRepo.Save(item2)

	t.Run("Execute", func(t *testing.T) {
		result, err := uc.Execute(context.Background(), restID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Categories, 2)

		// Check structure
		// Assuming result structure: struct { Categories []CategoryWithItems }
		// CategoryWithItems has Items field
		
		// For now, assuming pure domain entities returned or DTOs. 
		// Spec implies displaying menu organized by categories.
		// Let's assume DTO structure for now or check implementation plan.
		// Implementation plan says: "returns restaurant menu with categories and items"
	})
}

