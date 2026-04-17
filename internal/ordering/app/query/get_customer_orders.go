package query

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
)

// CustomerOrdersForSession lists orders for a customer session, newest first.
type CustomerOrdersForSession struct {
	SessionID string
}

type CustomerOrdersForSessionHandler decorator.QueryHandler[CustomerOrdersForSession, []*order.Order]

type customerOrdersForSessionHandler struct {
	orderRepo order.Repository
}

func NewCustomerOrdersForSessionHandler(orderRepo order.Repository, log *slog.Logger, metrics decorator.MetricsClient) CustomerOrdersForSessionHandler {
	h := customerOrdersForSessionHandler{orderRepo: orderRepo}
	return decorator.ApplyQueryDecorators[CustomerOrdersForSession, []*order.Order](h, log, metrics)
}

func (h customerOrdersForSessionHandler) Handle(ctx context.Context, q CustomerOrdersForSession) ([]*order.Order, error) {
	_ = ctx
	orders, err := h.orderRepo.FindBySessionID(q.SessionID)
	if err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

// CustomerOrderByLookup finds an order by session and human-readable order number.
type CustomerOrderByLookup struct {
	SessionID   string
	OrderNumber string
}

type CustomerOrderByLookupHandler decorator.QueryHandler[CustomerOrderByLookup, *order.Order]

type customerOrderByLookupHandler struct {
	orderRepo order.Repository
}

func NewCustomerOrderByLookupHandler(orderRepo order.Repository, log *slog.Logger, metrics decorator.MetricsClient) CustomerOrderByLookupHandler {
	h := customerOrderByLookupHandler{orderRepo: orderRepo}
	return decorator.ApplyQueryDecorators[CustomerOrderByLookup, *order.Order](h, log, metrics)
}

func (h customerOrderByLookupHandler) Handle(ctx context.Context, q CustomerOrderByLookup) (*order.Order, error) {
	_ = ctx
	if q.SessionID == "" || q.OrderNumber == "" {
		return nil, fmt.Errorf("order not found")
	}
	orders, err := h.orderRepo.FindBySessionID(q.SessionID)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		if string(o.OrderNumber) == q.OrderNumber {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}
