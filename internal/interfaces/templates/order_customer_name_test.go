package templates

import (
	"context"
	"strings"
	"testing"

	"bitmerchant/internal/common"
	"bitmerchant/internal/interfaces/templates/components"
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
		TotalAmount:       1800,
		CustomerName:      "Maya",
		TableLabel:        "7",
		PaymentStatus:     common.PaymentStatusPaid,
		FulfillmentStatus: common.FulfillmentStatusPaid,
	}
}

func TestIssue65_KitchenCardShowsCustomerNameAndTable(t *testing.T) {
	var sb strings.Builder
	if err := components.OrderCard(sampleOrder()).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	if !strings.Contains(html, "Maya") {
		t.Errorf("kitchen card missing customer name 'Maya'")
	}
	if !strings.Contains(html, "Table 7") {
		t.Errorf("kitchen card missing 'Table 7'")
	}
}

func TestIssue65_StatusScreenShowsCustomerName(t *testing.T) {
	view := &query.OrderStatusView{Order: sampleOrder()}
	var sb strings.Builder
	if err := OrderStatus(view).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	if !strings.Contains(html, "For Maya") {
		t.Errorf("status screen missing 'For Maya'")
	}
	if !strings.Contains(html, "Table 7") {
		t.Errorf("status screen missing 'Table 7'")
	}
}
