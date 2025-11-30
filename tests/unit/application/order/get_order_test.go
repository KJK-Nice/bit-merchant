package order_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
)

func TestGetOrderByNumberUseCase(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	uc := order.NewGetOrderByNumberUseCase(repo)

	t.Run("Execute Success", func(t *testing.T) {
		// Setup order
		item, _ := domain.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		existingOrder, _ := domain.NewOrder(
			"o1", 
			"1234", 
			"r1",
			"session_1",
			[]domain.OrderItem{*item}, 
			1000, 
			domain.PaymentMethodTypeCash,
		)
		repo.Save(existingOrder)

		result, err := uc.Execute(context.Background(), "r1", "1234")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, existingOrder.ID, result.ID)
	})

	t.Run("Execute Not Found", func(t *testing.T) {
		_, err := uc.Execute(context.Background(), "r1", "9999")
		assert.Error(t, err)
	})
}

