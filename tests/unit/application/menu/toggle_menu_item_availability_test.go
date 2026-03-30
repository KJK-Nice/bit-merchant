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

func TestToggleMenuItemAvailabilityUseCase(t *testing.T) {
	ctx := context.Background()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	rid := domain.RestaurantID("r1")
	cat, err := domain.NewMenuCategory("c1", rid, "M", 0)
	require.NoError(t, err)
	require.NoError(t, repoCat.Save(cat))
	item, err := domain.NewMenuItem("i1", cat.ID, rid, "X", 1)
	require.NoError(t, err)
	require.NoError(t, repoItem.Save(item))
	assert.True(t, item.IsAvailable)

	uc := menu.NewToggleMenuItemAvailabilityUseCase(repoItem)
	require.NoError(t, uc.Execute(ctx, rid, item.ID))
	after, err := repoItem.FindByID(item.ID)
	require.NoError(t, err)
	assert.False(t, after.IsAvailable)

	require.NoError(t, uc.Execute(ctx, rid, item.ID))
	after2, err := repoItem.FindByID(item.ID)
	require.NoError(t, err)
	assert.True(t, after2.IsAvailable)
}
