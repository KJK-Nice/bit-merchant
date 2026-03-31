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
		err := method.ValidatePayment(context.Background(), "o1")
		assert.NoError(t, err)
	})

	t.Run("ProcessPayment", func(t *testing.T) {
		payment, err := method.ProcessPayment(context.Background(), "o1", "r1", 10.0)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, domain.PaymentMethodTypeCash, payment.Method)
		assert.Equal(t, domain.PaymentStatusPending, payment.Status)
		assert.Equal(t, 10.0, payment.Amount)
	})
}
