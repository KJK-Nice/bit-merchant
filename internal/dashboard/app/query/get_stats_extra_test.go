package query

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

// TestRestaurantDashboardStats_AvgPrepAndPrevious verifies the avg-prep
// computation (mean of ReadyAt-PreparingAt across orders with both timestamps)
// and that the prior-period window is filled so the template can render deltas.
func TestRestaurantDashboardStats_AvgPrepAndPrevious(t *testing.T) {
	loc := time.UTC
	// 2026-05-13 13:00 UTC — well into the day for "today" math.
	now := time.Date(2026, 5, 13, 13, 0, 0, 0, loc)

	preparing := now.Add(-30 * time.Minute)
	ready := now.Add(-22 * time.Minute) // 8 min prep
	yesterdayPrep := now.AddDate(0, 0, -1).Add(-30 * time.Minute)
	yesterdayReady := yesterdayPrep.Add(12 * time.Minute) // 12 min prep

	mkPaid := func(t time.Time, amount float64, prepStart, prepEnd *time.Time) *order.Order {
		return &order.Order{
			RestaurantID:  "r1",
			PaymentStatus: common.PaymentStatusPaid,
			CreatedAt:     t,
			FiatAmount:    amount,
			PreparingAt:   prepStart,
			ReadyAt:       prepEnd,
		}
	}
	repo := &fakeOrderReadModel{orders: []*order.Order{
		mkPaid(now.Add(-1*time.Hour), 20, &preparing, &ready),                                // today, 8m prep
		mkPaid(now.Add(-2*time.Hour), 10, nil, nil),                                          // today, no prep timestamps
		mkPaid(now.AddDate(0, 0, -1).Add(-3*time.Hour), 30, &yesterdayPrep, &yesterdayReady), // yesterday, 12m prep
	}}
	h := restaurantDashboardStatsHandler{orders: repo, now: func() time.Time { return now }}
	stats, err := h.Handle(context.Background(), RestaurantDashboardStats{RestaurantID: "r1", Range: DateRangeToday})
	if err != nil {
		t.Fatal(err)
	}
	if stats.OrderCount != 2 {
		t.Fatalf("expected 2 today, got %d", stats.OrderCount)
	}
	if stats.TotalSales != 30 {
		t.Fatalf("expected $30 today, got %.2f", stats.TotalSales)
	}
	if stats.AvgPrepSeconds != (8 * 60) {
		t.Fatalf("expected 480s avg prep today, got %.0f", stats.AvgPrepSeconds)
	}
	if stats.Previous.OrderCount != 1 {
		t.Fatalf("expected 1 prior, got %d", stats.Previous.OrderCount)
	}
	if stats.Previous.AvgPrepSeconds != (12 * 60) {
		t.Fatalf("expected 720s avg prep prior, got %.0f", stats.Previous.AvgPrepSeconds)
	}
}
