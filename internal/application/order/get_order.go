package order

import (
	"context"

	"bitmerchant/internal/domain"
)

// GetOrderByNumberUseCase retrieves order by number
type GetOrderByNumberUseCase struct {
	orderRepo domain.OrderRepository
}

// NewGetOrderByNumberUseCase creates a new GetOrderByNumberUseCase
func NewGetOrderByNumberUseCase(orderRepo domain.OrderRepository) *GetOrderByNumberUseCase {
	return &GetOrderByNumberUseCase{
		orderRepo: orderRepo,
	}
}

// Execute retrieves order by number
func (uc *GetOrderByNumberUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error) {
	return uc.orderRepo.FindByOrderNumber(restaurantID, orderNumber)
}
