package restaurant_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestToggleRestaurantOpenUseCase(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	uc := restaurant.NewToggleRestaurantOpenUseCase(repo)
	
	// Setup
	r, _ := domain.NewRestaurant("r1", "Test Cafe")
	// Default open? domain logic says NewRestaurant creates it ... usually open or close.
	// Let's assume default is Open=false or Open=true. Let's check domain/restaurant.go later.
	// For now, set explicitly.
	r.IsOpen = true
	_ = repo.Save(r)

	t.Run("Toggle Open to Closed", func(t *testing.T) {
		newState, err := uc.Execute(context.Background(), "r1")
		assert.NoError(t, err)
		assert.False(t, newState)
		
		updated, _ := repo.FindByID("r1")
		assert.False(t, updated.IsOpen)
	})

	t.Run("Toggle Closed to Open", func(t *testing.T) {
		newState, err := uc.Execute(context.Background(), "r1")
		assert.NoError(t, err)
		assert.True(t, newState)
		
		updated, _ := repo.FindByID("r1")
		assert.True(t, updated.IsOpen)
	})
}

