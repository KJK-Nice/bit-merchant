package cash_test

import (
	"context"
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/payment/cash"

	"github.com/stretchr/testify/assert"
)

func TestCashPaymentMethod(t *testing.T) {
	method := cash.NewCashPaymentMethod()

	t.Run("GetPaymentMethodType", func(t *testing.T) {
		assert.Equal(t, domain.PaymentMethodTypeCash, method.GetPaymentMethodType())
	})

	t.Run("ValidatePayment", func(t *testing.T) {
		err := method.ValidatePayment(context.Background(), nil)
		assert.NoError(t, err)
	})

	t.Run("ProcessPayment", func(t *testing.T) {
		item, _ := domain.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		order, _ := domain.NewOrder("o1", "101", "r1", []domain.OrderItem{*item}, 1000, domain.PaymentMethodTypeCash)
		order.FiatAmount = 10.0

		payment, err := method.ProcessPayment(context.Background(), order)
		
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, domain.PaymentMethodTypeCash, payment.Method)
		assert.Equal(t, domain.PaymentStatusPending, payment.Status)
		assert.Equal(t, 10.0, payment.Amount)
		assert.Equal(t, order.ID, payment.OrderID)
	})
}

