package domain_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewRestaurant(t *testing.T) {
	t.Run("should create valid restaurant", func(t *testing.T) {
		id := common.RestaurantID("rest_123")
		name := "Tasty Burger"

		gotRestaurant, err := restaurant.NewRestaurant(id, name)

		assert.NoError(t, err)
		assert.NotNil(t, gotRestaurant)
		assert.Equal(t, id, gotRestaurant.ID)
		assert.Equal(t, name, gotRestaurant.Name)
		assert.True(t, gotRestaurant.IsOpen)
		assert.Equal(t, restaurant.MinTableCount, gotRestaurant.TableCount)
		assert.WithinDuration(t, time.Now(), gotRestaurant.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), gotRestaurant.UpdatedAt, time.Second)
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := restaurant.NewRestaurant("id", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restaurant name must be between 1 and 100 characters")
	})

	t.Run("should fail with too long name", func(t *testing.T) {
		longName := ""
		for i := 0; i < 101; i++ {
			longName += "a"
		}
		_, err := restaurant.NewRestaurant("id", longName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restaurant name must be between 1 and 100 characters")
	})
}

func TestValidateTableCount(t *testing.T) {
	assert.NoError(t, restaurant.ValidateTableCount(1))
	assert.NoError(t, restaurant.ValidateTableCount(200))
	assert.Error(t, restaurant.ValidateTableCount(0))
	assert.Error(t, restaurant.ValidateTableCount(201))
}

func TestNewRestaurant_DefaultKitchenThresholds(t *testing.T) {
	r, err := restaurant.NewRestaurant("id", "Diner")
	assert.NoError(t, err)
	assert.Equal(t, restaurant.DefaultKitchenWarningMinutes, r.KitchenWarningMinutes)
	assert.Equal(t, restaurant.DefaultKitchenOverdueMinutes, r.KitchenOverdueMinutes)
}

func TestValidateKitchenThresholds(t *testing.T) {
	assert.NoError(t, restaurant.ValidateKitchenThresholds(8, 12))
	assert.NoError(t, restaurant.ValidateKitchenThresholds(1, 120))
	// warning must be < overdue
	assert.ErrorIs(t, restaurant.ValidateKitchenThresholds(12, 12), restaurant.ErrInvalidKitchenThresholds)
	assert.ErrorIs(t, restaurant.ValidateKitchenThresholds(15, 10), restaurant.ErrInvalidKitchenThresholds)
	// out of range
	assert.ErrorIs(t, restaurant.ValidateKitchenThresholds(0, 10), restaurant.ErrInvalidKitchenThresholds)
	assert.ErrorIs(t, restaurant.ValidateKitchenThresholds(5, 121), restaurant.ErrInvalidKitchenThresholds)
}

func TestSetKitchenThresholds(t *testing.T) {
	r, _ := restaurant.NewRestaurant("id", "Diner")
	require.NoError(t, r.SetKitchenThresholds(6, 14))
	assert.Equal(t, 6, r.KitchenWarningMinutes)
	assert.Equal(t, 14, r.KitchenOverdueMinutes)

	// invalid input leaves the values unchanged
	err := r.SetKitchenThresholds(20, 10)
	assert.ErrorIs(t, err, restaurant.ErrInvalidKitchenThresholds)
	assert.Equal(t, 6, r.KitchenWarningMinutes)
	assert.Equal(t, 14, r.KitchenOverdueMinutes)
}

func TestEffectiveKitchenThresholds_FallBackForLegacyZero(t *testing.T) {
	r := &restaurant.Restaurant{} // legacy row: zero thresholds
	assert.Equal(t, restaurant.DefaultKitchenWarningMinutes, r.EffectiveKitchenWarningMinutes())
	assert.Equal(t, restaurant.DefaultKitchenOverdueMinutes, r.EffectiveKitchenOverdueMinutes())
}

func TestRestaurant_UpdateStatus(t *testing.T) {
	r, _ := restaurant.NewRestaurant("id", "name")
	originalUpdate := r.UpdatedAt

	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)

	r.UpdateStatus(false, "Closed for holiday", "Tomorrow 9am")

	assert.False(t, r.IsOpen)
	assert.Equal(t, "Closed for holiday", r.ClosedMessage)
	assert.Equal(t, "Tomorrow 9am", r.ReopeningHours)
	assert.True(t, r.UpdatedAt.After(originalUpdate))
}

func TestRestaurant_PauseAndResume(t *testing.T) {
	r, _ := restaurant.NewRestaurant("id", "Cafe")
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	t.Run("pause within bounds suppresses accepting orders", func(t *testing.T) {
		err := r.Pause(now, 15*time.Minute)
		assert.NoError(t, err)
		assert.True(t, r.IsOpen)
		assert.True(t, r.IsPausedAt(now))
		assert.False(t, r.AcceptingOrdersAt(now))
		assert.True(t, r.AcceptingOrdersAt(now.Add(16*time.Minute)))
	})

	t.Run("resume clears the pause window", func(t *testing.T) {
		r.Resume()
		assert.True(t, r.IsOpen)
		assert.False(t, r.IsPausedAt(now))
		assert.True(t, r.AcceptingOrdersAt(now))
		assert.Nil(t, r.PausedUntil)
	})

	t.Run("pause rejects bad durations", func(t *testing.T) {
		assert.ErrorIs(t, r.Pause(now, 0), restaurant.ErrInvalidPauseDuration)
		assert.ErrorIs(t, r.Pause(now, restaurant.MaxPauseDuration+time.Minute), restaurant.ErrInvalidPauseDuration)
	})

	t.Run("closing clears any active pause", func(t *testing.T) {
		_ = r.Pause(now, 30*time.Minute)
		r.Close("brb", "Tomorrow 9am")
		assert.False(t, r.IsOpen)
		assert.Nil(t, r.PausedUntil)
		assert.False(t, r.AcceptingOrdersAt(now))
	})

	t.Run("opening clears any leftover pause", func(t *testing.T) {
		r.PausedUntil = ptrTime(now.Add(5 * time.Minute))
		r.Open()
		assert.True(t, r.IsOpen)
		assert.Nil(t, r.PausedUntil)
		assert.True(t, r.AcceptingOrdersAt(now))
	})
}

func ptrTime(t time.Time) *time.Time { return &t }
