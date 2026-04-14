package memory_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/payment/domain/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemoryPaymentRepository(t *testing.T) {
	repo := memory.NewMemoryPaymentRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		p, _ := payment.NewPayment("p1", "o1", "r1", common.PaymentMethodTypeCash, 10.0)
		err := repo.Save(p)
		assert.NoError(t, err)

		found, err := repo.FindByID("p1")
		assert.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)
	})

	t.Run("FindByOrderID", func(t *testing.T) {
		p, _ := payment.NewPayment("p2", "o2", "r1", common.PaymentMethodTypeCash, 10.0)
		require.NoError(t, repo.Save(p))

		found, err := repo.FindByOrderID("o2")
		assert.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)

		_, err = repo.FindByOrderID("non_existent")
		assert.Error(t, err)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		p1, _ := payment.NewPayment("p3", "o3", "r2", common.PaymentMethodTypeCash, 10.0)
		p2, _ := payment.NewPayment("p4", "o4", "r2", common.PaymentMethodTypeCash, 10.0)
		p3, _ := payment.NewPayment("p5", "o5", "r3", common.PaymentMethodTypeCash, 10.0)

		require.NoError(t, repo.Save(p1))
		require.NoError(t, repo.Save(p2))
		require.NoError(t, repo.Save(p3))

		payments, err := repo.FindByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, payments, 2)
	})
}
