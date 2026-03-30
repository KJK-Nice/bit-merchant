package restaurant_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRestaurantTableCountUseCase(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	r, err := domain.NewRestaurant("r1", "Diner")
	require.NoError(t, err)
	require.NoError(t, repo.Save(r))

	uc := restaurant.NewUpdateRestaurantTableCountUseCase(repo)

	t.Run("valid update", func(t *testing.T) {
		require.NoError(t, uc.Execute(context.Background(), "r1", 12))
		updated, err := repo.FindByID("r1")
		require.NoError(t, err)
		assert.Equal(t, 12, updated.TableCount)
	})

	t.Run("validation error", func(t *testing.T) {
		err := uc.Execute(context.Background(), "r1", 0)
		assert.Error(t, err)
		err = uc.Execute(context.Background(), "r1", 999)
		assert.Error(t, err)
	})
}
