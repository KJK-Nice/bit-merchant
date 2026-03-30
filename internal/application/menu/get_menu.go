package menu

import (
	"context"
	"fmt"
	"sort"

	"bitmerchant/internal/domain"
)

// PhotoSignerConfig is used to recover S3 object keys from legacy full URLs stored in the database.
type PhotoSignerConfig struct {
	Bucket        string
	Endpoint      string
	PublicBaseURL string
}

// GetMenuUseCase retrieves restaurant menu
type GetMenuUseCase struct {
	catRepo        domain.MenuCategoryRepository
	itemRepo       domain.MenuItemRepository
	restRepo       domain.RestaurantRepository
	photos         domain.PhotoStorage
	photoSignerCfg PhotoSignerConfig
}

// NewGetMenuUseCase creates a new GetMenuUseCase.
// photos may be nil when S3 is not configured; photoSignerCfg is only used when photos is non-nil.
func NewGetMenuUseCase(catRepo domain.MenuCategoryRepository, itemRepo domain.MenuItemRepository, restRepo domain.RestaurantRepository, photos domain.PhotoStorage, photoSignerCfg PhotoSignerConfig) *GetMenuUseCase {
	return &GetMenuUseCase{
		catRepo:        catRepo,
		itemRepo:       itemRepo,
		restRepo:       restRepo,
		photos:         photos,
		photoSignerCfg: photoSignerCfg,
	}
}

// MenuResponse represents the menu structure
type MenuResponse struct {
	Restaurant *domain.Restaurant
	Categories []CategoryWithItems
}

// CategoryWithItems represents a category with its items
type CategoryWithItems struct {
	Category *domain.MenuCategory
	Items    []*domain.MenuItem
}

// Execute retrieves the customer-facing menu: available items only, active categories only,
// and omits empty categories.
func (uc *GetMenuUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) (*MenuResponse, error) {
	// Get restaurant
	restaurant, err := uc.restRepo.FindByID(restaurantID)
	if err != nil {
		return nil, err
	}

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
		Restaurant: restaurant,
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

func (uc *GetMenuUseCase) itemsWithPresignedPhotos(ctx context.Context, items []*domain.MenuItem) ([]*domain.MenuItem, error) {
	if uc.photos == nil {
		return items, nil
	}

	out := make([]*domain.MenuItem, len(items))
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
