package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
	"log/slog"
)

// PaidOrdersForRestaurant lists paid orders for a restaurant (newest first).
type PaidOrdersForRestaurant struct {
	RestaurantID common.RestaurantID
}

type PaidOrdersForRestaurantHandler decorator.QueryHandler[PaidOrdersForRestaurant, []*order.Order]

type paidOrdersForRestaurantHandler struct {
	orders OrderReadModel
}

func NewPaidOrdersForRestaurantHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) PaidOrdersForRestaurantHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := paidOrdersForRestaurantHandler{orders: orders}
	return decorator.ApplyQueryDecorators[PaidOrdersForRestaurant, []*order.Order](h, log, metrics)
}

func (h paidOrdersForRestaurantHandler) Handle(ctx context.Context, q PaidOrdersForRestaurant) ([]*order.Order, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
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
