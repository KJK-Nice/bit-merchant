package admin_test

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	menuCmd "bitmerchant/internal/menu/app/command"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bytes"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
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

	createRestUC := restaurantCmd.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menuCmd.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menuCmd.NewCreateMenuItemUseCase(repoItem)
	uploadPhotoUC := menuCmd.NewUploadPhotoUseCase(repoItem, mockStorage)

	// 1. Create Restaurant
	restReq := restaurantCmd.CreateRestaurantRequest{Name: "Burger King"}
	rest, err := createRestUC.Execute(context.Background(), restReq)
	require.NoError(t, err)
	require.NotEmpty(t, rest.ID)

	// 2. Create Category
	catReq := menuCmd.CreateMenuCategoryRequest{
		RestaurantID: rest.ID,
		Name:         "Burgers",
		DisplayOrder: 1,
	}
	cat, err := createCatUC.Execute(context.Background(), catReq)
	require.NoError(t, err)
	require.NotEmpty(t, cat.ID)

	// 3. Create Item
	itemReq := menuCmd.CreateMenuItemRequest{
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
	photoReq := menuCmd.UploadPhotoRequest{
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
