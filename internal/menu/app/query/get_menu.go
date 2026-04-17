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

// PhotoSignerConfig configures how stored photo references are turned into presigned URLs.
type PhotoSignerConfig struct {
	Bucket        string
	Endpoint      string
	PublicBaseURL string
}

// MenuForCustomer loads the customer-facing menu (active categories with available items).
type MenuForCustomer struct {
	RestaurantID common.RestaurantID
}

type MenuForCustomerHandler decorator.QueryHandler[MenuForCustomer, *MenuResponse]

type menuForCustomerHandler struct {
	catRepo        menu.CategoryRepository
	itemRepo       menu.ItemRepository
	restRepo       restaurant.Repository
	photos         menu.PhotoStorage
	photoSignerCfg PhotoSignerConfig
}

func NewMenuForCustomerHandler(catRepo menu.CategoryRepository, itemRepo menu.ItemRepository, restRepo restaurant.Repository, photos menu.PhotoStorage, photoSignerCfg PhotoSignerConfig, log *slog.Logger, metrics decorator.MetricsClient) MenuForCustomerHandler {
	h := menuForCustomerHandler{
		catRepo:        catRepo,
		itemRepo:       itemRepo,
		restRepo:       restRepo,
		photos:         photos,
		photoSignerCfg: photoSignerCfg,
	}
	return decorator.ApplyQueryDecorators[MenuForCustomer, *MenuResponse](h, log, metrics)
}

type MenuResponse struct {
	Restaurant *restaurant.Restaurant
	Categories []CategoryWithItems
}

type CategoryWithItems struct {
	Category *menu.MenuCategory
	Items    []*menu.MenuItem
}

func (h menuForCustomerHandler) Handle(ctx context.Context, q MenuForCustomer) (*MenuResponse, error) {
	rest, err := h.restRepo.FindByID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	categories, err := h.catRepo.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DisplayOrder < categories[j].DisplayOrder
	})

	items, err := h.itemRepo.FindAvailableByRestaurantID(q.RestaurantID)
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
