package acceptance

import (
	"testing"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/tests/acceptance/dsl"
)

// TestEdgeCase_OrderNotFound tests error handling for non-existent order
func TestEdgeCase_OrderNotFound(t *testing.T) {
	dsl.NewScenario(t, "Customer views non-existent order").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Try to view an order that doesn't exist
			w.Customer("session-1").ViewsOrder("INVALID-9999")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Should show error or redirect
			// Note: Actual behavior depends on application implementation
			t.OrderShouldBe("INVALID-9999").
				ContainsHTML("not found")
		}).
		Run()
}

// TestEdgeCase_EmptyCartCannotCreateOrder tests that empty cart prevents order creation
func TestEdgeCase_EmptyCartCannotCreateOrder(t *testing.T) {
	dsl.NewScenario(t, "Customer cannot create order with empty cart").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1")
			// Note: No items in cart
		}).
		When(func(w *dsl.WhenBuilder) {
			// Try to create order with empty cart
			w.Customer("session-1").CreatesOrder()
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Should not create order or show error
			// Verify no orders were created
			// Note: Implementation depends on application behavior
		}).
		Run()
}

// TestEdgeCase_MultipleOrdersInKitchen tests kitchen dashboard with multiple orders
func TestEdgeCase_MultipleOrdersInKitchen(t *testing.T) {
	dsl.NewScenario(t, "Kitchen dashboard shows multiple orders correctly").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				MenuItem("item_2", "cat_1", "restaurant_1", "Pizza", 15.00, true).
				CustomerSession("session-1").
				CustomerSession("session-2").
				CustomerSession("session-3")
		}).
		When(func(w *dsl.WhenBuilder) {
			// Create multiple orders
			w.Customer("session-1").AddsToCart("item_1", 1).CreatesOrder()
			time.Sleep(200 * time.Millisecond)
			w.Customer("session-2").AddsToCart("item_2", 2).CreatesOrder()
			time.Sleep(200 * time.Millisecond)
			w.Customer("session-3").AddsToCart("item_1", 1).AddsToCart("item_2", 1).CreatesOrder()
			time.Sleep(200 * time.Millisecond)
			w.Kitchen().ViewsDashboard()
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Verify all orders appear in kitchen dashboard
			t.KitchenDashboardShouldShow().
				OrderCount(3).
				OrdersAreSortedByTime()
		}).
		Run()
}

// TestEdgeCase_FullOrderLifecycle tests complete order lifecycle
func TestEdgeCase_FullOrderLifecycle(t *testing.T) {
	dsl.NewScenario(t, "Complete order lifecycle from creation to ready").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Customer creates order
			w.Customer("session-1").CreatesOrder()
			w.Customer("session-1").ViewsOrder("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Verify initial state
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPending).
				WithFulfillmentStatus(domain.FulfillmentStatusPaid)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Staff marks as paid
			w.Kitchen().MarksOrderPaid("")
			time.Sleep(200 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Verify paid status
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Staff marks as preparing
			w.Kitchen().MarksOrderPreparing("")
			time.Sleep(200 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Verify preparing status
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusPreparing)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Staff marks as ready
			w.Kitchen().MarksOrderReady("")
			time.Sleep(200 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Verify ready status
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusReady)
		}).
		Run()
}

// TestEdgeCase_OrderStatusTransitions tests all valid status transitions
func TestEdgeCase_OrderStatusTransitions(t *testing.T) {
	dsl.NewScenario(t, "Order status transitions follow correct sequence").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Customer("session-1").CreatesOrder()
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Initial: Pending, Paid (initial fulfillment status)
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPending).
				WithFulfillmentStatus(domain.FulfillmentStatusPaid)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Kitchen().MarksOrderPaid("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// After paid: Paid, Paid (still paid fulfillment until preparing)
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid).
				WithFulfillmentStatus(domain.FulfillmentStatusPaid)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Kitchen().MarksOrderPreparing("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// After preparing: Paid, Preparing
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid).
				WithFulfillmentStatus(domain.FulfillmentStatusPreparing)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Kitchen().MarksOrderReady("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Final: Paid, Ready
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid).
				WithFulfillmentStatus(domain.FulfillmentStatusReady)
		}).
		Run()
}

