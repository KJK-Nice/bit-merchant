package cash_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"

	"bitmerchant/internal/infrastructure/payment/cash"
	"context"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCashPaymentMethod(t *testing.T) {
	method := cash.NewCashPaymentMethod()

	t.Run("GetPaymentMethodType", func(t *testing.T) {
		assert.Equal(t, common.PaymentMethodTypeCash, method.GetPaymentMethodType())
	})

	t.Run("ValidatePayment", func(t *testing.T) {
		err := method.ValidatePayment(context.Background(), "o1")
		assert.NoError(t, err)
	})

	t.Run("ProcessPayment", func(t *testing.T) {
		payment, err := method.ProcessPayment(context.Background(), "o1", "r1", money.FromMajor(10.0, money.USD))

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, common.PaymentMethodTypeCash, payment.Method)
		assert.Equal(t, common.PaymentStatusPending, payment.Status)
		assert.Equal(t, 10.0, payment.Amount)
		assert.Equal(t, money.USD, payment.Currency)
	})

	t.Run("ProcessPayment SAT", func(t *testing.T) {
		payment, err := method.ProcessPayment(context.Background(), "o2", "r1", money.New(5000, money.SAT))

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, money.SAT, payment.Currency)
		assert.Equal(t, money.New(5000, money.SAT), payment.Money())
	})
}
