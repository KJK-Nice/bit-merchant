package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	"log/slog"
)

// StalledOrders is the query input — the restaurant whose in-flight orders
// should be inspected.
type StalledOrders struct {
	RestaurantID common.RestaurantID
}

// StalledOrdersView is the projection consumed by the dashboard banner.
// Count is the total number of in-flight orders past the overdue threshold;
// Sample is the longest-running of those orders (nil when Count is 0).
// Threshold and Now are echoed back so the renderer can compute the relative
// minute count without re-deriving them.
type StalledOrdersView struct {
	Count     int
	Sample    *order.Order
	Threshold time.Duration
	Now       time.Time
}

// SampleAgeMinutes returns how long the longest stalled order has been
// in-flight, rounded down to whole minutes. Falls back to 0 when no sample.
func (v StalledOrdersView) SampleAgeMinutes() int {
	if v.Sample == nil {
		return 0
	}
	return int(v.Now.Sub(stalledReference(v.Sample)).Minutes())
}

// ThresholdMinutes returns the configured cutoff in whole minutes for display.
func (v StalledOrdersView) ThresholdMinutes() int {
	return int(v.Threshold.Minutes())
}

type StalledOrdersHandler decorator.QueryHandler[StalledOrders, *StalledOrdersView]

type stalledOrdersHandler struct {
	orders OrderReadModel
	now    func() time.Time
}

// NewStalledOrdersHandler wires the read-side handler. The handler reuses the
// shared OverdueThreshold from the ordering query package so the dashboard
// banner and the kitchen overdue ring stay in agreement.
func NewStalledOrdersHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) StalledOrdersHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := stalledOrdersHandler{orders: orders, now: time.Now}
	return decorator.ApplyQueryDecorators[StalledOrders, *StalledOrdersView](h, log, metrics)
}

func (h stalledOrdersHandler) Handle(ctx context.Context, q StalledOrders) (*StalledOrdersView, error) {
	_ = ctx
	active, err := h.orders.FindActiveByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	return BuildStalledOrdersView(active, query.OverdueThreshold(), h.now()), nil
}

// BuildStalledOrdersView is the pure projection — given a snapshot of active
// orders, the configured threshold, and "now", return the banner view.
// Exposed for unit tests; the production path goes through Handle.
func BuildStalledOrdersView(active []*order.Order, threshold time.Duration, now time.Time) *StalledOrdersView {
	view := &StalledOrdersView{Threshold: threshold, Now: now}
	var longest *order.Order
	var longestAge time.Duration
	for _, o := range active {
		if !isStalledCandidate(o) {
			continue
		}
		age := now.Sub(stalledReference(o))
		if age <= threshold {
			continue
		}
		view.Count++
		if longest == nil || age > longestAge {
			longest = o
			longestAge = age
		}
	}
	view.Sample = longest
	return view
}

// isStalledCandidate returns whether the order is in a state where a long
// wait means trouble — paid (still in queue) or actively preparing.
func isStalledCandidate(o *order.Order) bool {
	if o == nil {
		return false
	}
	switch o.FulfillmentStatus {
	case common.FulfillmentStatusPaid, common.FulfillmentStatusPreparing:
		return true
	default:
		return false
	}
}

// stalledReference picks the timestamp the kitchen card ages from: PreparingAt
// when the order has been pulled into the lane, otherwise CreatedAt.
func stalledReference(o *order.Order) time.Time {
	if o.PreparingAt != nil {
		return *o.PreparingAt
	}
	return o.CreatedAt
}
