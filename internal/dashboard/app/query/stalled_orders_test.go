package query

import (
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

func ptrTime(t time.Time) *time.Time { return &t }

func mkOrder(id string, num string, status common.FulfillmentStatus, createdAt time.Time, preparingAt *time.Time) *order.Order {
	return &order.Order{
		ID:                common.OrderID(id),
		OrderNumber:       common.OrderNumber(num),
		FulfillmentStatus: status,
		CreatedAt:         createdAt,
		PreparingAt:       preparingAt,
	}
}

func TestBuildStalledOrdersView_Empty(t *testing.T) {
	now := time.Now()
	view := BuildStalledOrdersView(nil, 10*time.Minute, now)
	if view.Count != 0 || view.Sample != nil {
		t.Fatalf("expected empty view, got count=%d sample=%v", view.Count, view.Sample)
	}
	if view.Threshold != 10*time.Minute {
		t.Fatalf("threshold not echoed back")
	}
}

func TestBuildStalledOrdersView_UnderThreshold(t *testing.T) {
	now := time.Now()
	orders := []*order.Order{
		mkOrder("a", "A001", common.FulfillmentStatusPreparing, now.Add(-9*time.Minute), nil),
	}
	view := BuildStalledOrdersView(orders, 10*time.Minute, now)
	if view.Count != 0 {
		t.Fatalf("expected 0, got %d", view.Count)
	}
}

func TestBuildStalledOrdersView_OverThreshold(t *testing.T) {
	now := time.Now()
	orders := []*order.Order{
		mkOrder("a", "A001", common.FulfillmentStatusPreparing, now.Add(-16*time.Minute), nil),
	}
	view := BuildStalledOrdersView(orders, 10*time.Minute, now)
	if view.Count != 1 {
		t.Fatalf("expected 1, got %d", view.Count)
	}
	if view.Sample == nil || view.Sample.ID != "a" {
		t.Fatalf("sample missing or wrong: %+v", view.Sample)
	}
	if got := view.SampleAgeMinutes(); got != 16 {
		t.Fatalf("expected 16m age, got %d", got)
	}
	if got := view.ThresholdMinutes(); got != 10 {
		t.Fatalf("expected 10m threshold, got %d", got)
	}
}

func TestBuildStalledOrdersView_PicksLongest(t *testing.T) {
	now := time.Now()
	orders := []*order.Order{
		mkOrder("a", "A001", common.FulfillmentStatusPreparing, now.Add(-15*time.Minute), nil),
		mkOrder("b", "B002", common.FulfillmentStatusPreparing, now.Add(-22*time.Minute), nil),
		mkOrder("c", "C003", common.FulfillmentStatusPaid, now.Add(-18*time.Minute), nil),
	}
	view := BuildStalledOrdersView(orders, 10*time.Minute, now)
	if view.Count != 3 {
		t.Fatalf("expected 3, got %d", view.Count)
	}
	if view.Sample == nil || view.Sample.ID != "b" {
		t.Fatalf("expected sample b (oldest), got %+v", view.Sample)
	}
}

func TestBuildStalledOrdersView_UsesPreparingAt(t *testing.T) {
	now := time.Now()
	created := now.Add(-30 * time.Minute)
	preparing := now.Add(-5 * time.Minute)
	orders := []*order.Order{
		mkOrder("a", "A001", common.FulfillmentStatusPreparing, created, ptrTime(preparing)),
	}
	view := BuildStalledOrdersView(orders, 10*time.Minute, now)
	if view.Count != 0 {
		t.Fatalf("PreparingAt should reset the clock; expected 0, got %d", view.Count)
	}
}

func TestBuildStalledOrdersView_IgnoresTerminalStates(t *testing.T) {
	now := time.Now()
	orders := []*order.Order{
		mkOrder("ready", "R001", common.FulfillmentStatusReady, now.Add(-30*time.Minute), nil),
		mkOrder("completed", "K002", common.FulfillmentStatusCompleted, now.Add(-30*time.Minute), nil),
	}
	view := BuildStalledOrdersView(orders, 10*time.Minute, now)
	if view.Count != 0 {
		t.Fatalf("ready/completed should not count, got %d", view.Count)
	}
}
