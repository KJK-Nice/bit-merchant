package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"log/slog"
)

// HourlyOrdersView is the orders-by-hour bar-chart projection.
// Buckets is always 24 entries (00..23) so the renderer can iterate by
// index without bounds checks; PeakHour points at the bucket with the
// highest count (defaults to 0 when all buckets are empty).
type HourlyOrdersView struct {
	Buckets  [24]int
	PeakHour int
	Max      int
	Total    int
}

// OrdersByHour is the query input — restaurant + the active range whose
// orders should be bucketed.
type OrdersByHour struct {
	RestaurantID common.RestaurantID
	Range        DateRange
}

type OrdersByHourHandler decorator.QueryHandler[OrdersByHour, *HourlyOrdersView]

type ordersByHourHandler struct {
	orders OrderReadModel
	now    func() time.Time
}

func NewOrdersByHourHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) OrdersByHourHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := ordersByHourHandler{orders: orders, now: time.Now}
	return decorator.ApplyQueryDecorators[OrdersByHour, *HourlyOrdersView](h, log, metrics)
}

func (h ordersByHourHandler) Handle(ctx context.Context, q OrdersByHour) (*HourlyOrdersView, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	now := h.now()
	start, end := RangeWindow(now, q.Range)
	view := &HourlyOrdersView{}
	for _, o := range orders {
		if o.CreatedAt.Before(start) {
			continue
		}
		if !end.IsZero() && !o.CreatedAt.Before(end) {
			continue
		}
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		hour := o.CreatedAt.In(now.Location()).Hour()
		view.Buckets[hour]++
		view.Total++
	}
	for h, c := range view.Buckets {
		if c > view.Max {
			view.Max = c
			view.PeakHour = h
		}
	}
	return view, nil
}
