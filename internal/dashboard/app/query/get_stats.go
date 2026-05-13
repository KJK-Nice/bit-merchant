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
		if o.CreatedAt.Before(start) {
			continue
		}
		if !end.IsZero() && !o.CreatedAt.Before(end) {
			continue
		}
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		count++
		totalSales += o.FiatAmount
		if o.PreparingAt != nil && o.ReadyAt != nil {
			d := o.ReadyAt.Sub(*o.PreparingAt)
			if d > 0 {
				prepSumSeconds += d.Seconds()
				prepCount++
			}
		}
	}
	avg := 0.0
	if count > 0 {
		avg = totalSales / float64(count)
	}
	avgPrep := 0.0
	if prepCount > 0 {
		avgPrep = prepSumSeconds / float64(prepCount)
	}
	return PeriodStats{
		OrderCount:        count,
		TotalSales:        totalSales,
		AverageOrderValue: avg,
		AvgPrepSeconds:    avgPrep,
	}
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

// getRangeStart preserves the legacy boundary callers that haven't yet
// migrated to RangeWindow. Returns the inclusive lower bound only.
func getRangeStart(now time.Time, rangeType DateRange) time.Time {
	start, _ := RangeWindow(now, rangeType)
	return start
}
