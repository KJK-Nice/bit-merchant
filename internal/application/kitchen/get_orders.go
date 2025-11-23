package kitchen

import (
	"bitmerchant/internal/domain"
	"context"
	"sort"
)

type GetKitchenOrdersUseCase struct {
	repo domain.OrderRepository
}

func NewGetKitchenOrdersUseCase(repo domain.OrderRepository) *GetKitchenOrdersUseCase {
	return &GetKitchenOrdersUseCase{
		repo: repo,
	}
}

func (uc *GetKitchenOrdersUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	orders, err := uc.repo.FindActiveByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}
	
	// Sort by CreatedAt ascending (oldest first)
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	return orders, nil
}
