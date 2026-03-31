package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type GetOrderHistoryUseCase struct {
	orderRepo order.Repository
}

func NewGetOrderHistoryUseCase(orderRepo order.Repository) *GetOrderHistoryUseCase {
	return &GetOrderHistoryUseCase{orderRepo: orderRepo}
}

func (uc *GetOrderHistoryUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID) ([]*order.Order, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}
	var paidOrders []*order.Order
	for _, o := range orders {
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		paidOrders = append(paidOrders, o)
	}
	sort.Slice(paidOrders, func(i, j int) bool {
		return paidOrders[i].CreatedAt.After(paidOrders[j].CreatedAt)
	})
	return paidOrders, nil
}
