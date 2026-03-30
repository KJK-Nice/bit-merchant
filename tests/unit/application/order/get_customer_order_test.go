package order_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCustomerOrderByNumberUseCase(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	item, _ := domain.NewOrderItem("oi", "o1", "mi", "Burger", 1, 10)
	o, _ := domain.NewOrder("o1", "5555", "restaurant_1", "sess-z", []domain.OrderItem{*item}, 1000, domain.PaymentMethodTypeCash)
	require.NoError(t, repo.Save(o))

	uc := order.NewGetCustomerOrderByNumberUseCase(repo)
	got, err := uc.Execute(context.Background(), "sess-z", "5555")
	require.NoError(t, err)
	assert.Equal(t, domain.OrderNumber("5555"), got.OrderNumber)

	_, err = uc.Execute(context.Background(), "other-session", "5555")
	assert.Error(t, err)

	_, err = uc.Execute(context.Background(), "sess-z", "9999")
	assert.Error(t, err)
}
