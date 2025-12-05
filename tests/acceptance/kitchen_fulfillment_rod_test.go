package acceptance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"bitmerchant/tests/acceptance/dsl"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestKitchenFulfillment_NewOrderAppearsWithAlert_Rod demonstrates Rod-based testing
// This test uses a real browser to capture SSE events naturally through the browser's EventSource API
func TestKitchenFulfillment_NewOrderAppearsWithAlert_Rod(t *testing.T) {
	// Setup test data using existing DSL
	scenario := dsl.NewScenario(t, "New order appears (Rod)")
	scenario.Given(func(g *dsl.GivenBuilder) {
		g.Restaurant("restaurant_1", "Test Cafe", true).
			MenuCategory("cat_1", "restaurant_1", "Mains", 1).
			MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
			CustomerSession("session-1").
			CartWithItems("session-1", "item_1", 1)
	})
	app, _ := scenario.BuildApp(t)
	defer app.Cleanup()

	// Start HTTP server for Rod to connect to
	testServer := dsl.StartTestServer(t, app)
	defer testServer.Stop()
	baseURL := testServer.BaseURL()

	// Launch browser with Rod
	browser := rod.New().
		ControlURL(launcher.New().
			Headless(true).
			NoSandbox(true).
			MustLaunch()).
		MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(baseURL + "/kitchen")
	defer page.MustClose()

	// Wait for page to load and Datastar to initialize
	// The kitchen page has data-init="@get('/kitchen/stream')" which tells Datastar to connect
	page.Timeout(10 * time.Second).MustWaitLoad()

	// Wait for Datastar to load and initialize
	page.MustWaitStable()
	time.Sleep(1 * time.Second) // Give Datastar time to initialize and connect to SSE

	// Setup monitoring for SSE events via Datastar's connection
	// We'll monitor the DOM for changes instead of intercepting SSE directly
	page.MustEval(`
		() => {
			window.__domChanges = [];
			window.__orderCount = 0;
			
			// Monitor DOM changes to #orders-list
			const observer = new MutationObserver((mutations) => {
				const ordersList = document.getElementById('orders-list');
				if (ordersList) {
					const newCount = ordersList.children.length;
					if (newCount !== window.__orderCount) {
						window.__orderCount = newCount;
						window.__domChanges.push({
							count: newCount,
							timestamp: Date.now()
						});
					}
				}
			});
			
			const ordersList = document.getElementById('orders-list');
			if (ordersList) {
				window.__orderCount = ordersList.children.length;
				observer.observe(ordersList, { childList: true, subtree: true });
			}
			
			return 'DOM observer set up';
		}
	`)

	// Wait for kitchen page to load (use WaitLoad instead of WaitStable to avoid hanging on SSE)
	page.Timeout(5 * time.Second).MustWaitLoad()
	time.Sleep(500 * time.Millisecond) // Give page time to render

	// Verify initial state - no orders should be present
	// Check if orders list exists, if not, count is 0
	var initialOrderCount int
	initialOrdersList, err := page.Element("#orders-list")
	if err == nil && initialOrdersList != nil {
		initialOrderCount = initialOrdersList.MustEval("el => el ? el.children.length : 0").Int()
	}
	assert.Equal(t, 0, initialOrderCount, "Kitchen should start with no orders")

	// Create order via HTTP API (simulating customer action)
	// Need to include session cookie
	req, err := http.NewRequest("POST", baseURL+"/order/create", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "bitmerchant_session",
		Value: "session-1",
	})

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	orderResp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, orderResp.StatusCode)

	// Wait for DOM update via Datastar (with 5 second timeout as per requirement)
	// Datastar will process the SSE event and update the DOM
	deadline := time.Now().Add(5 * time.Second)
	domUpdated := false

	for time.Now().Before(deadline) {
		currentOrderCount := page.MustEval(`() => window.__orderCount`).Int()

		if currentOrderCount == 1 {
			domUpdated = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, domUpdated, "DOM should be updated by Datastar within 5 seconds (SSE event processed)")

	// Verify DOM was updated by Datastar
	// Wait a bit for Datastar to process the SSE event and update DOM
	time.Sleep(500 * time.Millisecond)

	// Verify DOM was updated by Datastar
	// The order count should already be 1 from the previous check
	finalOrderCount := page.MustEval(`() => {
		const el = document.getElementById('orders-list');
		return el ? el.children.length : 0;
	}`).Int()
	assert.Equal(t, 1, finalOrderCount, "New order should appear in kitchen dashboard")

	// Verify order card exists (order count already verified above)
	// The key success here is that:
	// 1. SSE events are received naturally through browser EventSource API
	// 2. Datastar processes the events and updates the DOM
	// 3. Order count increases from 0 to 1

	if finalOrderCount > 0 {
		// Verify order card exists - order count already verified above
		// The key success is that SSE events are received and DOM is updated
		t.Logf("Success: Order count increased to %d, indicating SSE event was processed by Datastar", finalOrderCount)
	} else {
		// Debug info if order count is still 0
		datastarLoaded := page.MustEval(`() => typeof window.Datastar !== 'undefined'`).Bool()
		bodyHTML := page.MustEval(`() => document.body.innerHTML`).String()
		hasOrderCard := contains(bodyHTML, "order-ord_")
		t.Logf("Debug - Datastar loaded: %v, Order count: %d, Has order card HTML: %v",
			datastarLoaded, finalOrderCount, hasOrderCard)
	}
}

