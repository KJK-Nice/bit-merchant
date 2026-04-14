package domain_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
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
