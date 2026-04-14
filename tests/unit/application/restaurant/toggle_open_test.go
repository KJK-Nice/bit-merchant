package restaurant_test

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestToggleRestaurantOpenUseCase(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	uc := restaurantCmd.NewToggleRestaurantOpenUseCase(repo)

	r, _ := restaurant.NewRestaurant("r1", "Test Cafe")
	r.IsOpen = true
	_ = repo.Save(r)

	t.Run("Toggle Open to Closed", func(t *testing.T) {
		newState, err := uc.Execute(context.Background(), "r1", "", "")
		assert.NoError(t, err)
		assert.False(t, newState)

		updated, _ := repo.FindByID("r1")
		assert.False(t, updated.IsOpen)
	})

	t.Run("Toggle Closed to Open", func(t *testing.T) {
		newState, err := uc.Execute(context.Background(), "r1", "", "")
		assert.NoError(t, err)
		assert.True(t, newState)

		updated, _ := repo.FindByID("r1")
		assert.True(t, updated.IsOpen)
	})
}

func TestToggleRestaurantOpenUseCase_ClosedMessageAndClearOnReopen(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	uc := restaurantCmd.NewToggleRestaurantOpenUseCase(repo)
	ctx := context.Background()

	r, _ := restaurant.NewRestaurant("r2", "Bistro")
	require.NoError(t, repo.Save(r))

	_, err := uc.Execute(ctx, "r2", "Private party tonight", "Tomorrow 11:00")
	require.NoError(t, err)
	closed, _ := repo.FindByID("r2")
	assert.False(t, closed.IsOpen)
	assert.Equal(t, "Private party tonight", closed.ClosedMessage)
	assert.Equal(t, "Tomorrow 11:00", closed.ReopeningHours)

	_, err = uc.Execute(ctx, "r2", "", "")
	require.NoError(t, err)
	again, _ := repo.FindByID("r2")
	assert.True(t, again.IsOpen)
	assert.Empty(t, again.ClosedMessage)
	assert.Empty(t, again.ReopeningHours)
}

func TestToggleRestaurantOpenUseCase_LongMessageRejected(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	uc := restaurantCmd.NewToggleRestaurantOpenUseCase(repo)
	r, _ := restaurant.NewRestaurant("r3", "X")
	require.NoError(t, repo.Save(r))

	long := strings.Repeat("a", 501)
	_, err := uc.Execute(context.Background(), "r3", long, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed message")
}
