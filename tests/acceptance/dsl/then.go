package dsl

import (
	"bitmerchant/internal/domain"
)

// ThenBuilder provides a fluent API for defining assertions
type ThenBuilder struct {
	scenario *Scenario
}

// OrderShouldBe asserts order state
func (t *ThenBuilder) OrderShouldBe(orderNumber string) *OrderAssertion {
	assertion := &OrderAssertion{
		OrderNumber: orderNumber,
	}
	t.scenario.addAssertion(assertion)
	return assertion
}

// OrderAssertion provides assertions for order state
type OrderAssertion struct {
	OrderNumber        string
	PaymentStatus      *domain.PaymentStatus
	FulfillmentStatus  *domain.FulfillmentStatus
	ExpectedHTML       []string
}

// WithPaymentStatus asserts payment status
func (a *OrderAssertion) WithPaymentStatus(status domain.PaymentStatus) *OrderAssertion {
	a.PaymentStatus = &status
	return a
}

// WithFulfillmentStatus asserts fulfillment status
func (a *OrderAssertion) WithFulfillmentStatus(status domain.FulfillmentStatus) *OrderAssertion {
	a.FulfillmentStatus = &status
	return a
}

// ContainsHTML asserts that the response contains specific HTML
func (a *OrderAssertion) ContainsHTML(html string) *OrderAssertion {
	a.ExpectedHTML = append(a.ExpectedHTML, html)
	return a
}

// SSEStreamShouldReceive asserts SSE events
func (t *ThenBuilder) SSEStreamShouldReceive(stream string) *SSEAssertion {
	assertion := &SSEAssertion{
		Stream: stream,
	}
	t.scenario.addAssertion(assertion)
	return assertion
}

// SSEAssertion provides assertions for SSE events
type SSEAssertion struct {
	Stream       string
	EventType    string
	ExpectedHTML []string
	Selector     string
}

// Event asserts the event type
func (a *SSEAssertion) Event(eventType string) *SSEAssertion {
	a.EventType = eventType
	return a
}

// WithSelector asserts the CSS selector
func (a *SSEAssertion) WithSelector(selector string) *SSEAssertion {
	a.Selector = selector
	return a
}

// ContainsHTML asserts that the SSE event contains specific HTML
func (a *SSEAssertion) ContainsHTML(html string) *SSEAssertion {
	a.ExpectedHTML = append(a.ExpectedHTML, html)
	return a
}

// KitchenDashboardShouldShow asserts kitchen dashboard state
func (t *ThenBuilder) KitchenDashboardShouldShow() *KitchenDashboardAssertion {
	assertion := &KitchenDashboardAssertion{}
	t.scenario.addAssertion(assertion)
	return assertion
}

// KitchenDashboardAssertion provides assertions for kitchen dashboard
type KitchenDashboardAssertion struct {
	ExpectedOrderCount *int
	ContainsOrderNumber []string
	OrderStatus        map[string]string // orderNumber -> status
}

// OrderCount asserts the number of orders
func (a *KitchenDashboardAssertion) OrderCount(count int) *KitchenDashboardAssertion {
	a.ExpectedOrderCount = &count
	return a
}

// ContainsOrder asserts that the dashboard contains an order
func (a *KitchenDashboardAssertion) ContainsOrder(orderNumber string) *KitchenDashboardAssertion {
	a.ContainsOrderNumber = append(a.ContainsOrderNumber, orderNumber)
	return a
}

// OrderWithStatus asserts an order has a specific status
func (a *KitchenDashboardAssertion) OrderWithStatus(orderNumber, status string) *KitchenDashboardAssertion {
	if a.OrderStatus == nil {
		a.OrderStatus = make(map[string]string)
	}
	a.OrderStatus[orderNumber] = status
	return a
}

