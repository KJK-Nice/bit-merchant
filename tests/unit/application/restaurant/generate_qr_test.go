package restaurant_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubQR struct{}

func (stubQR) GeneratePNG(url string, size int) ([]byte, error) {
	return []byte(url), nil
}

func TestGenerateRestaurantQRUseCase_MenuURLForTable(t *testing.T) {
	u := restaurant.MenuURLForTable("http://localhost:8080", "rest_1", 5)
	assert.Contains(t, u, "restaurantID=rest_1")
	assert.Contains(t, u, "table=5")
	assert.Contains(t, u, "/menu")
}

func TestGenerateRestaurantQRUseCase_Execute(t *testing.T) {
	repo := memory.NewMemoryRestaurantRepository()
	r, err := domain.NewRestaurant("r1", "Cafe")
	require.NoError(t, err)
	r.TableCount = 3
	require.NoError(t, repo.Save(r))

	uc := restaurant.NewGenerateRestaurantQRUseCase(stubQR{}, "http://host", repo)

	t.Run("table in range", func(t *testing.T) {
		b, err := uc.Execute(context.Background(), "r1", 2)
		require.NoError(t, err)
		assert.Equal(t, "http://host/menu?restaurantID=r1&table=2", string(b))
	})

	t.Run("coerces low table count when loading", func(t *testing.T) {
		r2, _ := domain.NewRestaurant("r2", "Low")
		r2.TableCount = 0
		require.NoError(t, repo.Save(r2))
		b, err := uc.Execute(context.Background(), "r2", 1)
		require.NoError(t, err)
		assert.Contains(t, string(b), "table=1")
	})

	t.Run("table above configured count", func(t *testing.T) {
		_, err := uc.Execute(context.Background(), "r1", 10)
		assert.Error(t, err)
	})
}
