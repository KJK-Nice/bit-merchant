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

	var paidOrders []*domain.Order
	for _, order := range orders {
		if order.PaymentStatus != domain.PaymentStatusPaid {
			continue
		}
		paidOrders = append(paidOrders, order)
	}

	// Sort by CreatedAt desc
	sort.Slice(paidOrders, func(i, j int) bool {
		return paidOrders[i].CreatedAt.After(paidOrders[j].CreatedAt)
	})

	return paidOrders, nil
}
