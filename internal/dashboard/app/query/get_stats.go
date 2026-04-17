package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"log/slog"
)

type DateRange string

const (
	DateRangeToday DateRange = "today"
	DateRangeWeek  DateRange = "week"
	DateRangeMonth DateRange = "month"
)

type DashboardStats struct {
	OrderCount        int
	TotalSales        float64
	AverageOrderValue float64
}

// RestaurantDashboardStats returns aggregate stats for a restaurant in a date range.
type RestaurantDashboardStats struct {
	RestaurantID common.RestaurantID
	Range        DateRange
}

type RestaurantDashboardStatsHandler decorator.QueryHandler[RestaurantDashboardStats, *DashboardStats]

type restaurantDashboardStatsHandler struct {
	orders OrderReadModel
}

func NewRestaurantDashboardStatsHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) RestaurantDashboardStatsHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := restaurantDashboardStatsHandler{orders: orders}
	return decorator.ApplyQueryDecorators[RestaurantDashboardStats, *DashboardStats](h, log, metrics)
}

func (h restaurantDashboardStatsHandler) Handle(ctx context.Context, q RestaurantDashboardStats) (*DashboardStats, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}

	var count int
	var totalSales float64

	now := time.Now()
	rangeStart := getRangeStart(now, q.Range)

	for _, o := range orders {
		if o.CreatedAt.Before(rangeStart) {
			continue
		}
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		count++
		totalSales += o.FiatAmount
	}

	avg := 0.0
	if count > 0 {
		avg = totalSales / float64(count)
	}

	return &DashboardStats{
		OrderCount: count, TotalSales: totalSales, AverageOrderValue: avg,
	}, nil
}

func getRangeStart(now time.Time, rangeType DateRange) time.Time {
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch rangeType {
	case DateRangeToday:
		return startOfToday
	case DateRangeWeek:
		return startOfToday.AddDate(0, 0, -6)
	case DateRangeMonth:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return startOfToday
	}
}
