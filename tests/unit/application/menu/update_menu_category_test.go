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

func TestUpdateMenuCategoryUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewMemoryMenuCategoryRepository()
	rid := common.RestaurantID("r1")
	cat, err := menu.NewMenuCategory("c1", rid, "Old", 3)
	require.NoError(t, err)
	require.NoError(t, repo.Save(cat))

	uc := menuCmd.NewUpdateMenuCategoryUseCase(repo)
	err = uc.Execute(ctx, menuCmd.UpdateMenuCategoryRequest{
		RestaurantID: rid,
		CategoryID:   cat.ID,
		Name:         "New Name",
		DisplayOrder: 10,
		IsActive:     false,
	})
	require.NoError(t, err)

	out, err := repo.FindByID(cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", out.Name)
	assert.Equal(t, 10, out.DisplayOrder)
	assert.False(t, out.IsActive)
}
