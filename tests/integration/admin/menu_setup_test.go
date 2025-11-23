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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPhotoStorage
type MockPhotoStorage struct {
	mock.Mock
}

func (m *MockPhotoStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockPhotoStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestMenuSetupWorkflow(t *testing.T) {
	// Setup
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	mockStorage := new(MockPhotoStorage)

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
	mockStorage.On("Upload", mock.Anything, mock.AnythingOfType("string"), mock.Anything, "image/jpeg").
		Return("https://s3.aws.com/bucket/photo.jpg", nil)

	photoReq := menu.UploadPhotoRequest{
		RestaurantID: rest.ID,
		ItemID:       item.ID,
		File:         bytes.NewBufferString("fake image data"),
		Filename:     "whopper.jpg",
		ContentType:  "image/jpeg",
	}

	url, err := uploadPhotoUC.Execute(context.Background(), photoReq)
	require.NoError(t, err)
	assert.Equal(t, "https://s3.aws.com/bucket/photo.jpg", url)

	// Verify Item Updated
	updatedItem, _ := repoItem.FindByID(item.ID)
	assert.Equal(t, "https://s3.aws.com/bucket/photo.jpg", updatedItem.PhotoURL)
}

