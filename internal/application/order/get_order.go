package order

import (
	"errors"

	"bitmerchant/internal/domain"
)

// GetOrderByNumberUseCase retrieves order by order number
type GetOrderByNumberUseCase struct {
	orderRepo domain.OrderRepository
}

// NewGetOrderByNumberUseCase creates a new GetOrderByNumberUseCase
func NewGetOrderByNumberUseCase(orderRepo domain.OrderRepository) *GetOrderByNumberUseCase {
	return &GetOrderByNumberUseCase{
		orderRepo: orderRepo,
	}
}

// Execute retrieves order by order number
func (uc *GetOrderByNumberUseCase) Execute(restaurantID domain.RestaurantID, orderNumber domain.OrderNumber) (*domain.Order, error) {
	order, err := uc.orderRepo.FindByOrderNumber(restaurantID, string(orderNumber))
	if err != nil {
		return nil, errors.New("order not found")
	}

	return order, nil
}
