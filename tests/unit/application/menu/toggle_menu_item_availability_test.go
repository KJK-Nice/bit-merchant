package menu_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	menuCmd "bitmerchant/internal/menu/app/command"
	"bitmerchant/internal/menu/domain/menu"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToggleMenuItemAvailabilityHandler(t *testing.T) {
	ctx := context.Background()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	rid := common.RestaurantID("r1")
	cat, err := menu.NewMenuCategory("c1", rid, "M", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat))
	item, err := menu.NewMenuItem("i1", cat.ID, rid, "X", 1)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))
	assert.True(t, item.IsAvailable)

	uc := menuCmd.NewToggleMenuItemAvailabilityHandler(repoItem, nil, nil)
	require.NoError(t, uc.Handle(ctx, menuCmd.ToggleMenuItemAvailability{RestaurantID: rid, ItemID: item.ID}))
	after, err := repoItem.FindByID(item.ID)
	require.NoError(t, err)
	assert.False(t, after.IsAvailable)

	require.NoError(t, uc.Handle(ctx, menuCmd.ToggleMenuItemAvailability{RestaurantID: rid, ItemID: item.ID}))
	after2, err := repoItem.FindByID(item.ID)
	require.NoError(t, err)
	assert.True(t, after2.IsAvailable)
}
