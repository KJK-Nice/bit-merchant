package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type GetMenuForAdminUseCase struct {
	catRepo        menu.CategoryRepository
	itemRepo       menu.ItemRepository
	restRepo       restaurant.Repository
	photos         menu.PhotoStorage
	photoSignerCfg PhotoSignerConfig
}

func NewGetMenuForAdminUseCase(catRepo menu.CategoryRepository, itemRepo menu.ItemRepository, restRepo restaurant.Repository, photos menu.PhotoStorage, photoSignerCfg PhotoSignerConfig) *GetMenuForAdminUseCase {
	return &GetMenuForAdminUseCase{
		catRepo:        catRepo,
		itemRepo:       itemRepo,
		restRepo:       restRepo,
		photos:         photos,
		photoSignerCfg: photoSignerCfg,
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
			if catItems[i].DisplayOrder != catItems[j].DisplayOrder {
				return catItems[i].DisplayOrder < catItems[j].DisplayOrder
			}
			return catItems[i].Name < catItems[j].Name
		})
		displayItems, err := ItemsWithPresignedPhotos(ctx, catItems, uc.photos, uc.photoSignerCfg)
		if err != nil {
			return nil, err
		}
		response.Categories = append(response.Categories, CategoryWithItems{
			Category: cat,
			Items:    displayItems,
		})
	}

	return response, nil
}
