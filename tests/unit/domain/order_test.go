package domain_test

import (
	"testing"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder(t *testing.T) {
	item, _ := domain.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 2, 10.0)
	items := []domain.OrderItem{*item}

	t.Run("should create valid order", func(t *testing.T) {
		order, err := domain.NewOrder(
			"o_1",
			"101",
			"rest_1",
			"session_1",
			items,
			2000, // 20.00 in cents/satoshis abstract unit
			domain.PaymentMethodTypeCash,
		)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, domain.PaymentStatusPending, order.PaymentStatus)
		assert.Equal(t, domain.FulfillmentStatusPaid, order.FulfillmentStatus) // Using default from constructor
		assert.Equal(t, domain.PaymentMethodTypeCash, order.PaymentMethod)
	})

	t.Run("should fail with no items", func(t *testing.T) {
		_, err := domain.NewOrder("o_1", "101", "rest_1", "session_1", []domain.OrderItem{}, 100, domain.PaymentMethodTypeCash)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid total", func(t *testing.T) {
		_, err := domain.NewOrder("o_1", "101", "rest_1", "session_1", items, 0, domain.PaymentMethodTypeCash)
		assert.Error(t, err)
	})
}

func TestNewOrderItem(t *testing.T) {
	t.Run("should create valid order item", func(t *testing.T) {
		item, err := domain.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 2, 10.50)
		assert.NoError(t, err)
		assert.Equal(t, 21.0, item.Subtotal)
		assert.Equal(t, "Burger", item.Name)
	})

	t.Run("should fail with invalid quantity", func(t *testing.T) {
		_, err := domain.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 0, 10.0)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid price", func(t *testing.T) {
		_, err := domain.NewOrderItem("oi_1", "o_1", "mi_1", "Burger", 1, 0)
		assert.Error(t, err)
	})
	
	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := domain.NewOrderItem("oi_1", "o_1", "mi_1", "", 1, 10.0)
		assert.Error(t, err)
	})
}

func TestOrder_UpdateFulfillmentStatus(t *testing.T) {
	order, _ := domain.NewOrder("o_1", "101", "r_1", "session_1", []domain.OrderItem{{}}, 100, domain.PaymentMethodTypeCash)
	// Initial status is FulfillmentStatusPaid (from NewOrder)

	t.Run("valid transitions", func(t *testing.T) {
		// Paid -> Preparing
		err := order.UpdateFulfillmentStatus(domain.FulfillmentStatusPreparing)
		assert.NoError(t, err)
		assert.Equal(t, domain.FulfillmentStatusPreparing, order.FulfillmentStatus)

		// Preparing -> Ready
		err = order.UpdateFulfillmentStatus(domain.FulfillmentStatusReady)
		assert.NoError(t, err)
		assert.Equal(t, domain.FulfillmentStatusReady, order.FulfillmentStatus)

		// Ready -> Completed
		err = order.UpdateFulfillmentStatus(domain.FulfillmentStatusCompleted)
		assert.NoError(t, err)
		assert.Equal(t, domain.FulfillmentStatusCompleted, order.FulfillmentStatus)
		assert.NotNil(t, order.CompletedAt)
	})

	t.Run("invalid transition", func(t *testing.T) {
		// Reset order to Paid
		order, _ := domain.NewOrder("o_1", "101", "r_1", "session_1", []domain.OrderItem{{}}, 100, domain.PaymentMethodTypeCash)
		
		// Paid -> Ready (skipping Preparing)
		err := order.UpdateFulfillmentStatus(domain.FulfillmentStatusReady)
		assert.Error(t, err)
	})
}
