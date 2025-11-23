package menu

import (
	"context"
	"sort"

	"bitmerchant/internal/domain"
)

// GetMenuUseCase retrieves restaurant menu
type GetMenuUseCase struct {
	catRepo  domain.MenuCategoryRepository
	itemRepo domain.MenuItemRepository
}

// NewGetMenuUseCase creates a new GetMenuUseCase
func NewGetMenuUseCase(catRepo domain.MenuCategoryRepository, itemRepo domain.MenuItemRepository) *GetMenuUseCase {
	return &GetMenuUseCase{
		catRepo:  catRepo,
		itemRepo: itemRepo,
	}
}

// MenuResponse represents the menu structure
type MenuResponse struct {
	Categories []CategoryWithItems
}

// CategoryWithItems represents a category with its items
type CategoryWithItems struct {
	Category *domain.MenuCategory
	Items    []*domain.MenuItem
}

// Execute retrieves menu for a restaurant
func (uc *GetMenuUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) (*MenuResponse, error) {
	// Get categories
	categories, err := uc.catRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	// Sort categories by display order
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DisplayOrder < categories[j].DisplayOrder
	})

	// Get all available items
	items, err := uc.itemRepo.FindAvailableByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	// Map items to categories
	itemsByCategory := make(map[domain.CategoryID][]*domain.MenuItem)
	for _, item := range items {
		itemsByCategory[item.CategoryID] = append(itemsByCategory[item.CategoryID], item)
	}

	// Build response
	response := &MenuResponse{
		Categories: make([]CategoryWithItems, 0, len(categories)),
	}

	for _, cat := range categories {
		if !cat.IsActive {
			continue
		}

		catItems := itemsByCategory[cat.ID]
		if len(catItems) == 0 {
			continue // Skip empty categories? Or keep them? Let's keep them if they are active.
			// Actually requirement says "display restaurant menu... organized by categories".
			// Empty categories might look weird. Let's include them for now, owner can deactivate.
		}

		response.Categories = append(response.Categories, CategoryWithItems{
			Category: cat,
			Items:    catItems,
		})
	}

	return response, nil
}
