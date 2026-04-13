package domain_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/payment/domain/payment"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPayment(t *testing.T) {
	t.Run("should create valid payment", func(t *testing.T) {
		p, err := payment.NewPayment(
			"pay_1",
			"order_1",
			"rest_1",
			common.PaymentMethodTypeCash,
			25.50,
		)

		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, common.PaymentStatusPending, p.Status)
		assert.Equal(t, 25.50, p.Amount)
	})

	t.Run("should fail with invalid amount", func(t *testing.T) {
		_, err := payment.NewPayment("id", "oid", "rid", "cash", 0)
		assert.Error(t, err)
	})
}

func TestPayment_StateTransitions(t *testing.T) {
	t.Run("MarkAsPaid", func(t *testing.T) {
		p, _ := payment.NewPayment("id", "oid", "rid", "cash", 10)
		p.MarkAsPaid()
		assert.Equal(t, common.PaymentStatusPaid, p.Status)
		assert.NotNil(t, p.PaidAt)
	})

	t.Run("MarkPaid with OrderID", func(t *testing.T) {
		p, _ := payment.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkPaid("new_order_id")
		assert.NoError(t, err)
		assert.Equal(t, common.PaymentStatusPaid, p.Status)
		assert.Equal(t, common.OrderID("new_order_id"), p.OrderID)
	})

	t.Run("MarkFailed", func(t *testing.T) {
		p, _ := payment.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkFailed("insufficient funds")
		assert.NoError(t, err)
		assert.Equal(t, common.PaymentStatusFailed, p.Status)
		assert.Equal(t, "insufficient funds", p.FailureReason)
	})

	t.Run("MarkExpired", func(t *testing.T) {
		p, _ := payment.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkExpired()
		assert.NoError(t, err)
		assert.Equal(t, common.PaymentStatusExpired, p.Status)
	})

	t.Run("Invalid Transitions", func(t *testing.T) {
		p, _ := payment.NewPayment("id", "oid", "rid", "cash", 10)
		p.MarkAsPaid()

		// Cannot mark failed if already paid
		err := p.MarkFailed("reason")
		assert.Error(t, err)

		// Cannot mark expired if already paid
		err = p.MarkExpired()
		assert.Error(t, err)
	})
}
