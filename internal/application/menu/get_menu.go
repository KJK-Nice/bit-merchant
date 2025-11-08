package menu

import (
	"errors"

	"bitmerchant/internal/domain"
)

// GetMenuUseCase retrieves menu with categories and items for a restaurant
type GetMenuUseCase struct {
	restaurantRepo domain.RestaurantRepository
	categoryRepo   domain.MenuCategoryRepository
	itemRepo       domain.MenuItemRepository
}

// NewGetMenuUseCase creates a new GetMenuUseCase
func NewGetMenuUseCase(
	restaurantRepo domain.RestaurantRepository,
	categoryRepo domain.MenuCategoryRepository,
	itemRepo domain.MenuItemRepository,
) *GetMenuUseCase {
	return &GetMenuUseCase{
		restaurantRepo: restaurantRepo,
		categoryRepo:   categoryRepo,
		itemRepo:       itemRepo,
	}
}

// MenuResult represents the menu data structure
type MenuResult struct {
	Restaurant *domain.Restaurant
	Categories []CategoryWithItems
}

// CategoryWithItems represents a category with its items
type CategoryWithItems struct {
	Category *domain.MenuCategory
	Items    []*domain.MenuItem
}

// Execute retrieves menu for restaurant
func (uc *GetMenuUseCase) Execute(restaurantID domain.RestaurantID) (*MenuResult, error) {
	restaurant, err := uc.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return nil, errors.New("restaurant not found")
	}

	if !restaurant.IsOpen {
		return nil, errors.New("restaurant is closed")
	}

	categories, err := uc.categoryRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	allItems, err := uc.itemRepo.FindAvailableByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	// Group items by category
	categoryMap := make(map[domain.CategoryID][]*domain.MenuItem)
	for _, item := range allItems {
		categoryMap[item.CategoryID] = append(categoryMap[item.CategoryID], item)
	}

	// Build result with categories and items
	categoriesWithItems := make([]CategoryWithItems, 0, len(categories))
	for _, cat := range categories {
		if cat.IsActive {
			categoriesWithItems = append(categoriesWithItems, CategoryWithItems{
				Category: cat,
				Items:    categoryMap[cat.ID],
			})
		}
	}

	return &MenuResult{
		Restaurant: restaurant,
		Categories: categoriesWithItems,
	}, nil
}
