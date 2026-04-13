package order_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetOrderByNumberUseCase(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	uc := orderQuery.NewGetOrderByNumberUseCase(repo)

	t.Run("Execute Success", func(t *testing.T) {
		// Setup order
		item, _ := order.NewOrderItem("oi1", "o1", "mi1", "Burger", 1, 10.0)
		existingOrder, _ := order.NewOrder(
			"o1",
			"1234",
			"r1",
			"session_1",
			[]order.OrderItem{*item},
			1000,
			common.PaymentMethodTypeCash,
		)
		require.NoError(t, repo.Save(existingOrder))

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
