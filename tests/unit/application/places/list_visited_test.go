package places_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/application/places"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListVisitedRestaurantsUseCase(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	ord := memory.NewMemoryOrderRepository()

	r1, _ := domain.NewRestaurant("r1", "First")
	r2, _ := domain.NewRestaurant("r2", "Second")
	require.NoError(t, rest.Save(r1))
	require.NoError(t, rest.Save(r2))

	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	require.NoError(t, visits.Upsert(&domain.SessionRestaurantVisit{
		SessionID: "s1", RestaurantID: "r1", FirstVisitedAt: older, LastVisitedAt: older,
	}))
	require.NoError(t, visits.Upsert(&domain.SessionRestaurantVisit{
		SessionID: "s1", RestaurantID: "r2", FirstVisitedAt: newer, LastVisitedAt: newer,
	}))

	item, _ := domain.NewOrderItem("oi", "o1", "mi", "X", 1, 1)
	o, _ := domain.NewOrder("o1", "1001", "r1", "s1", []domain.OrderItem{*item}, 100, domain.PaymentMethodTypeCash)
	o.FiatAmount = 1
	require.NoError(t, ord.Save(o))

	uc := places.NewListVisitedRestaurantsUseCase(visits, rest, ord)
	out, err := uc.Execute(context.Background(), "s1")
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, domain.RestaurantID("r2"), out[0].RestaurantID)
	assert.Equal(t, domain.RestaurantID("r1"), out[1].RestaurantID)
	var ordered *places.VisitedPlace
	for i := range out {
		if out[i].RestaurantID == "r1" {
			ordered = &out[i]
			break
		}
	}
	require.NotNil(t, ordered)
	assert.True(t, ordered.HasOrderedHere)
}
