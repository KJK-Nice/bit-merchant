package order

import (
	"context"
	"sort"

	"bitmerchant/internal/domain"
)

// GetCustomerOrdersUseCase retrieves orders for a customer session
type GetCustomerOrdersUseCase struct {
	orderRepo domain.OrderRepository
}

// NewGetCustomerOrdersUseCase creates a new GetCustomerOrdersUseCase
func NewGetCustomerOrdersUseCase(orderRepo domain.OrderRepository) *GetCustomerOrdersUseCase {
	return &GetCustomerOrdersUseCase{
		orderRepo: orderRepo,
	}
}

// Execute retrieves orders for a customer session
func (uc *GetCustomerOrdersUseCase) Execute(ctx context.Context, sessionID string) ([]*domain.Order, error) {
	orders, err := uc.orderRepo.FindBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	// Sort orders by creation time (newest first)
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	return orders, nil
}

