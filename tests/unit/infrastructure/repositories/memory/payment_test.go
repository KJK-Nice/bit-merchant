package memory_test

import (
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestMemoryPaymentRepository(t *testing.T) {
	repo := memory.NewMemoryPaymentRepository()

	t.Run("Save and FindByID", func(t *testing.T) {
		p, _ := domain.NewPayment("p1", "o1", "r1", domain.PaymentMethodTypeCash, 10.0)
		err := repo.Save(p)
		assert.NoError(t, err)

		found, err := repo.FindByID("p1")
		assert.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)
	})

	t.Run("FindByOrderID", func(t *testing.T) {
		p, _ := domain.NewPayment("p2", "o2", "r1", domain.PaymentMethodTypeCash, 10.0)
		repo.Save(p)

		found, err := repo.FindByOrderID("o2")
		assert.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)

		_, err = repo.FindByOrderID("non_existent")
		assert.Error(t, err)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		p1, _ := domain.NewPayment("p3", "o3", "r2", domain.PaymentMethodTypeCash, 10.0)
		p2, _ := domain.NewPayment("p4", "o4", "r2", domain.PaymentMethodTypeCash, 10.0)
		p3, _ := domain.NewPayment("p5", "o5", "r3", domain.PaymentMethodTypeCash, 10.0)

		repo.Save(p1)
		repo.Save(p2)
		repo.Save(p3)

		payments, err := repo.FindByRestaurantID("r2")
		assert.NoError(t, err)
		assert.Len(t, payments, 2)
	})
}

