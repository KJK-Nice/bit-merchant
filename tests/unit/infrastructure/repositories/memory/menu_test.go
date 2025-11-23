package memory_test

import (
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestMemoryMenuCategoryRepository(t *testing.T) {
	repo := memory.NewMemoryMenuCategoryRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		cat, _ := domain.NewMenuCategory("c1", "r1", "Cat 1", 1)
		err := repo.Save(cat)
		assert.NoError(t, err)

		found, err := repo.FindByID("c1")
		assert.NoError(t, err)
		assert.Equal(t, cat.ID, found.ID)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		cat1, _ := domain.NewMenuCategory("c2", "r2", "Cat 1", 1)
		cat2, _ := domain.NewMenuCategory("c3", "r2", "Cat 2", 2)
		cat3, _ := domain.NewMenuCategory("c4", "r3", "Cat 3", 1)

		repo.Save(cat1)
		repo.Save(cat2)
		repo.Save(cat3)

		cats, err := repo.FindByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, cats, 2)
	})
}

func TestMemoryMenuItemRepository(t *testing.T) {
	repo := memory.NewMemoryMenuItemRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		item, _ := domain.NewMenuItem("i1", "c1", "r1", "Item 1", 10.0)
		err := repo.Save(item)
		assert.NoError(t, err)

		found, err := repo.FindByID("i1")
		assert.NoError(t, err)
		assert.Equal(t, item.ID, found.ID)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		item1, _ := domain.NewMenuItem("i2", "c1", "r2", "Item 1", 10.0)
		item2, _ := domain.NewMenuItem("i3", "c2", "r2", "Item 2", 12.0)
		item3, _ := domain.NewMenuItem("i4", "c3", "r3", "Item 3", 15.0)

		repo.Save(item1)
		repo.Save(item2)
		repo.Save(item3)

		items, err := repo.FindByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("FindAvailableByRestaurantID", func(t *testing.T) {
		item1, _ := domain.NewMenuItem("i5", "c1", "r4", "Item 1", 10.0)
		item2, _ := domain.NewMenuItem("i6", "c2", "r4", "Item 2", 12.0)
		item2.SetAvailable(false)

		repo.Save(item1)
		repo.Save(item2)

		items, err := repo.FindAvailableByRestaurantID("r4")
		assert.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "i5", string(items[0].ID))
	})
}

