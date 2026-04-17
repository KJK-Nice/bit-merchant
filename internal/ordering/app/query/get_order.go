package query

import (
	"context"
	"log/slog"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
)

// OrderByNumberForRestaurant loads an order by restaurant and order number string.
type OrderByNumberForRestaurant struct {
	RestaurantID common.RestaurantID
	OrderNumber  string
}

type OrderByNumberForRestaurantHandler decorator.QueryHandler[OrderByNumberForRestaurant, *order.Order]

type orderByNumberForRestaurantHandler struct {
	orderRepo order.Repository
}

func NewOrderByNumberForRestaurantHandler(orderRepo order.Repository, log *slog.Logger, metrics decorator.MetricsClient) OrderByNumberForRestaurantHandler {
	h := orderByNumberForRestaurantHandler{orderRepo: orderRepo}
	return decorator.ApplyQueryDecorators[OrderByNumberForRestaurant, *order.Order](h, log, metrics)
}

func (h orderByNumberForRestaurantHandler) Handle(ctx context.Context, q OrderByNumberForRestaurant) (*order.Order, error) {
	_ = ctx
	return h.orderRepo.FindByOrderNumber(q.RestaurantID, q.OrderNumber)
}
