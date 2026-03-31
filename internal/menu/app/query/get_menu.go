package query

import (
	"context"
	"fmt"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type PhotoSignerConfig struct {
	Bucket        string
	Endpoint      string
	PublicBaseURL string
}

type GetMenuUseCase struct {
	catRepo        menu.CategoryRepository
	itemRepo       menu.ItemRepository
	restRepo       restaurant.Repository
	photos         menu.PhotoStorage
	photoSignerCfg PhotoSignerConfig
}

func NewGetMenuUseCase(catRepo menu.CategoryRepository, itemRepo menu.ItemRepository, restRepo restaurant.Repository, photos menu.PhotoStorage, photoSignerCfg PhotoSignerConfig) *GetMenuUseCase {
	return &GetMenuUseCase{
		catRepo:        catRepo,
		itemRepo:       itemRepo,
		restRepo:       restRepo,
		photos:         photos,
		photoSignerCfg: photoSignerCfg,
	}
}

type MenuResponse struct {
	Restaurant *restaurant.Restaurant
	Categories []CategoryWithItems
}

type CategoryWithItems struct {
	Category *menu.MenuCategory
	Items    []*menu.MenuItem
}

func (uc *GetMenuUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID) (*MenuResponse, error) {
	rest, err := uc.restRepo.FindByID(restaurantID)
	if err != nil {
		return nil, err
	}

	categories, err := uc.catRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DisplayOrder < categories[j].DisplayOrder
	})

	items, err := uc.itemRepo.FindAvailableByRestaurantID(restaurantID)
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
		if !cat.IsActive {
			continue
		}
		catItems := itemsByCategory[cat.ID]
		if len(catItems) == 0 {
			continue
		}

		displayItems, err := uc.itemsWithPresignedPhotos(ctx, catItems)
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

func (uc *GetMenuUseCase) itemsWithPresignedPhotos(ctx context.Context, items []*menu.MenuItem) ([]*menu.MenuItem, error) {
	if uc.photos == nil {
		return items, nil
	}

	out := make([]*menu.MenuItem, len(items))
	for i, item := range items {
		cp := *item
		if cp.PhotoURL != "" {
			key := PhotoObjectKeyFromStoredValue(cp.PhotoURL, uc.photoSignerCfg.Bucket, uc.photoSignerCfg.Endpoint, uc.photoSignerCfg.PublicBaseURL)
			if key == "" {
				key = cp.PhotoURL
			}
			signed, err := uc.photos.PresignGet(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("presign menu photo: %w", err)
			}
			cp.PhotoURL = signed
		}
		out[i] = &cp
	}
	return out, nil
}
