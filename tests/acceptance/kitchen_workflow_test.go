package acceptance

import (
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/tests/acceptance/dsl"
)

func TestKitchenWorkflow_OrderLifecycle(t *testing.T) {
	dsl.NewScenario(t, "Kitchen processes order from creation to ready").
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
			t.KitchenDashboardShouldShow().
				OrderCount(1)
			// Note: Order number is randomly generated, so we can't assert exact number
			// But we can assert that an order exists
		}).
		When(func(w *dsl.WhenBuilder) {
			// Empty string means use the last created order
			w.Kitchen().MarksOrderPaid("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Use empty string to use context's created order number
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Kitchen().MarksOrderPreparing("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusPreparing)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Kitchen().MarksOrderReady("")
		}).
		Then(func(t *dsl.ThenBuilder) {
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusReady)
		}).
		Run()
}

func TestKitchenWorkflow_SimpleOrderCreation(t *testing.T) {
	dsl.NewScenario(t, "Customer creates order and kitchen sees it").
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
			t.KitchenDashboardShouldShow().
				OrderCount(1)
			// Order number is randomly generated, so we just verify count
		}).
		Run()
}

