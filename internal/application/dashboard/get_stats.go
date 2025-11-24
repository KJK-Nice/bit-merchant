package dashboard

import (
	"context"
	"time"

	"bitmerchant/internal/domain"
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
	orderRepo domain.OrderRepository
}

func NewGetDashboardStatsUseCase(orderRepo domain.OrderRepository) *GetDashboardStatsUseCase {
	return &GetDashboardStatsUseCase{
		orderRepo: orderRepo,
	}
}

func (uc *GetDashboardStatsUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID, rangeType DateRange) (*DashboardStats, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	var count int
	var totalSales float64
	
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, o := range orders {
		// Filter by date
		if o.CreatedAt.Before(startOfDay) {
			continue
		}

		// Filter by status (only paid orders count for sales)
		if o.PaymentStatus != domain.PaymentStatusPaid {
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
		OrderCount:        count,
		TotalSales:        totalSales,
		AverageOrderValue: avg,
	}, nil
}

