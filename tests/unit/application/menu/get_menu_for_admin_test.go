package menu_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMenuForAdminUseCase_IncludesUnavailableItemsAndEmptyCategories(t *testing.T) {
	ctx := context.Background()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	rid := domain.RestaurantID("r1")
	rest, err := domain.NewRestaurant(rid, "Cafe")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	catEmpty, err := domain.NewMenuCategory("cat_empty", rid, "Empty Section", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(catEmpty))

	catWith, err := domain.NewMenuCategory("cat_items", rid, "Mains", 1)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(catWith))

	avail, err := domain.NewMenuItem("it1", catWith.ID, rid, "Burger", 9)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(avail))

	unavail, err := domain.NewMenuItem("it2", catWith.ID, rid, "Soup", 5)
	require.NoError(t, err)
	unavail.SetAvailable(false)
	require.NoError(t, repoItem.Save(unavail))

	uc := menu.NewGetMenuForAdminUseCase(repoCat, repoItem, repoRest)
	resp, err := uc.Execute(ctx, rid)
	require.NoError(t, err)
	require.Len(t, resp.Categories, 2)

	var emptyFound bool
	var mains *menu.CategoryWithItems
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
	pub := menu.NewGetMenuUseCase(repoCat, repoItem, repoRest)
	pubResp, err := pub.Execute(ctx, rid)
	require.NoError(t, err)
	require.Len(t, pubResp.Categories, 1)
	assert.Equal(t, catWith.ID, pubResp.Categories[0].Category.ID)
	require.Len(t, pubResp.Categories[0].Items, 1)
	assert.Equal(t, "Burger", pubResp.Categories[0].Items[0].Name)
}
