package domain_test

import (
	"testing"
	"time"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewRestaurant(t *testing.T) {
	t.Run("should create valid restaurant", func(t *testing.T) {
		id := domain.RestaurantID("rest_123")
		name := "Tasty Burger"

		restaurant, err := domain.NewRestaurant(id, name)

		assert.NoError(t, err)
		assert.NotNil(t, restaurant)
		assert.Equal(t, id, restaurant.ID)
		assert.Equal(t, name, restaurant.Name)
		assert.True(t, restaurant.IsOpen)
		assert.WithinDuration(t, time.Now(), restaurant.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), restaurant.UpdatedAt, time.Second)
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := domain.NewRestaurant("id", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restaurant name must be between 1 and 100 characters")
	})

	t.Run("should fail with too long name", func(t *testing.T) {
		longName := ""
		for i := 0; i < 101; i++ {
			longName += "a"
		}
		_, err := domain.NewRestaurant("id", longName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restaurant name must be between 1 and 100 characters")
	})
}

func TestRestaurant_UpdateStatus(t *testing.T) {
	r, _ := domain.NewRestaurant("id", "name")
	originalUpdate := r.UpdatedAt

	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)

	r.UpdateStatus(false, "Closed for holiday", "Tomorrow 9am")

	assert.False(t, r.IsOpen)
	assert.Equal(t, "Closed for holiday", r.ClosedMessage)
	assert.Equal(t, "Tomorrow 9am", r.ReopeningHours)
	assert.True(t, r.UpdatedAt.After(originalUpdate))
}

