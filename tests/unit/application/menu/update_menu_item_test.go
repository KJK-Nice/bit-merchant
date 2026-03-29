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

func TestUpdateMenuItemUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	rid := domain.RestaurantID("r1")
	cat1, err := domain.NewMenuCategory("c1", rid, "Drinks", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat1))
	cat2, err := domain.NewMenuCategory("c2", rid, "Food", 1)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat2))

	item, err := domain.NewMenuItem("i1", cat1.ID, rid, "Cola", 2)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))

	uc := menu.NewUpdateMenuItemUseCase(repoItem, repoCat)
	err = uc.Execute(ctx, menu.UpdateMenuItemRequest{
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

	r1 := domain.RestaurantID("r1")
	r2 := domain.RestaurantID("r2")
	cat, err := domain.NewMenuCategory("c1", r1, "X", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat))
	item, err := domain.NewMenuItem("i1", cat.ID, r1, "A", 1)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))

	uc := menu.NewUpdateMenuItemUseCase(repoItem, repoCat)
	err = uc.Execute(ctx, menu.UpdateMenuItemRequest{
		RestaurantID: r2,
		ItemID:       item.ID,
		CategoryID:   cat.ID,
		Name:         "B",
		Price:        2,
		Available:    true,
	})
	require.Error(t, err)
}
