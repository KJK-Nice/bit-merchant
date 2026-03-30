package menu

import (
	"context"
	"sort"

	"bitmerchant/internal/domain"
)

// GetMenuForAdminUseCase loads the full menu for owners: all items (any availability),
// all categories including empty and inactive, for use on the admin management screen.
type GetMenuForAdminUseCase struct {
	catRepo  domain.MenuCategoryRepository
	itemRepo domain.MenuItemRepository
	restRepo domain.RestaurantRepository
}

// NewGetMenuForAdminUseCase builds the admin menu use case.
func NewGetMenuForAdminUseCase(catRepo domain.MenuCategoryRepository, itemRepo domain.MenuItemRepository, restRepo domain.RestaurantRepository) *GetMenuForAdminUseCase {
	return &GetMenuForAdminUseCase{
		catRepo:  catRepo,
		itemRepo: itemRepo,
		restRepo: restRepo,
	}
}

// Execute returns menu data for admin editing.
func (uc *GetMenuForAdminUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) (*MenuResponse, error) {
	restaurant, err := uc.restRepo.FindByID(restaurantID)
	if err != nil {
		return nil, err
	}

	categories, err := uc.catRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	sort.Slice(categories, func(i, j int) bool {
		if categories[i].DisplayOrder != categories[j].DisplayOrder {
			return categories[i].DisplayOrder < categories[j].DisplayOrder
		}
		return categories[i].Name < categories[j].Name
	})

	items, err := uc.itemRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	itemsByCategory := make(map[domain.CategoryID][]*domain.MenuItem)
	for _, item := range items {
		itemsByCategory[item.CategoryID] = append(itemsByCategory[item.CategoryID], item)
	}

	response := &MenuResponse{
		Restaurant: restaurant,
		Categories: make([]CategoryWithItems, 0, len(categories)),
	}

	for _, cat := range categories {
		catItems := itemsByCategory[cat.ID]
		sort.Slice(catItems, func(i, j int) bool {
			return catItems[i].Name < catItems[j].Name
		})
		response.Categories = append(response.Categories, CategoryWithItems{
			Category: cat,
			Items:    catItems,
		})
	}

	return response, nil
}
