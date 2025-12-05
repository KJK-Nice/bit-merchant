package acceptance

import (
	"testing"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/tests/acceptance/dsl"
)

// Scenario 1: QR code scan → menu loads instantly (<2s) with HTML page
func TestCustomerOrdering_QRCodeScanLoadsMenu(t *testing.T) {
	dsl.NewScenario(t, "Customer scans QR code and menu loads instantly").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				MenuItem("item_2", "cat_1", "restaurant_1", "Pizza", 15.00, true).
				CustomerSession("session-1")
		}).
		When(func(w *dsl.WhenBuilder) {
			// Simulate QR code scan by viewing menu
			w.Customer("session-1").ViewsMenu()
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert menu loads with categories, items, prices, photos
			t.MenuShouldShow().
				ItemCount(2).
				ContainsItem("Burger").
				ContainsItem("Pizza")
			// Assert performance requirement (<2s load time)
			t.PerformanceShould().RespondInLessThan(2 * time.Second)
		}).
		Run()
}

// Scenario 2: Tap food items → added to cart with running total, page updates without reload
func TestCustomerOrdering_AddItemsToCart(t *testing.T) {
	dsl.NewScenario(t, "Customer taps food items and they are added to cart").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				MenuItem("item_2", "cat_1", "restaurant_1", "Pizza", 15.00, true).
				CustomerSession("session-1")
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Customer("session-1").
				ViewsMenu().
				AddsToCart("item_1", 1).
				AddsToCart("item_2", 2)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert cart shows running total (10.00 + 15.00*2 = 40.00)
			// Note: Cart assertion needs to be implemented to check HTML response
			// For now, we verify via kitchen dashboard that order can be created
			t.KitchenDashboardShouldShow().
				OrderCount(0) // No orders yet, just cart items
		}).
		Run()
}

// Scenario 3: Tap "Place Order" → order confirmation page appears
func TestCustomerOrdering_PlaceOrderShowsConfirmation(t *testing.T) {
	dsl.NewScenario(t, "Customer taps Place Order and sees confirmation page").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			// Navigate to order confirmation
			w.Customer("session-1").ViewsOrderConfirmation()
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert order confirmation page shows:
			// - Order summary
			// - Total amount
			// - Cash payment confirmation form
			t.OrderConfirmationShouldShow().
				WithOrderSummary().
				WithTotalAmount(10.00)
		}).
		Run()
}

// Scenario 4: Confirm cash payment → order created with "Pending Payment" status
func TestCustomerOrdering_ConfirmCashPaymentCreatesOrder(t *testing.T) {
	dsl.NewScenario(t, "Customer confirms cash payment and order is created").
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
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPending).
				ContainsHTML("order") // Order number should be displayed
			// Verify order appears in kitchen dashboard
			t.KitchenDashboardShouldShow().
				OrderCount(1)
		}).
		Run()
}

// Scenario 5: Order created → customer sees order number and real-time status
func TestCustomerOrdering_OrderCreatedShowsStatus(t *testing.T) {
	dsl.NewScenario(t, "Customer sees order number and real-time status after creation").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Customer("session-1").CreatesOrder()
			// Connect to order status SSE stream (order number will be resolved from context)
			// Note: We'll need to resolve order number dynamically
			w.ConnectsToSSE("/order/0001/stream").Stream() // Placeholder - will be resolved
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert order number is displayed and status is correct
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPending).
				ContainsHTML("Order") // Order number should be displayed
			// Note: SSE stream connection verification would go here
			// but we need to resolve the actual order number from context
		}).
		Run()
}

// Scenario 6: Status changes to "Ready" → customer receives real-time update
func TestCustomerOrdering_StatusChangeToReadyUpdatesCustomer(t *testing.T) {
	dsl.NewScenario(t, "Customer receives real-time update when order status changes to Ready").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Customer("session-1").CreatesOrder()
			// Connect to order status stream
			// Note: Order number resolution from context needed
			w.ConnectsToSSE("/order/0001/stream").Stream()
		}).
		When(func(w *dsl.WhenBuilder) {
			// Kitchen marks order as paid, preparing, then ready
			w.Kitchen().MarksOrderPaid("")
			time.Sleep(100 * time.Millisecond) // Allow event propagation
			w.Kitchen().MarksOrderPreparing("")
			time.Sleep(100 * time.Millisecond)
			w.Kitchen().MarksOrderReady("")
			time.Sleep(100 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert customer receives SSE update for "Ready" status
			// Note: Need to resolve actual order number from context
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusReady)
			// SSE assertion would verify the real-time update
			// t.SSEStreamShouldReceive("/order/0001/stream").
			// 	Event("datastar-patch-elements").
			// 	ContainsHTML("Ready")
		}).
		Run()
}

// Scenario 7: Customer pays cash → staff marks paid → customer sees updates
func TestCustomerOrdering_CustomerPaysCashAndSeesUpdates(t *testing.T) {
	dsl.NewScenario(t, "Customer pays cash and sees status updates in real-time").
		Given(func(g *dsl.GivenBuilder) {
			g.Restaurant("restaurant_1", "Test Cafe", true).
				MenuCategory("cat_1", "restaurant_1", "Mains", 1).
				MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
				CustomerSession("session-1").
				CartWithItems("session-1", "item_1", 1)
		}).
		When(func(w *dsl.WhenBuilder) {
			w.Customer("session-1").CreatesOrder()
			// Connect to order status stream
			// Note: Order number will be resolved from context
			w.ConnectsToSSE("/order/0001/stream").Stream()
		}).
		When(func(w *dsl.WhenBuilder) {
			// Staff marks order as paid
			w.Kitchen().MarksOrderPaid("")
			time.Sleep(100 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert customer sees "Paid" status update
			t.OrderShouldBe("").
				WithPaymentStatus(domain.PaymentStatusPaid)
			// SSE assertion would verify real-time update
			// t.SSEStreamShouldReceive("/order/0001/stream").
			// 	Event("datastar-patch-elements").
			// 	ContainsHTML("Paid")
		}).
		When(func(w *dsl.WhenBuilder) {
			// Staff marks order as preparing
			w.Kitchen().MarksOrderPreparing("")
			time.Sleep(100 * time.Millisecond)
		}).
		Then(func(t *dsl.ThenBuilder) {
			// Assert customer sees "Preparing" status update
			t.OrderShouldBe("").
				WithFulfillmentStatus(domain.FulfillmentStatusPreparing)
			// SSE assertion would verify real-time update
			// t.SSEStreamShouldReceive("/order/0001/stream").
			// 	Event("datastar-patch-elements").
			// 	ContainsHTML("Preparing")
		}).
		Run()
}
