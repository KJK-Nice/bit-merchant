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

func TestUpdateMenuItemUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	rid := common.RestaurantID("r1")
	cat1, err := menu.NewMenuCategory("c1", rid, "Drinks", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat1))
	cat2, err := menu.NewMenuCategory("c2", rid, "Food", 1)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat2))

	item, err := menu.NewMenuItem("i1", cat1.ID, rid, "Cola", 2)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))

	uc := menuCmd.NewUpdateMenuItemUseCase(repoItem, repoCat)
	err = uc.Execute(ctx, menuCmd.UpdateMenuItemRequest{
		RestaurantID: rid,
		ItemID:       item.ID,
		CategoryID:   cat2.ID,
		Name:         "Cola Zero",
		Description:  "iced",
		Price:        2.5,
		Available:    false,
	})
	require.NoError(t, err)

	updated, err := repoItem.FindByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, cat2.ID, updated.CategoryID)
	assert.Equal(t, "Cola Zero", updated.Name)
	assert.Equal(t, 2.5, updated.Price)
	assert.False(t, updated.IsAvailable)
	assert.Equal(t, "iced", updated.Description)
}

func TestUpdateMenuItemUseCase_WrongRestaurant(t *testing.T) {
	ctx := context.Background()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	r1 := common.RestaurantID("r1")
	r2 := common.RestaurantID("r2")
	cat, err := menu.NewMenuCategory("c1", r1, "X", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat))
	item, err := menu.NewMenuItem("i1", cat.ID, r1, "A", 1)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))

	uc := menuCmd.NewUpdateMenuItemUseCase(repoItem, repoCat)
	err = uc.Execute(ctx, menuCmd.UpdateMenuItemRequest{
		RestaurantID: r2,
		ItemID:       item.ID,
		CategoryID:   cat.ID,
		Name:         "B",
		Price:        2,
		Available:    true,
	})
	require.Error(t, err)
}
