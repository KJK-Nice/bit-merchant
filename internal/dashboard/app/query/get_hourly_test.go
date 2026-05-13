package query

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type fakeOrderReadModel struct{ orders []*order.Order }

func (f *fakeOrderReadModel) FindByRestaurantID(_ common.RestaurantID) ([]*order.Order, error) {
	return f.orders, nil
}

func (f *fakeOrderReadModel) FindActiveByRestaurantID(_ common.RestaurantID) ([]*order.Order, error) {
	return nil, nil
}

func TestOrdersByHour_BucketsAndPeak(t *testing.T) {
	now := time.Date(2026, 5, 13, 23, 59, 0, 0, time.UTC)
	mk := func(hour, minute int) *order.Order {
		return &order.Order{
			RestaurantID:  "r1",
			PaymentStatus: common.PaymentStatusPaid,
			CreatedAt:     time.Date(2026, 5, 13, hour, minute, 0, 0, time.UTC),
			FiatAmount:    10,
		}
	}
	repo := &fakeOrderReadModel{orders: []*order.Order{
		mk(8, 5),
		mk(8, 30),
		mk(12, 0),
		mk(12, 15),
		mk(12, 45),
		mk(19, 0),
		// Unpaid — should be excluded.
		{RestaurantID: "r1", PaymentStatus: common.PaymentStatusPending, CreatedAt: time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)},
	}}
	h := ordersByHourHandler{orders: repo, now: func() time.Time { return now }}
	view, err := h.Handle(context.Background(), OrdersByHour{RestaurantID: "r1", Range: DateRangeToday})
	if err != nil {
		t.Fatal(err)
	}
	if view.Total != 6 {
		t.Fatalf("expected 6 paid orders, got %d", view.Total)
	}
	if view.Buckets[12] != 3 {
		t.Fatalf("expected hour 12 to have 3, got %d", view.Buckets[12])
	}
	if view.PeakHour != 12 || view.Max != 3 {
		t.Fatalf("expected peak hour 12/max 3, got peak=%d max=%d", view.PeakHour, view.Max)
	}
}

func TestOrdersByHour_EmptyWindow(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	repo := &fakeOrderReadModel{}
	h := ordersByHourHandler{orders: repo, now: func() time.Time { return now }}
	view, err := h.Handle(context.Background(), OrdersByHour{RestaurantID: "r1", Range: DateRangeToday})
	if err != nil {
		t.Fatal(err)
	}
	if view.Total != 0 || view.Max != 0 || view.PeakHour != 0 {
		t.Fatalf("empty window should produce zeroed view, got %+v", view)
	}
}
