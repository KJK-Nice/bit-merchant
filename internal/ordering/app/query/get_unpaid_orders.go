package query

import (
	"context"
	"log/slog"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
)

// UnpaidServerOrders lists orders awaiting front-of-house payment confirmation
// for the server (FOH) view.
type UnpaidServerOrders struct {
	RestaurantID common.RestaurantID
}

type UnpaidServerOrdersHandler decorator.QueryHandler[UnpaidServerOrders, []*order.Order]

type unpaidServerOrdersHandler struct {
	repo order.Repository
}

func NewUnpaidServerOrdersHandler(repo order.Repository, log *slog.Logger, metrics decorator.MetricsClient) UnpaidServerOrdersHandler {
	h := unpaidServerOrdersHandler{repo: repo}
	return decorator.ApplyQueryDecorators[UnpaidServerOrders, []*order.Order](h, log, metrics)
}

func (h unpaidServerOrdersHandler) Handle(ctx context.Context, q UnpaidServerOrders) ([]*order.Order, error) {
	_ = ctx
	orders, err := h.repo.FindActiveByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	unpaid := orders[:0]
	for _, o := range orders {
		if o.PaymentStatus != common.PaymentStatusPaid {
			unpaid = append(unpaid, o)
		}
	}
	sort.Slice(unpaid, func(i, j int) bool {
		return unpaid[i].CreatedAt.Before(unpaid[j].CreatedAt)
	})
	return unpaid, nil
}
