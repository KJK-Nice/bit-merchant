package dashboard_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/common"
	dashboard "bitmerchant/internal/dashboard/app/query"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStalledOrdersHandler_Integration(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	handler := dashboard.NewStalledOrdersHandler(orderRepo, nil, nil)

	items := []order.OrderItem{{MenuItemID: "i1", Name: "X", Quantity: 1, UnitPrice: 1.0, Subtotal: 1.0}}

	stale, err := order.NewOrder("stale", "S001", "restaurant_1", "sess_a", items, 100, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	stale.FulfillmentStatus = common.FulfillmentStatusPreparing
	stale.CreatedAt = time.Now().Add(-20 * time.Minute)
	require.NoError(t, orderRepo.Save(stale))

	fresh, err := order.NewOrder("fresh", "F001", "restaurant_1", "sess_b", items, 100, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	fresh.FulfillmentStatus = common.FulfillmentStatusPreparing
	fresh.CreatedAt = time.Now().Add(-2 * time.Minute)
	require.NoError(t, orderRepo.Save(fresh))

	view, err := handler.Handle(context.Background(), dashboard.StalledOrders{RestaurantID: "restaurant_1"})
	require.NoError(t, err)
	assert.Equal(t, 1, view.Count)
	require.NotNil(t, view.Sample)
	assert.Equal(t, common.OrderID("stale"), view.Sample.ID)
	assert.GreaterOrEqual(t, view.SampleAgeMinutes(), 20)
}

func TestStalledOrdersHandler_NoneActive(t *testing.T) {
	orderRepo := memory.NewMemoryOrderRepository()
	handler := dashboard.NewStalledOrdersHandler(orderRepo, nil, nil)

	view, err := handler.Handle(context.Background(), dashboard.StalledOrders{RestaurantID: "restaurant_1"})
	require.NoError(t, err)
	assert.Equal(t, 0, view.Count)
	assert.Nil(t, view.Sample)
}
