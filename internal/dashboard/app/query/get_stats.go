package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
	"log/slog"
)

type DateRange string

const (
	DateRangeToday DateRange = "today"
	DateRangeWeek  DateRange = "week"
	DateRangeMonth DateRange = "month"
)

// PeriodStats is the subset of metrics computed for a single window. Counts
// are nil-safe — TotalSales / OrderCount / AvgPrepSeconds are zero when no
// paid orders fell in the window.
type PeriodStats struct {
	OrderCount        int
	TotalSales        float64
	AverageOrderValue float64
	// AvgPrepSeconds is the mean (PreparingAt..ReadyAt) duration over orders
	// that crossed both transitions inside the window. Zero when the window
	// holds no fully-prepared orders.
	AvgPrepSeconds float64
}

// DashboardStats now bundles the active period plus the prior comparable
// period so the template can render deltas without re-computing math.
type DashboardStats struct {
	OrderCount        int
	TotalSales        float64
	AverageOrderValue float64
	AvgPrepSeconds    float64
	// Previous holds the same metrics over the prior comparable window
	// (today vs yesterday, week vs the seven days before, month vs the
	// preceding calendar month). Used only for delta display.
	Previous PeriodStats
}

// RestaurantDashboardStats returns aggregate stats for a restaurant in a date range.
type RestaurantDashboardStats struct {
	RestaurantID common.RestaurantID
	Range        DateRange
}

type RestaurantDashboardStatsHandler decorator.QueryHandler[RestaurantDashboardStats, *DashboardStats]

type restaurantDashboardStatsHandler struct {
	orders OrderReadModel
	now    func() time.Time
}

func NewRestaurantDashboardStatsHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) RestaurantDashboardStatsHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := restaurantDashboardStatsHandler{orders: orders, now: time.Now}
	return decorator.ApplyQueryDecorators[RestaurantDashboardStats, *DashboardStats](h, log, metrics)
}

func (h restaurantDashboardStatsHandler) Handle(ctx context.Context, q RestaurantDashboardStats) (*DashboardStats, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	now := h.now()
	curStart, curEnd := RangeWindow(now, q.Range)
	prevStart, prevEnd := PreviousWindow(now, q.Range)

	cur := computePeriodStats(orders, curStart, curEnd)
	prev := computePeriodStats(orders, prevStart, prevEnd)

	return &DashboardStats{
		OrderCount:        cur.OrderCount,
		TotalSales:        cur.TotalSales,
		AverageOrderValue: cur.AverageOrderValue,
		AvgPrepSeconds:    cur.AvgPrepSeconds,
		Previous:          prev,
	}, nil
}

// computePeriodStats walks paid orders within [start, end) and returns the
// summary tile data. AvgPrepSeconds is averaged across orders that have
// both PreparingAt and ReadyAt — best signal of actual kitchen throughput
// for the window in which they were created.
func computePeriodStats(orders []*order.Order, start, end time.Time) PeriodStats {
	var (
		count          int
		totalSales     float64
		prepSumSeconds float64
		prepCount      int
	)
	for _, o := range orders {
		if !orderInWindow(o, start, end) {
			continue
		}
		count++
		totalSales += o.FiatAmount
		if d, ok := orderPrepDuration(o); ok {
			prepSumSeconds += d.Seconds()
			prepCount++
		}
	}
	return PeriodStats{
		OrderCount:        count,
		TotalSales:        totalSales,
		AverageOrderValue: safeMean(totalSales, count),
		AvgPrepSeconds:    safeMean(prepSumSeconds, prepCount),
	}
}

// orderInWindow reports whether a paid order falls in [start, end). end == 0
// is treated as open-ended (no upper bound).
func orderInWindow(o *order.Order, start, end time.Time) bool {
	if o.PaymentStatus != common.PaymentStatusPaid {
		return false
	}
	if o.CreatedAt.Before(start) {
		return false
	}
	if !end.IsZero() && !o.CreatedAt.Before(end) {
		return false
	}
	return true
}

// orderPrepDuration returns the kitchen prep duration (ReadyAt - PreparingAt)
// when both timestamps are set and positive, otherwise ok=false.
func orderPrepDuration(o *order.Order) (time.Duration, bool) {
	if o.PreparingAt == nil || o.ReadyAt == nil {
		return 0, false
	}
	d := o.ReadyAt.Sub(*o.PreparingAt)
	if d <= 0 {
		return 0, false
	}
	return d, true
}

func safeMean(sum float64, n int) float64 {
	if n <= 0 {
		return 0
	}
	return sum / float64(n)
}

// RangeWindow returns [start, end) for the active range. Today/week/month
// match the existing semantics; end == now for live ranges so prior-period
// math compares apples to apples (e.g. "today through noon" vs
// "yesterday through noon").
func RangeWindow(now time.Time, rangeType DateRange) (time.Time, time.Time) {
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch rangeType {
	case DateRangeWeek:
		return startOfToday.AddDate(0, 0, -6), now
	case DateRangeMonth:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), now
	case DateRangeToday:
		return startOfToday, now
	default:
		return startOfToday, now
	}
}

// PreviousWindow returns the previous comparable [start, end) window used for
// KPI delta math. "Today" compares against yesterday up to the same wall
// clock time; "week" compares against the prior 7 days; "month" compares
// against the preceding calendar month up to the same calendar offset.
func PreviousWindow(now time.Time, rangeType DateRange) (time.Time, time.Time) {
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch rangeType {
	case DateRangeWeek:
		end := startOfToday.AddDate(0, 0, -6)
		return end.AddDate(0, 0, -7), end
	case DateRangeMonth:
		thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		prevMonth := thisMonth.AddDate(0, -1, 0)
		// Compare same-day offset within the month, capped at the prior
		// month's last day so a 31st doesn't roll into the next month.
		offset := now.Sub(thisMonth)
		end := prevMonth.Add(offset)
		if end.After(thisMonth) {
			end = thisMonth
		}
		return prevMonth, end
	case DateRangeToday:
		yesterdayStart := startOfToday.AddDate(0, 0, -1)
		offset := now.Sub(startOfToday)
		return yesterdayStart, yesterdayStart.Add(offset)
	default:
		yesterdayStart := startOfToday.AddDate(0, 0, -1)
		offset := now.Sub(startOfToday)
		return yesterdayStart, yesterdayStart.Add(offset)
	}
}
