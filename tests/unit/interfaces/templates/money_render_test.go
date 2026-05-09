package templates_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/ordering/domain/order"
)

// TestOrderStatus_RendersSatosheTotal proves the niche-currency wiring
// reaches the templated UI: an Order with SAT currency formats as
// "15,000 sats" with a thousands separator and the "sats" suffix —
// not as "$15000.00".
func TestOrderStatus_RendersSatoshiTotal(t *testing.T) {
	items := []order.OrderItem{
		mustOrderItemSAT(t, "oi1", "ord1", "i1", "Espresso", 3, 5_000),
	}
	o, err := order.NewOrderWithCurrency(
		"ord1", "0001", "r1", "sess",
		items, 15_000, common.PaymentMethodTypeCash, money.SAT,
	)
	require.NoError(t, err)

	var sb strings.Builder
	require.NoError(t, templates.OrderStatus(o).Render(context.Background(), &sb))

	out := sb.String()
	assert.Contains(t, out, "15,000 sats", "rendered HTML must show satoshi-formatted total")
	assert.NotContains(t, out, "$15000.00", "rendered HTML must not fall back to USD formatting")
}

func TestOrderStatus_RendersFiatTotal(t *testing.T) {
	items := []order.OrderItem{
		mustOrderItemUSD(t, "oi1", "ord1", "i1", "Burger", 2, 10.00),
	}
	o, err := order.NewOrderWithCurrency(
		"ord1", "0001", "r1", "sess",
		items, 2000, common.PaymentMethodTypeCash, money.USD,
	)
	require.NoError(t, err)

	var sb strings.Builder
	require.NoError(t, templates.OrderStatus(o).Render(context.Background(), &sb))

	assert.Contains(t, sb.String(), "$20.00")
}

func mustOrderItemSAT(t *testing.T, id, oid, mid, name string, qty int, unitPrice float64) order.OrderItem {
	t.Helper()
	oi, err := order.NewOrderItemWithCurrency(common.OrderItemID(id), common.OrderID(oid), common.ItemID(mid), name, qty, unitPrice, money.SAT)
	require.NoError(t, err)
	return *oi
}

func mustOrderItemUSD(t *testing.T, id, oid, mid, name string, qty int, unitPrice float64) order.OrderItem {
	t.Helper()
	oi, err := order.NewOrderItem(common.OrderItemID(id), common.OrderID(oid), common.ItemID(mid), name, qty, unitPrice)
	require.NoError(t, err)
	return *oi
}
