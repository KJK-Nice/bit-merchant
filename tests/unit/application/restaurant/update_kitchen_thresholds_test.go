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

func TestUpdateKitchenThresholdsHandler(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	r, err := restaurant.NewRestaurant("r1", "Diner")
	require.NoError(t, err)
	require.NoError(t, repo.Save(r))

	h := restaurantCmd.NewUpdateKitchenThresholdsHandler(repo, nil, nil)

	t.Run("valid update persists both thresholds", func(t *testing.T) {
		require.NoError(t, h.Handle(context.Background(), restaurantCmd.UpdateKitchenThresholds{
			RestaurantID:   "r1",
			WarningMinutes: 6,
			OverdueMinutes: 15,
		}))
		updated, err := repo.FindByID("r1")
		require.NoError(t, err)
		assert.Equal(t, 6, updated.KitchenWarningMinutes)
		assert.Equal(t, 15, updated.KitchenOverdueMinutes)
	})

	t.Run("rejects warning >= overdue", func(t *testing.T) {
		err := h.Handle(context.Background(), restaurantCmd.UpdateKitchenThresholds{
			RestaurantID:   "r1",
			WarningMinutes: 12,
			OverdueMinutes: 12,
		})
		assert.ErrorIs(t, err, restaurant.ErrInvalidKitchenThresholds)
	})

	t.Run("rejects out-of-range values", func(t *testing.T) {
		assert.Error(t, h.Handle(context.Background(), restaurantCmd.UpdateKitchenThresholds{
			RestaurantID:   "r1",
			WarningMinutes: 0,
			OverdueMinutes: 10,
		}))
		assert.Error(t, h.Handle(context.Background(), restaurantCmd.UpdateKitchenThresholds{
			RestaurantID:   "r1",
			WarningMinutes: 5,
			OverdueMinutes: 999,
		}))
	})
}