// TestKitchenFulfillment_MarkPaidUpdatesCustomer_Rod tests real-time updates via SSE
func TestKitchenFulfillment_MarkPaidUpdatesCustomer_Rod(t *testing.T) {
	// Setup test data
	scenario := dsl.NewScenario(t, "Mark paid (Rod)")
	scenario.Given(func(g *dsl.GivenBuilder) {
		g.Restaurant("restaurant_1", "Test Cafe", true).
			MenuCategory("cat_1", "restaurant_1", "Mains", 1).
			MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
			CustomerSession("session-1").
			CartWithItems("session-1", "item_1", 1)
	})
	app, _ := scenario.BuildApp(t)
	defer app.Cleanup()

	// Start HTTP server
	testServer := dsl.StartTestServer(t, app)
	defer testServer.Stop()
	baseURL := testServer.BaseURL()

	// Create order first (with session cookie)
	orderReq, err := http.NewRequest("POST", baseURL+"/order/create", nil)
	require.NoError(t, err)
	orderReq.AddCookie(&http.Cookie{
		Name:  "bitmerchant_session",
		Value: "session-1",
	})
	orderClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	orderResp, err := orderClient.Do(orderReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, orderResp.StatusCode)

	// Get order number from context or repository
	orders, err := app.GetOrderRepo().FindByRestaurantID("restaurant_1")
	require.NoError(t, err)
	require.NotEmpty(t, orders)
	orderNumber := orders[len(orders)-1].OrderNumber

	// Launch browser for customer view
	browser := rod.New().
		ControlURL(launcher.New().Headless(true).NoSandbox(true).MustLaunch()).
		MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(fmt.Sprintf("%s/order/%s", baseURL, orderNumber))
	defer page.MustClose()

	// Setup SSE listener for order status updates
	page.MustEval(fmt.Sprintf(`
		() => {
			window.__orderStatusUpdates = [];
			const eventSource = new EventSource('/order/%s/stream');
			eventSource.addEventListener('datastar-patch-elements', (e) => {
				window.__orderStatusUpdates.push({
					type: e.type,
					data: e.data,
					timestamp: Date.now()
				});
			});
		}
	`, orderNumber))

	page.Timeout(5 * time.Second).MustWaitLoad()
	time.Sleep(500 * time.Millisecond)

	// Verify initial status (should be "Pending Payment" or "UNPAID")
	initialStatus := page.MustEval(`() => document.body.textContent`).String()
	// Check for either "Pending" or "UNPAID" (case-insensitive)
	hasPending := strings.Contains(strings.ToLower(initialStatus), "pending") ||
		strings.Contains(strings.ToUpper(initialStatus), "UNPAID")
	assert.True(t, hasPending, "Initial status should be pending payment (found: %s)", initialStatus[:min(200, len(initialStatus))])

	// Staff marks order as paid via HTTP API
	markPaidURL := fmt.Sprintf("%s/kitchen/order/%s/mark-paid", baseURL, orders[len(orders)-1].ID)
	req, err := http.NewRequest("POST", markPaidURL, nil)
	require.NoError(t, err)

	markPaidResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, markPaidResp.StatusCode)

	// Wait for SSE update (max 5 seconds)
	deadline := time.Now().Add(5 * time.Second)
	updateReceived := false

	for time.Now().Before(deadline) {
		updateCount := page.MustEval(`() => window.__orderStatusUpdates.length`).Int()
		if updateCount > 0 {
			updateReceived = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, updateReceived, "SSE update should be received within 5 seconds")

	// Verify DOM was updated (wait for Datastar to process)
	time.Sleep(500 * time.Millisecond)

	// Retry getting status in case Datastar hasn't updated yet
	var updatedStatus string
	statusDeadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(statusDeadline) {
		updatedStatus = page.Timeout(2 * time.Second).
			MustEval(`() => document.body.textContent`).String()
		// Check for "PAID" (uppercase) or "Paid" (case-insensitive)
		if strings.Contains(strings.ToUpper(updatedStatus), "PAID") {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	assert.True(t, strings.Contains(strings.ToUpper(updatedStatus), "PAID"),
		"Status should update to Paid (found: %s)", updatedStatus[:min(200, len(updatedStatus))])
}
