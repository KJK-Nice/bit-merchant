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

func TestGetCustomerOrderByNumberUseCase(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	item, _ := order.NewOrderItem("oi", "o1", "mi", "Burger", 1, 10)
	o, _ := order.NewOrder("o1", "5555", "restaurant_1", "sess-z", []order.OrderItem{*item}, 1000, common.PaymentMethodTypeCash)
	require.NoError(t, repo.Save(o))

	uc := orderQuery.NewGetCustomerOrderByNumberUseCase(repo)
	got, err := uc.Execute(context.Background(), "sess-z", "5555")
	require.NoError(t, err)
	assert.Equal(t, common.OrderNumber("5555"), got.OrderNumber)

	_, err = uc.Execute(context.Background(), "other-session", "5555")
	assert.Error(t, err)

	_, err = uc.Execute(context.Background(), "sess-z", "9999")
	assert.Error(t, err)
}
