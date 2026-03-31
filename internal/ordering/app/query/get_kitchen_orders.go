package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type GetKitchenOrdersUseCase struct {
	repo order.Repository
}

func NewGetKitchenOrdersUseCase(repo order.Repository) *GetKitchenOrdersUseCase {
	return &GetKitchenOrdersUseCase{repo: repo}
}

func (uc *GetKitchenOrdersUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID) ([]*order.Order, error) {
	orders, err := uc.repo.FindActiveByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})
	return orders, nil
}
