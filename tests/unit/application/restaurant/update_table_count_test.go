package restaurant_test

import (
	"context"
	"testing"

	"bitmerchant/internal/infrastructure/repositories/memory"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRestaurantTableCountHandler(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	r, err := restaurant.NewRestaurant("r1", "Diner")
	require.NoError(t, err)
	require.NoError(t, repo.Save(r))

	h := restaurantCmd.NewUpdateRestaurantTableCountHandler(repo, nil, nil)

	t.Run("valid update", func(t *testing.T) {
		require.NoError(t, h.Handle(context.Background(), restaurantCmd.UpdateRestaurantTableCount{
			RestaurantID: "r1",
			TableCount:   12,
		}))
		updated, err := repo.FindByID("r1")
		require.NoError(t, err)
		assert.Equal(t, 12, updated.TableCount)
	})

	t.Run("validation error", func(t *testing.T) {
		err := h.Handle(context.Background(), restaurantCmd.UpdateRestaurantTableCount{
			RestaurantID: "r1",
			TableCount:   0,
		})
		assert.Error(t, err)
		err = h.Handle(context.Background(), restaurantCmd.UpdateRestaurantTableCount{
			RestaurantID: "r1",
			TableCount:   999,
		})
		assert.Error(t, err)
	})
}
