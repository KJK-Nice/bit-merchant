package query

import (
	"context"
	"log/slog"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
)

// ActiveKitchenOrders lists in-progress orders for a restaurant kitchen view.
type ActiveKitchenOrders struct {
	RestaurantID common.RestaurantID
}

type ActiveKitchenOrdersHandler decorator.QueryHandler[ActiveKitchenOrders, []*order.Order]

type activeKitchenOrdersHandler struct {
	repo order.Repository
}

func NewActiveKitchenOrdersHandler(repo order.Repository, log *slog.Logger, metrics decorator.MetricsClient) ActiveKitchenOrdersHandler {
	h := activeKitchenOrdersHandler{repo: repo}
	return decorator.ApplyQueryDecorators[ActiveKitchenOrders, []*order.Order](h, log, metrics)
}

func (h activeKitchenOrdersHandler) Handle(ctx context.Context, q ActiveKitchenOrders) ([]*order.Order, error) {
	_ = ctx
	orders, err := h.repo.FindActiveByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})
	return orders, nil
}
