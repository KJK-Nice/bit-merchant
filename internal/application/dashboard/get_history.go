package dashboard

import (
	"context"
	"sort"

	"bitmerchant/internal/domain"
)

type GetOrderHistoryUseCase struct {
	orderRepo domain.OrderRepository
}

func NewGetOrderHistoryUseCase(orderRepo domain.OrderRepository) *GetOrderHistoryUseCase {
	return &GetOrderHistoryUseCase{
		orderRepo: orderRepo,
	}
}

func (uc *GetOrderHistoryUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	// Sort by CreatedAt desc
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	return orders, nil
}

