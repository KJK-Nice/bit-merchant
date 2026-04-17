package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// MenuForAdmin loads the full menu for merchant administration.
type MenuForAdmin struct {
	RestaurantID common.RestaurantID
}

type MenuForAdminHandler decorator.QueryHandler[MenuForAdmin, *MenuResponse]

type menuForAdminHandler struct {
	catRepo        menu.CategoryRepository
	itemRepo       menu.ItemRepository
	restRepo       restaurant.Repository
	photos         menu.PhotoStorage
	photoSignerCfg PhotoSignerConfig
}

func NewMenuForAdminHandler(catRepo menu.CategoryRepository, itemRepo menu.ItemRepository, restRepo restaurant.Repository, photos menu.PhotoStorage, photoSignerCfg PhotoSignerConfig, log *slog.Logger, metrics decorator.MetricsClient) MenuForAdminHandler {
	h := menuForAdminHandler{
		catRepo:        catRepo,
		itemRepo:       itemRepo,
		restRepo:       restRepo,
		photos:         photos,
		photoSignerCfg: photoSignerCfg,
	}
	return decorator.ApplyQueryDecorators[MenuForAdmin, *MenuResponse](h, log, metrics)
}

func (h menuForAdminHandler) Handle(ctx context.Context, q MenuForAdmin) (*MenuResponse, error) {
	rest, err := h.restRepo.FindByID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	categories, err := h.catRepo.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	sort.Slice(categories, func(i, j int) bool {
		if categories[i].DisplayOrder != categories[j].DisplayOrder {
			return categories[i].DisplayOrder < categories[j].DisplayOrder
		}
		return categories[i].Name < categories[j].Name
	})

	items, err := h.itemRepo.FindByRestaurantID(q.RestaurantID)
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
		displayItems, err := ItemsWithPresignedPhotos(ctx, catItems, h.photos, h.photoSignerCfg)
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
