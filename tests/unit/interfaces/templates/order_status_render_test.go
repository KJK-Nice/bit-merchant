package templates_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmerchant/internal/common"
	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
)

func renderStatus(t *testing.T, view *query.OrderStatusView) string {
	t.Helper()
	var sb strings.Builder
	require.NoError(t, templates.OrderStatus(view).Render(context.Background(), &sb))
	return sb.String()
}

func paidOrder(t *testing.T) *order.Order {
	t.Helper()
	item, err := order.NewOrderItem("oi", "ord", "mi", "Pork Belly Bao", 2, 6.5)
	require.NoError(t, err)
	o, err := order.NewOrder("ord", "A29F", "r1", "sess", []order.OrderItem{*item}, 1300, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	o.MarkPaid()
	return o
}

func TestOrderStatus_RendersTimelineSteps(t *testing.T) {
	o := paidOrder(t)
	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)

	out := renderStatus(t, view)
	for _, label := range []string{"Sent to kitchen", "Preparing", "Ready to serve", "Delivered to table"} {
		assert.Contains(t, out, label, "timeline must include step %q", label)
	}
}

func TestOrderStatus_PaidOrderShowsQueuePosition(t *testing.T) {
	o := paidOrder(t)
	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)
	view.QueueAhead = 2

	out := renderStatus(t, view)
	assert.Contains(t, out, "Position in queue")
	assert.Contains(t, out, "2 ahead of you")
	assert.Contains(t, out, "#3")
}

func TestOrderStatus_PreparingOrderHidesQueueAndShowsETA(t *testing.T) {
	o := paidOrder(t)
	prep := time.Now().Add(-2 * time.Minute)
	o.PreparingAt = &prep
	o.FulfillmentStatus = common.FulfillmentStatusPreparing

	view, err := query.BuildOrderStatusView(nil, o, 10*time.Minute)
	require.NoError(t, err)

	out := renderStatus(t, view)
	assert.NotContains(t, out, "Position in queue", "queue card hidden once preparing")
	assert.Contains(t, out, "Cooking now")
	assert.Contains(t, out, "Should be ready at")
}

func TestOrderStatus_ReadyOrderShowsReadyHeader(t *testing.T) {
	o := paidOrder(t)
	o.FulfillmentStatus = common.FulfillmentStatusReady
	now := time.Now()
	o.ReadyAt = &now

	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)

	out := renderStatus(t, view)
	assert.Contains(t, out, "Ready to serve")
}

func TestOrderStatus_UnpaidShowsCashUnpaidBadge(t *testing.T) {
	o := paidOrder(t)
	o.PaymentStatus = common.PaymentStatusPending

	view, err := query.BuildOrderStatusView(nil, o, query.DefaultPrepTarget)
	require.NoError(t, err)

	out := renderStatus(t, view)
	assert.Contains(t, out, "Cash · unpaid")
}
