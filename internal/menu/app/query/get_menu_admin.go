package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type GetMenuForAdminUseCase struct {
	catRepo  menu.CategoryRepository
	itemRepo menu.ItemRepository
	restRepo restaurant.Repository
}

func NewGetMenuForAdminUseCase(catRepo menu.CategoryRepository, itemRepo menu.ItemRepository, restRepo restaurant.Repository) *GetMenuForAdminUseCase {
	return &GetMenuForAdminUseCase{
		catRepo:  catRepo,
		itemRepo: itemRepo,
		restRepo: restRepo,
	}
}

func (uc *GetMenuForAdminUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID) (*MenuResponse, error) {
	rest, err := uc.restRepo.FindByID(restaurantID)
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

	itemsByCategory := make(map[common.CategoryID][]*menu.MenuItem)
	for _, item := range items {
		itemsByCategory[item.CategoryID] = append(itemsByCategory[item.CategoryID], item)
	}

	response := &MenuResponse{
		Restaurant: rest,
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
