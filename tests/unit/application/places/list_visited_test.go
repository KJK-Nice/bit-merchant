package places_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/domain/order"
	placesQuery "bitmerchant/internal/places/app/query"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionVisitedPlacesHandler(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	ord := memory.NewMemoryOrderRepository()

	r1, _ := restaurant.NewRestaurant("r1", "First")
	r1.UpdateStatus(false, "", "")
	r2, _ := restaurant.NewRestaurant("r2", "Second")
	require.NoError(t, rest.Save(r1))
	require.NoError(t, rest.Save(r2))

	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	require.NoError(t, visits.Upsert(context.Background(), visit.NewSessionRestaurantVisit("s1", "r1", older, older)))
	require.NoError(t, visits.Upsert(context.Background(), visit.NewSessionRestaurantVisit("s1", "r2", newer, newer)))

	item, _ := order.NewOrderItem("oi", "o1", "mi", "X", 1, 1)
	o, _ := order.NewOrder("o1", "1001", "r1", "s1", []order.OrderItem{*item}, 100, common.PaymentMethodTypeCash)
	o.FiatAmount = 1
	require.NoError(t, ord.Save(o))

	h := placesQuery.NewSessionVisitedPlacesHandler(visits, rest, ord, nil, nil)
	out, err := h.Handle(context.Background(), placesQuery.SessionVisitedPlaces{SessionID: "s1"})
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, common.RestaurantID("r2"), out[0].RestaurantID)
	assert.Equal(t, common.RestaurantID("r1"), out[1].RestaurantID)
	assert.True(t, out[0].IsOpen)
	assert.False(t, out[1].IsOpen)
	var ordered *placesQuery.VisitedPlace
	for i := range out {
		if out[i].RestaurantID == "r1" {
			ordered = &out[i]
			break
		}
	}
	require.NotNil(t, ordered)
	assert.True(t, ordered.HasOrderedHere)
}
