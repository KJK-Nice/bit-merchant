package templates

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/app/cart"
	"bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
)

func sampleOrder() *order.Order {
	return &order.Order{
		ID:                common.OrderID("ord-1"),
		OrderNumber:       common.OrderNumber("A29F"),
		RestaurantID:      common.RestaurantID("rest-1"),
		SessionID:         "sess-1",
		Items:             []order.OrderItem{{ID: common.OrderItemID("oi-1"), Name: "Bao Bun", Quantity: 2}},
		Subtotal:          1500,
		TaxAmount:         120,
		TipAmount:         180,
		TotalAmount:       1800,
		CustomerName:      "Maya",
		TableLabel:        "7",
		PaymentStatus:     common.PaymentStatusPaid,
		FulfillmentStatus: common.FulfillmentStatusPaid,
	}
}

func renderCard(t *testing.T, o *order.Order) string {
	t.Helper()
	var sb strings.Builder
	if err := components.OrderCard(o).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func renderStatus(t *testing.T, view *query.OrderStatusView) string {
	t.Helper()
	var sb strings.Builder
	if err := OrderStatus(view).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

// #65 — pickup name shown on the kitchen ticket and status screen.
func TestIssue65_KitchenCardShowsCustomerNameAndTable(t *testing.T) {
	html := renderCard(t, sampleOrder())
	for _, want := range []string{"Maya", "Table 7"} {
		if !strings.Contains(html, want) {
			t.Errorf("kitchen card missing %q", want)
		}
	}
}

func TestIssue65_StatusScreenShowsCustomerName(t *testing.T) {
	html := renderStatus(t, &query.OrderStatusView{Order: sampleOrder()})
	for _, want := range []string{"For Maya", "Table 7"} {
		if !strings.Contains(html, want) {
			t.Errorf("status screen missing %q", want)
		}
	}
}

// #77 — kitchen card leads with table + customer name; order # is secondary mono.
func TestIssue77_KitchenCardLeadsWithTableAndName(t *testing.T) {
	html := renderCard(t, sampleOrder())
	if !strings.Contains(html, "font-mono") {
		t.Errorf("kitchen card order number not rendered as mono secondary")
	}
	// The visible secondary handle renders as "#A29F" (the bare "A29F" also appears
	// in the data-order-number attribute, so match the hashed visible form).
	if !strings.Contains(html, "#A29F") {
		t.Errorf("kitchen card missing visible order handle #A29F")
	}
	// The large title is the table/name, so "Table 7" must precede the visible handle.
	if iTable, iNum := strings.Index(html, "Table 7"), strings.Index(html, "#A29F"); iTable == -1 || iNum == -1 || iTable > iNum {
		t.Errorf("expected Table 7 before order handle #A29F (table=%d num=%d)", iTable, iNum)
	}
}

// When no handle is set, the card falls back to the order number as the title.
func TestIssue77_KitchenCardFallsBackToOrderNumber(t *testing.T) {
	o := sampleOrder()
	o.CustomerName = ""
	o.TableLabel = ""
	html := renderCard(t, o)
	if !strings.Contains(html, "Order #A29F") {
		t.Errorf("kitchen card without handle should show 'Order #A29F'")
	}
}

// #72 — a standalone, saveable receipt page renders the order's key details.
func TestIssue72_ReceiptRenders(t *testing.T) {
	var sb strings.Builder
	if err := OrderReceiptPage(sampleOrder(), "Bao & Brew").Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	for _, want := range []string{"Brew", "A29F", "Bao Bun", "Total", "Save as PDF / Print", "Subtotal"} {
		if !strings.Contains(html, want) {
			t.Errorf("receipt missing %q", want)
		}
	}
}

// #72 — the order-status page links to the saveable receipt.
func TestIssue72_StatusLinksToReceipt(t *testing.T) {
	html := renderStatus(t, &query.OrderStatusView{Order: sampleOrder()})
	if !strings.Contains(html, "/order/A29F/receipt") {
		t.Errorf("status page missing receipt link")
	}
}

// #66 — status screen shows the itemised breakdown, not just a total.
func TestIssue66_StatusScreenShowsBreakdown(t *testing.T) {
	html := renderStatus(t, &query.OrderStatusView{Order: sampleOrder()})
	for _, want := range []string{"Subtotal", "Tax", "Tip", "Total"} {
		if !strings.Contains(html, want) {
			t.Errorf("status breakdown missing %q row", want)
		}
	}
}

// #66 — confirm page wires the tip chip to the displayed Tip/Total so the shown
// total always equals the total the server will charge for the selected tip.
func TestIssue66_ConfirmTipIsReactiveAndExact(t *testing.T) {
	c := &cart.Cart{
		RestaurantID: common.RestaurantID("rest-1"),
		Items: []cart.CartItem{
			{ItemID: common.ItemID("i1"), Name: "Bao Bun", Quantity: 2, UnitPrice: 9.00, Subtotal: 18.00},
		},
		Total: 18.00,
	}
	var sb strings.Builder
	comp := OrderConfirmationPage(c, "rest-1", "Bao & Brew", "7", 0.08, 10*time.Minute, "tok", "")
	if err := comp.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()

	// Reactive wiring: a Tip row plus per-tier data-show toggles driven by $tipPercent.
	if !strings.Contains(html, ">Tip<") {
		t.Errorf("confirm page missing a Tip row")
	}
	for _, pct := range cart.AllowedTipPercents {
		needle := "$tipPercent == " + strconv.Itoa(pct)
		if !strings.Contains(html, needle) {
			t.Errorf("confirm page missing data-show toggle for %q", needle)
		}
	}

	// Exactness: the server-computed total for each tier must be rendered, so the
	// shown total matches what /order/create will charge. subtotal 18.00 + 8% tax:
	//   no tip  -> 19.44
	//   20% tip -> 23.04
	for _, want := range []string{"19.44", "23.04"} {
		if !strings.Contains(html, want) {
			t.Errorf("confirm page missing exact total %q for a tip tier", want)
		}
	}
}
