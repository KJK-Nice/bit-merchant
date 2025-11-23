package domain_test

import (
	"testing"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewPayment(t *testing.T) {
	t.Run("should create valid payment", func(t *testing.T) {
		p, err := domain.NewPayment(
			"pay_1",
			"order_1",
			"rest_1",
			domain.PaymentMethodTypeCash,
			25.50,
		)

		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, domain.PaymentStatusPending, p.Status)
		assert.Equal(t, 25.50, p.Amount)
	})

	t.Run("should fail with invalid amount", func(t *testing.T) {
		_, err := domain.NewPayment("id", "oid", "rid", "cash", 0)
		assert.Error(t, err)
	})
}

func TestPayment_StateTransitions(t *testing.T) {
	t.Run("MarkAsPaid", func(t *testing.T) {
		p, _ := domain.NewPayment("id", "oid", "rid", "cash", 10)
		p.MarkAsPaid()
		assert.Equal(t, domain.PaymentStatusPaid, p.Status)
		assert.NotNil(t, p.PaidAt)
	})

	t.Run("MarkPaid with OrderID", func(t *testing.T) {
		p, _ := domain.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkPaid("new_order_id")
		assert.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusPaid, p.Status)
		assert.Equal(t, domain.OrderID("new_order_id"), p.OrderID)
	})

	t.Run("MarkFailed", func(t *testing.T) {
		p, _ := domain.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkFailed("insufficient funds")
		assert.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusFailed, p.Status)
		assert.Equal(t, "insufficient funds", p.FailureReason)
	})

	t.Run("MarkExpired", func(t *testing.T) {
		p, _ := domain.NewPayment("id", "oid", "rid", "cash", 10)
		err := p.MarkExpired()
		assert.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusExpired, p.Status)
	})

	t.Run("Invalid Transitions", func(t *testing.T) {
		p, _ := domain.NewPayment("id", "oid", "rid", "cash", 10)
		p.MarkAsPaid()

		// Cannot mark failed if already paid
		err := p.MarkFailed("reason")
		assert.Error(t, err)

		// Cannot mark expired if already paid
		err = p.MarkExpired()
		assert.Error(t, err)
	})
}
