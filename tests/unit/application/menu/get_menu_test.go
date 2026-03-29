package menu_test

import (
	"context"
	"io"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePhotoStorage struct{}

func (fakePhotoStorage) Upload(context.Context, string, io.Reader, string) (string, error) {
	panic("unused")
}

func (fakePhotoStorage) Delete(context.Context, string) error { return nil }

func (fakePhotoStorage) PresignGet(_ context.Context, key string) (string, error) {
	return "https://signed.example/" + key, nil
}

func TestGetMenuUseCase(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	uc := menu.NewGetMenuUseCase(catRepo, itemRepo, restRepo, nil, menu.PhotoSignerConfig{})

	restID := domain.RestaurantID("r1")
	restaurant, _ := domain.NewRestaurant(restID, "Test Restaurant")
	require.NoError(t, restRepo.Save(restaurant))

	// Setup data
	cat1, _ := domain.NewMenuCategory("c1", restID, "Starters", 1)
	cat2, _ := domain.NewMenuCategory("c2", restID, "Mains", 2)
	require.NoError(t, catRepo.Save(cat1))
	require.NoError(t, catRepo.Save(cat2))

	item1, _ := domain.NewMenuItem("i1", "c1", restID, "Salad", 10.0)
	item2, _ := domain.NewMenuItem("i2", "c2", restID, "Steak", 20.0)
	require.NoError(t, itemRepo.Save(item1))
	require.NoError(t, itemRepo.Save(item2))

	t.Run("Execute", func(t *testing.T) {
		result, err := uc.Execute(context.Background(), restID)
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
		ucPhoto := menu.NewGetMenuUseCase(catRepo, itemRepo, restRepo, fakePhotoStorage{}, menu.PhotoSignerConfig{
			Bucket: "mybucket",
		})
		itemPhoto, _ := domain.NewMenuItem("iphoto", "c1", restID, "With Pix", 10.0)
		itemPhoto.SetPhotoURLs("restaurants/r1/items/x.jpg", "restaurants/r1/items/x.jpg")
		require.NoError(t, itemRepo.Save(itemPhoto))

		result, err := ucPhoto.Execute(context.Background(), restID)
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
