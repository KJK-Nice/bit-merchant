package menu_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type fakePhotoStorage struct{}

func (fakePhotoStorage) Upload(context.Context, string, io.Reader, string) (string, error) {
	panic("unused")
}

func (fakePhotoStorage) Delete(context.Context, string) error { return nil }

func (fakePhotoStorage) PresignGet(_ context.Context, key string) (string, error) {
	return "https://signed.example/" + key, nil
}

func TestMenuForCustomerHandler(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	uc := menuQuery.NewMenuForCustomerHandler(catRepo, itemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{}, nil, nil)

	restID := common.RestaurantID("r1")
	restaurant, _ := restaurant.NewRestaurant(restID, "Test Restaurant")
	require.NoError(t, restRepo.Save(restaurant))

	// Setup data
	cat1, _ := menu.NewMenuCategory("c1", restID, "Starters", 1)
	cat2, _ := menu.NewMenuCategory("c2", restID, "Mains", 2)
	require.NoError(t, catRepo.Save(cat1))
	require.NoError(t, catRepo.Save(cat2))

	item1, _ := menu.NewMenuItem("i1", "c1", restID, "Salad", 10.0)
	item2, _ := menu.NewMenuItem("i2", "c2", restID, "Steak", 20.0)
	require.NoError(t, itemRepo.Save(item1))
	require.NoError(t, itemRepo.Save(item2))

	t.Run("Handle", func(t *testing.T) {
		result, err := uc.Handle(context.Background(), menuQuery.MenuForCustomer{RestaurantID: restID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Restaurant", result.Restaurant.Name)
		assert.Len(t, result.Categories, 2)

		// Check structure
		// Assuming result structure: struct { Categories []CategoryWithItems }
		// CategoryWithItems has Items field

		// For now, assuming pure domain entities returned or DTOs.
		// Spec implies displaying menu organized by categories.
		// Let's assume DTO structure for now or check implementation plan.
		// Implementation plan says: "returns restaurant menu with categories and items"
	})

	t.Run("presigned_photo_urls", func(t *testing.T) {
		ucPhoto := menuQuery.NewMenuForCustomerHandler(catRepo, itemRepo, restRepo, fakePhotoStorage{}, menuQuery.PhotoSignerConfig{
			Bucket: "mybucket",
		}, nil, nil)
		itemPhoto, _ := menu.NewMenuItem("iphoto", "c1", restID, "With Pix", 10.0)
		itemPhoto.SetPhotoURLs("restaurants/r1/items/x.jpg", "restaurants/r1/items/x.jpg")
		require.NoError(t, itemRepo.Save(itemPhoto))

		result, err := ucPhoto.Handle(context.Background(), menuQuery.MenuForCustomer{RestaurantID: restID})
		require.NoError(t, err)
		var found bool
		for _, c := range result.Categories {
			for _, it := range c.Items {
				if it.ID == itemPhoto.ID {
					found = true
					assert.Equal(t, "https://signed.example/restaurants/r1/items/x.jpg", it.PhotoURL)
				}
			}
		}
		assert.True(t, found)
	})
}
