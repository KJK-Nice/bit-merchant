package domain_test

import (
	"testing"
	"time"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewMenuCategory(t *testing.T) {
	t.Run("should create valid category", func(t *testing.T) {
		id := domain.CategoryID("cat_1")
		restID := domain.RestaurantID("rest_1")
		name := "Appetizers"
		order := 1

		cat, err := domain.NewMenuCategory(id, restID, name, order)

		assert.NoError(t, err)
		assert.NotNil(t, cat)
		assert.Equal(t, id, cat.ID)
		assert.Equal(t, restID, cat.RestaurantID)
		assert.Equal(t, name, cat.Name)
		assert.Equal(t, order, cat.DisplayOrder)
		assert.True(t, cat.IsActive)
		assert.WithinDuration(t, time.Now(), cat.CreatedAt, time.Second)
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := domain.NewMenuCategory("id", "rid", "", 1)
		assert.Error(t, err)
	})

	t.Run("should fail with negative order", func(t *testing.T) {
		_, err := domain.NewMenuCategory("id", "rid", "name", -1)
		assert.Error(t, err)
	})
}

func TestMenuCategory_SetActive(t *testing.T) {
	cat, _ := domain.NewMenuCategory("id", "rid", "name", 1)

	cat.SetActive(false)
	assert.False(t, cat.IsActive)

	cat.SetActive(true)
	assert.True(t, cat.IsActive)
}

func TestNewMenuItem(t *testing.T) {
	t.Run("should create valid item", func(t *testing.T) {
		id := domain.ItemID("item_1")
		catID := domain.CategoryID("cat_1")
		restID := domain.RestaurantID("rest_1")
		name := "Burger"
		price := 12.50

		item, err := domain.NewMenuItem(id, catID, restID, name, price)

		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, id, item.ID)
		assert.Equal(t, name, item.Name)
		assert.Equal(t, price, item.Price)
		assert.True(t, item.IsAvailable)
	})

	t.Run("should fail with invalid price", func(t *testing.T) {
		_, err := domain.NewMenuItem("id", "cid", "rid", "name", 0)
		assert.Error(t, err)
		_, err = domain.NewMenuItem("id", "cid", "rid", "name", -10)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		_, err := domain.NewMenuItem("id", "cid", "rid", "", 10)
		assert.Error(t, err)
	})
}

func TestMenuItem_Setters(t *testing.T) {
	item, _ := domain.NewMenuItem("id", "cid", "rid", "name", 10)

	// Description
	err := item.SetDescription("Tasty burger")
	assert.NoError(t, err)
	assert.Equal(t, "Tasty burger", item.Description)

	// Description too long
	longDesc := ""
	for i := 0; i < 501; i++ {
		longDesc += "a"
	}
	err = item.SetDescription(longDesc)
	assert.Error(t, err)

	// Photos
	item.SetPhotoURLs("http://thumb.jpg", "http://orig.jpg")
	assert.Equal(t, "http://thumb.jpg", item.PhotoURL)
	assert.Equal(t, "http://orig.jpg", item.PhotoOriginalURL)

	// Availability
	item.SetAvailable(false)
	assert.False(t, item.IsAvailable)
}
