package admin_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPhotoStorage struct{}

func (stubPhotoStorage) Upload(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
	return key, nil
}

func (stubPhotoStorage) Delete(context.Context, string) error { return nil }

func (stubPhotoStorage) PresignGet(_ context.Context, key string) (string, error) {
	return "https://signed.example/" + key, nil
}

func TestMenuSetupWorkflow(t *testing.T) {
	// Setup
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	var mockStorage stubPhotoStorage

	createRestUC := restaurant.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menu.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menu.NewCreateMenuItemUseCase(repoItem)
	uploadPhotoUC := menu.NewUploadPhotoUseCase(repoItem, mockStorage)

	// 1. Create Restaurant
	restReq := restaurant.CreateRestaurantRequest{Name: "Burger King"}
	rest, err := createRestUC.Execute(context.Background(), restReq)
	require.NoError(t, err)
	require.NotEmpty(t, rest.ID)

	// 2. Create Category
	catReq := menu.CreateMenuCategoryRequest{
		RestaurantID: rest.ID,
		Name:         "Burgers",
		DisplayOrder: 1,
	}
	cat, err := createCatUC.Execute(context.Background(), catReq)
	require.NoError(t, err)
	require.NotEmpty(t, cat.ID)

	// 3. Create Item
	itemReq := menu.CreateMenuItemRequest{
		RestaurantID: rest.ID,
		CategoryID:   cat.ID,
		Name:         "Whopper",
		Price:        5.99,
		Available:    true,
	}
	item, err := createItemUC.Execute(context.Background(), itemReq)
	require.NoError(t, err)
	require.NotEmpty(t, item.ID)

	// 4. Upload Photo
	photoReq := menu.UploadPhotoRequest{
		RestaurantID: rest.ID,
		ItemID:       item.ID,
		File:         bytes.NewBufferString("fake image data"),
		Filename:     "whopper.jpg",
		ContentType:  "image/jpeg",
	}

	storedKey, err := uploadPhotoUC.Execute(context.Background(), photoReq)
	require.NoError(t, err)
	assert.Contains(t, storedKey, "restaurants/")
	assert.Contains(t, storedKey, string(item.ID))

	// Verify Item Updated (object key, not a public URL)
	updatedItem, _ := repoItem.FindByID(item.ID)
	assert.Equal(t, storedKey, updatedItem.PhotoURL)
}
