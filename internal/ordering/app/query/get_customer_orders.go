package query

import (
	"context"
	"fmt"
	"sort"

	"bitmerchant/internal/ordering/domain/order"
)

type GetCustomerOrdersUseCase struct {
	orderRepo order.Repository
}

func NewGetCustomerOrdersUseCase(orderRepo order.Repository) *GetCustomerOrdersUseCase {
	return &GetCustomerOrdersUseCase{orderRepo: orderRepo}
}

func (uc *GetCustomerOrdersUseCase) Execute(ctx context.Context, sessionID string) ([]*order.Order, error) {
	orders, err := uc.orderRepo.FindBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

type GetCustomerOrderByNumberUseCase struct {
	orderRepo order.Repository
}

func NewGetCustomerOrderByNumberUseCase(orderRepo order.Repository) *GetCustomerOrderByNumberUseCase {
	return &GetCustomerOrderByNumberUseCase{orderRepo: orderRepo}
}

func (uc *GetCustomerOrderByNumberUseCase) Execute(ctx context.Context, sessionID, orderNumber string) (*order.Order, error) {
	if sessionID == "" || orderNumber == "" {
		return nil, fmt.Errorf("order not found")
	}
	orders, err := uc.orderRepo.FindBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		if string(o.OrderNumber) == orderNumber {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}
