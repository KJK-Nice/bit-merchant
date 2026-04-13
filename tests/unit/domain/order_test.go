package domain_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOrder(t *testing.T) {
	item, _ := order.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 2, 10.0)
	items := []order.OrderItem{*item}

	t.Run("should create valid order", func(t *testing.T) {
		gotOrder, err := order.NewOrder(
			"o_1",
			"101",
			"rest_1",
			"session_1",
			items,
			2000, // 20.00 in cents/satoshis abstract unit
			common.PaymentMethodTypeCash,
		)

		assert.NoError(t, err)
		assert.NotNil(t, gotOrder)
		assert.Equal(t, common.PaymentStatusPending, gotOrder.PaymentStatus)
		assert.Equal(t, common.FulfillmentStatusPaid, gotOrder.FulfillmentStatus) // Using default from constructor
		assert.Equal(t, common.PaymentMethodTypeCash, gotOrder.PaymentMethod)
	})

	t.Run("should fail with no items", func(t *testing.T) {
		_, err := order.NewOrder("o_1", "101", "rest_1", "session_1", []order.OrderItem{}, 100, common.PaymentMethodTypeCash)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid total", func(t *testing.T) {
		_, err := order.NewOrder("o_1", "101", "rest_1", "session_1", items, 0, common.PaymentMethodTypeCash)
		assert.Error(t, err)
	})
}

func TestNewOrderItem(t *testing.T) {
	t.Run("should create valid order item", func(t *testing.T) {
		item, err := order.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 2, 10.50)
		assert.NoError(t, err)
		assert.Equal(t, 21.0, item.Subtotal)
		assert.Equal(t, "Burger", item.Name)
	})

	t.Run("should fail with invalid quantity", func(t *testing.T) {
		_, err := order.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 0, 10.0)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid price", func(t *testing.T) {
		_, err := order.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 1, 0)
		assert.Error(t, err)
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := order.NewOrderItem("oi_1", "o_1", "mi_1", "", 1, 10.0)
		assert.Error(t, err)
	})
}

func TestOrder_UpdateFulfillmentStatus(t *testing.T) {
	testOrder, _ := order.NewOrder("o_1", "101", "r_1", "session_1", []order.OrderItem{{}}, 100, common.PaymentMethodTypeCash)
	// Initial status is FulfillmentStatusPaid (from NewOrder)

	t.Run("valid transitions", func(t *testing.T) {
		// Paid -> Preparing
		err := testOrder.UpdateFulfillmentStatus(common.FulfillmentStatusPreparing)
		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusPreparing, testOrder.FulfillmentStatus)

		// Preparing -> Ready
		err = testOrder.UpdateFulfillmentStatus(common.FulfillmentStatusReady)
		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusReady, testOrder.FulfillmentStatus)

		// Ready -> Completed
		err = testOrder.UpdateFulfillmentStatus(common.FulfillmentStatusCompleted)
		assert.NoError(t, err)
		assert.Equal(t, common.FulfillmentStatusCompleted, testOrder.FulfillmentStatus)
		assert.NotNil(t, testOrder.CompletedAt)
	})

	t.Run("invalid transition", func(t *testing.T) {
		// Reset order to Paid
		testOrder, _ := order.NewOrder("o_1", "101", "r_1", "session_1", []order.OrderItem{{}}, 100, common.PaymentMethodTypeCash)

		// Paid -> Ready (skipping Preparing)
		err := testOrder.UpdateFulfillmentStatus(common.FulfillmentStatusReady)
		assert.Error(t, err)
	})
}
