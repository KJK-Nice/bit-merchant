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

func TestUpdateMenuCategoryUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewMemoryMenuCategoryRepository()
	rid := domain.RestaurantID("r1")
	cat, err := domain.NewMenuCategory("c1", rid, "Old", 3)
	require.NoError(t, err)
	require.NoError(t, repo.Save(cat))

	uc := menu.NewUpdateMenuCategoryUseCase(repo)
	err = uc.Execute(ctx, menu.UpdateMenuCategoryRequest{
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
