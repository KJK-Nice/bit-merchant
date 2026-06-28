package order_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
)

func mustOrder(t *testing.T, id, num string, restID common.RestaurantID, createdOffset time.Duration, status common.FulfillmentStatus) *order.Order {
	t.Helper()
	item, err := order.NewOrderItem(common.OrderItemID("oi-"+id), common.OrderID(id), common.ItemID("mi-"+id), "Bao", 1, 6.5)
	require.NoError(t, err)
	o, err := order.NewOrder(common.OrderID(id), common.OrderNumber(num), restID, "sess-"+id, []order.OrderItem{*item}, 650, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	o.MarkPaid()
	o.CreatedAt = time.Now().Add(createdOffset)
	o.FulfillmentStatus = status
	return o
}

func TestBuildOrderStatusView_QueueAheadCountsEarlierActiveOrders(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	rest := common.RestaurantID("r1")

	older1 := mustOrder(t, "o1", "0001", rest, -10*time.Minute, common.FulfillmentStatusPaid)
	older2 := mustOrder(t, "o2", "0002", rest, -5*time.Minute, common.FulfillmentStatusPreparing)
	target := mustOrder(t, "o3", "0003", rest, -1*time.Minute, common.FulfillmentStatusPaid)
	newer := mustOrder(t, "o4", "0004", rest, +1*time.Minute, common.FulfillmentStatusPaid)
	other := mustOrder(t, "o5", "0005", common.RestaurantID("r2"), -20*time.Minute, common.FulfillmentStatusPaid)

	for _, o := range []*order.Order{older1, older2, target, newer, other} {
		require.NoError(t, repo.Save(o))
	}

	view, err := query.BuildOrderStatusView(repo, target, query.DefaultPrepTarget)
	require.NoError(t, err)

	assert.Equal(t, 2, view.QueueAhead, "should count two older active orders in same restaurant")
	assert.Equal(t, 3, view.PositionLabel(), "1-indexed position is QueueAhead+1")
}

func TestBuildOrderStatusView_PreparingETAUsesPreparingAt(t *testing.T) {
	o := mustOrder(t, "o", "0010", "r1", -3*time.Minute, common.FulfillmentStatusPreparing)
	prep := time.Now().Add(-2 * time.Minute)
	o.PreparingAt = &prep

	view, err := query.BuildOrderStatusView(nil, o, 10*time.Minute)
	require.NoError(t, err)

	expected := prep.Add(10 * time.Minute)
	assert.WithinDuration(t, expected, view.EstimatedReadyAt, time.Second)
	assert.Equal(t, 0, view.PositionLabel(), "preparing orders no longer carry queue position")
}

func TestBuildOrderStatusView_PaidETAStaggersByQueue(t *testing.T) {
	repo := memory.NewMemoryOrderRepository()
	rest := common.RestaurantID("r1")
	older := mustOrder(t, "older", "0020", rest, -2*time.Minute, common.FulfillmentStatusPaid)
	target := mustOrder(t, "tgt", "0021", rest, -1*time.Minute, common.FulfillmentStatusPaid)
	require.NoError(t, repo.Save(older))
	require.NoError(t, repo.Save(target))

	view, err := query.BuildOrderStatusView(repo, target, 5*time.Minute)
	require.NoError(t, err)

	base := target.CreatedAt
	if target.PaidAt != nil && target.PaidAt.After(base) {
		base = *target.PaidAt
	}
	expected := base.Add(5 * time.Minute * 2) // 1 ahead + self
	assert.WithinDuration(t, expected, view.EstimatedReadyAt, time.Second)
}

func TestBuildOrderStatusView_ReadyUsesReadyAt(t *testing.T) {
	o := mustOrder(t, "o", "0030", "r1", -10*time.Minute, common.FulfillmentStatusReady)
	ready := time.Now().Add(-30 * time.Second)
	o.ReadyAt = &ready

	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)

	assert.True(t, view.IsTerminal())
	assert.WithinDuration(t, ready, view.EstimatedReadyAt, time.Second)
}

func TestBuildOrderStatusView_NilRepoLeavesQueueZero(t *testing.T) {
	o := mustOrder(t, "o", "0040", "r1", 0, common.FulfillmentStatusPaid)
	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)
	assert.Equal(t, 0, view.QueueAhead)
	assert.Equal(t, 1, view.PositionLabel())
}

func TestEstimatedMenuWaitMinutes(t *testing.T) {
	prep := query.DefaultPrepTarget // 10m
	mk := func(s common.FulfillmentStatus) *order.Order { return &order.Order{FulfillmentStatus: s} }

	// Empty queue → base prep target.
	assert.Equal(t, 10, query.EstimatedMenuWaitMinutes(nil, prep))

	// 3 queued (paid/preparing) → one extra prep cycle (10*(1+1)).
	assert.Equal(t, 20, query.EstimatedMenuWaitMinutes(
		[]*order.Order{mk(common.FulfillmentStatusPaid), mk(common.FulfillmentStatusPreparing), mk(common.FulfillmentStatusPaid)}, prep))

	// Ready/completed orders are out of the pipeline and don't count.
	assert.Equal(t, 10, query.EstimatedMenuWaitMinutes(
		[]*order.Order{mk(common.FulfillmentStatusReady), mk(common.FulfillmentStatusCompleted), mk(common.FulfillmentStatusPaid)}, prep))

	// Busy rush is capped at 90 minutes.
	big := make([]*order.Order, 60)
	for i := range big {
		big[i] = mk(common.FulfillmentStatusPreparing)
	}
	assert.Equal(t, 90, query.EstimatedMenuWaitMinutes(big, prep))
}
