package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
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

type GetDashboardStatsUseCase struct {
	orderRepo order.Repository
}

func NewGetDashboardStatsUseCase(orderRepo order.Repository) *GetDashboardStatsUseCase {
	return &GetDashboardStatsUseCase{orderRepo: orderRepo}
}

func (uc *GetDashboardStatsUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, rangeType DateRange) (*DashboardStats, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	var count int
	var totalSales float64

	now := time.Now()
	rangeStart := getRangeStart(now, rangeType)

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
