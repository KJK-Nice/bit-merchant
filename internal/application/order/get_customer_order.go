package order

import (
	"context"
	"fmt"

	"bitmerchant/internal/domain"
)

// GetCustomerOrderByNumberUseCase loads an order by number only if it belongs to the session.
type GetCustomerOrderByNumberUseCase struct {
	orderRepo domain.OrderRepository
}

// NewGetCustomerOrderByNumberUseCase constructs the use case.
func NewGetCustomerOrderByNumberUseCase(orderRepo domain.OrderRepository) *GetCustomerOrderByNumberUseCase {
	return &GetCustomerOrderByNumberUseCase{orderRepo: orderRepo}
}

// Execute finds an order with the given number in the session's history.
func (uc *GetCustomerOrderByNumberUseCase) Execute(ctx context.Context, sessionID, orderNumber string) (*domain.Order, error) {
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
