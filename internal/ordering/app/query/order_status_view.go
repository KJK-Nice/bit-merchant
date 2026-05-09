package query

import (
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

// DefaultPrepTarget is the per-order kitchen prep target used when a restaurant
// has not configured its own. Tracked under #76 (configurable thresholds) — for
// now this single value drives ETA, kitchen overdue tiers, and dashboard alerts.
const DefaultPrepTarget = 10 * time.Minute

// OrderStatusView is a render-time projection of an order with the two pieces
// the customer cares about on the status screen: where they sit in the queue
// and when their food should be ready.
type OrderStatusView struct {
	Order            *order.Order
	QueueAhead       int
	EstimatedReadyAt time.Time
	PrepTarget       time.Duration
}

// PositionLabel is the 1-indexed queue slot ("#3"). Once the order is
// preparing or further, position is no longer meaningful and we return 0.
func (v OrderStatusView) PositionLabel() int {
	if v.Order == nil {
		return 0
	}
	switch v.Order.FulfillmentStatus {
	case common.FulfillmentStatusPaid:
		return v.QueueAhead + 1
	default:
		return 0
	}
}

// IsTerminal reports whether the order has reached ready or completed —
// at which point the timeline shows the final tick and the ETA is the actual
// ReadyAt/CompletedAt timestamp, not a projection.
func (v OrderStatusView) IsTerminal() bool {
	if v.Order == nil {
		return false
	}
	return v.Order.FulfillmentStatus == common.FulfillmentStatusReady ||
		v.Order.FulfillmentStatus == common.FulfillmentStatusCompleted
}

// BuildOrderStatusView projects an order plus its current queue context.
// When repo is nil, ETA is computed but QueueAhead falls back to 0 — useful
// for unit tests that exercise rendering without a real repository.
func BuildOrderStatusView(repo order.Repository, o *order.Order, prepTarget time.Duration) (*OrderStatusView, error) {
	if prepTarget <= 0 {
		prepTarget = DefaultPrepTarget
	}
	view := &OrderStatusView{Order: o, PrepTarget: prepTarget}
	if o == nil {
		return view, nil
	}

	if repo != nil {
		active, err := repo.FindActiveByRestaurantID(o.RestaurantID)
		if err != nil {
			return nil, err
		}
		view.QueueAhead = countAhead(active, o)
	}
	view.EstimatedReadyAt = computeETA(o, view.QueueAhead, prepTarget)
	return view, nil
}

func countAhead(active []*order.Order, target *order.Order) int {
	ahead := 0
	for _, a := range active {
		if a == nil || a.ID == target.ID {
			continue
		}
		if !a.CreatedAt.Before(target.CreatedAt) {
			continue
		}
		switch a.FulfillmentStatus {
		case common.FulfillmentStatusPaid, common.FulfillmentStatusPreparing:
			ahead++
		}
	}
	return ahead
}

func computeETA(o *order.Order, queueAhead int, prepTarget time.Duration) time.Time {
	switch o.FulfillmentStatus {
	case common.FulfillmentStatusReady:
		if o.ReadyAt != nil {
			return *o.ReadyAt
		}
		return time.Now()
	case common.FulfillmentStatusCompleted:
		if o.CompletedAt != nil {
			return *o.CompletedAt
		}
		return time.Now()
	case common.FulfillmentStatusPreparing:
		base := o.CreatedAt
		if o.PreparingAt != nil {
			base = *o.PreparingAt
		}
		return base.Add(prepTarget)
	default: // paid / pending
		base := o.CreatedAt
		if o.PaidAt != nil && o.PaidAt.After(base) {
			base = *o.PaidAt
		}
		return base.Add(prepTarget * time.Duration(1+queueAhead))
	}
}
