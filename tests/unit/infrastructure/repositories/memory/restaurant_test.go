package memory_test

import (
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRestaurantRepository(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		rest, _ := domain.NewRestaurant("r1", "Test Rest")
		err := repo.Save(rest)
		assert.NoError(t, err)

		found, err := repo.FindByID("r1")
		assert.NoError(t, err)
		assert.Equal(t, rest.ID, found.ID)
		assert.Equal(t, rest.Name, found.Name)
	})

	t.Run("FindByID Not Found", func(t *testing.T) {
		_, err := repo.FindByID("non_existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restaurant not found")
	})

	t.Run("Update", func(t *testing.T) {
		rest, _ := domain.NewRestaurant("r2", "Rest 2")
		repo.Save(rest)

		rest.UpdateStatus(false, "Closed", "Tomorrow")
		err := repo.Update(rest)
		assert.NoError(t, err)

		found, _ := repo.FindByID("r2")
		assert.False(t, found.IsOpen)
		assert.Equal(t, "Closed", found.ClosedMessage)
	})

	t.Run("Update Not Found", func(t *testing.T) {
		rest, _ := domain.NewRestaurant("r3", "Rest 3")
		err := repo.Update(rest)
		assert.Error(t, err)
	})
}

