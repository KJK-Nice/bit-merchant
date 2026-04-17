package restaurant_test

import (
	"context"
	"testing"

	"bitmerchant/internal/infrastructure/repositories/memory"
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubQR struct{}

func (stubQR) GeneratePNG(url string, size int) ([]byte, error) {
	return []byte(url), nil
}

func TestRestaurantTableQRImageHandler_MenuURLForTable(t *testing.T) {
	u := restaurantQuery.MenuURLForTable("http://localhost:8080", "rest_1", 5)
	assert.Contains(t, u, "restaurantID=rest_1")
	assert.Contains(t, u, "table=5")
	assert.Contains(t, u, "/menu")
}

func TestRestaurantTableQRImageHandler_Handle(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	r, err := restaurant.NewRestaurant("r1", "Cafe")
	require.NoError(t, err)
	r.TableCount = 3
	require.NoError(t, repo.Save(r))

	h := restaurantQuery.NewRestaurantTableQRImageHandler(stubQR{}, "http://host", repo, nil, nil)

	t.Run("table in range", func(t *testing.T) {
		b, err := h.Handle(context.Background(), restaurantQuery.RestaurantTableQRImage{
			RestaurantID: "r1",
			TableNumber:  2,
		})
		require.NoError(t, err)
		assert.Equal(t, "http://host/menu?restaurantID=r1&table=2", string(b))
	})

	t.Run("coerces low table count when loading", func(t *testing.T) {
		r2, _ := restaurant.NewRestaurant("r2", "Low")
		r2.TableCount = 0
		require.NoError(t, repo.Save(r2))
		b, err := h.Handle(context.Background(), restaurantQuery.RestaurantTableQRImage{
			RestaurantID: "r2",
			TableNumber:  1,
		})
		require.NoError(t, err)
		assert.Contains(t, string(b), "table=1")
	})

	t.Run("table above configured count", func(t *testing.T) {
		_, err := h.Handle(context.Background(), restaurantQuery.RestaurantTableQRImage{
			RestaurantID: "r1",
			TableNumber:  10,
		})
		assert.Error(t, err)
	})
}
