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

type adminFakePhotoStorage struct{}

func (adminFakePhotoStorage) Upload(context.Context, string, io.Reader, string) (string, error) {
	panic("unused")
}
func (adminFakePhotoStorage) Delete(context.Context, string) error { return nil }
func (adminFakePhotoStorage) PresignGet(_ context.Context, key string) (string, error) {
	return "https://signed.example/" + key, nil
}

func TestMenuForAdminHandler_IncludesUnavailableItemsAndEmptyCategories(t *testing.T) {
	ctx := context.Background()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	rid := common.RestaurantID("r1")
	rest, err := restaurant.NewRestaurant(rid, "Cafe")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	catEmpty, err := menu.NewMenuCategory("cat_empty", rid, "Empty Section", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(catEmpty))

	catWith, err := menu.NewMenuCategory("cat_items", rid, "Mains", 1)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(catWith))

	avail, err := menu.NewMenuItem("it1", catWith.ID, rid, "Burger", 9)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(avail))

	unavail, err := menu.NewMenuItem("it2", catWith.ID, rid, "Soup", 5)
	require.NoError(t, err)
	unavail.SetAvailable(false)
	require.NoError(t, repoItem.Save(unavail))

	uc := menuQuery.NewMenuForAdminHandler(repoCat, repoItem, repoRest, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	resp, err := uc.Handle(ctx, menuQuery.MenuForAdmin{RestaurantID: rid})
	require.NoError(t, err)
	require.Len(t, resp.Categories, 2)

	var emptyFound bool
	var mains *menuQuery.CategoryWithItems
	for i := range resp.Categories {
		if resp.Categories[i].Category.ID == catEmpty.ID {
			emptyFound = true
			assert.Empty(t, resp.Categories[i].Items)
		}
		if resp.Categories[i].Category.ID == catWith.ID {
			mains = &resp.Categories[i]
		}
	}
	assert.True(t, emptyFound)
	require.NotNil(t, mains)
	require.Len(t, mains.Items, 2)

	names := make(map[string]bool)
	for _, it := range mains.Items {
		names[it.Name] = it.IsAvailable
	}
	assert.True(t, names["Burger"])
	assert.False(t, names["Soup"])

	// Public menu omits unavailable and empty categories
	pub := menuQuery.NewMenuForCustomerHandler(repoCat, repoItem, repoRest, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	pubResp, err := pub.Handle(ctx, menuQuery.MenuForCustomer{RestaurantID: rid})
	require.NoError(t, err)
	require.Len(t, pubResp.Categories, 1)
	assert.Equal(t, catWith.ID, pubResp.Categories[0].Category.ID)
	require.Len(t, pubResp.Categories[0].Items, 1)
	assert.Equal(t, "Burger", pubResp.Categories[0].Items[0].Name)
}

func TestMenuForAdminHandler_PresignsPhotoURLs(t *testing.T) {
	ctx := context.Background()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	rid := common.RestaurantID("r1")
	rest, err := restaurant.NewRestaurant(rid, "Cafe")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	cat, err := menu.NewMenuCategory("c1", rid, "Mains", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat))

	it, err := menu.NewMenuItem("i1", cat.ID, rid, "Burger", 9)
	require.NoError(t, err)
	it.SetPhotoURLs("restaurants/r1/items/x.jpg", "restaurants/r1/items/x.jpg")
	require.NoError(t, repoItem.Save(it))

	uc := menuQuery.NewMenuForAdminHandler(repoCat, repoItem, repoRest, adminFakePhotoStorage{}, menuQuery.PhotoSignerConfig{Bucket: "b"}, nil, nil)
	resp, err := uc.Handle(ctx, menuQuery.MenuForAdmin{RestaurantID: rid})
	require.NoError(t, err)
	require.Len(t, resp.Categories, 1)
	require.Len(t, resp.Categories[0].Items, 1)
	assert.Contains(t, resp.Categories[0].Items[0].PhotoURL, "https://signed.example/")
}
