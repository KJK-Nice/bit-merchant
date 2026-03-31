package query

import (
	"context"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type GetOrderByNumberUseCase struct {
	orderRepo order.Repository
}

func NewGetOrderByNumberUseCase(orderRepo order.Repository) *GetOrderByNumberUseCase {
	return &GetOrderByNumberUseCase{orderRepo: orderRepo}
}

func (uc *GetOrderByNumberUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, orderNumber string) (*order.Order, error) {
	return uc.orderRepo.FindByOrderNumber(restaurantID, orderNumber)
}
