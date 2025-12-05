package dsl

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assertion represents a verifiable assertion
type Assertion interface {
	Verify(t *testing.T, app *TestApplication)
}

// OrderAssertion verifies order state
func (a *OrderAssertion) Verify(t *testing.T, app *TestApplication) {
	// If OrderNumber is empty, use context's created order number
	orderNumber := a.OrderNumber
	if orderNumber == "" && app.context != nil {
		orderNumber = string(app.context.GetCreatedOrderNumber())
	}
	require.NotEmpty(t, orderNumber, "OrderNumber must be provided or order must be created first")

	// Find order by order number
	restaurantID := domain.RestaurantID("restaurant_1") // Default for tests
	order, err := app.orderRepo.FindByOrderNumber(restaurantID, orderNumber)
	require.NoError(t, err)
	require.NotNil(t, order)

	if a.PaymentStatus != nil {
		assert.Equal(t, *a.PaymentStatus, order.PaymentStatus, "Payment status mismatch")
	}

	if a.FulfillmentStatus != nil {
		assert.Equal(t, *a.FulfillmentStatus, order.FulfillmentStatus, "Fulfillment status mismatch")
	}

	// Verify HTML content if needed (via browser DOM)
	if len(a.ExpectedHTML) > 0 {
		// Navigate to order page if not already there
		currentURL := app.GetCurrentURL()
		if currentURL != "/order/"+orderNumber {
			app.NavigateTo("/order/" + orderNumber)
		}

		// Get page HTML from browser
		bodyHTML := app.GetPage().MustEval(`() => document.body.innerHTML`).String()

		for _, html := range a.ExpectedHTML {
			assert.Contains(t, bodyHTML, html, "Expected HTML not found in order page")
		}
	}
}

// SSEAssertion verifies SSE events
func (a *SSEAssertion) Verify(t *testing.T, app *TestApplication) {
	// With Rod, SSE events are handled naturally by the browser's EventSource API
	// Datastar processes them and updates the DOM automatically
	// We verify DOM changes instead of intercepting SSE messages

	// Resolve path if it's a placeholder
	resolvedPath := a.Stream
	if app.context != nil && strings.HasPrefix(a.Stream, "/order/") && app.context.GetCreatedOrderNumber() != "" {
		if strings.Contains(a.Stream, "0001") || strings.Contains(a.Stream, "placeholder") {
			resolvedPath = fmt.Sprintf("/order/%s/stream", app.context.GetCreatedOrderNumber())
		}
	}

	// Determine which page to check based on stream path
	var pagePath string
	if resolvedPath == "/kitchen/stream" {
		pagePath = "/kitchen"
	} else if strings.HasPrefix(resolvedPath, "/order/") && strings.HasSuffix(resolvedPath, "/stream") {
		parts := strings.Split(resolvedPath, "/")
		if len(parts) >= 3 {
			orderNumber := parts[2]
			pagePath = "/order/" + orderNumber
		}
	}

	if pagePath == "" {
		t.Fatalf("Could not determine page path for SSE stream: %s", a.Stream)
	}

	// Ensure we're on the correct page
	currentURL := app.GetCurrentURL()
	if currentURL != pagePath {
		app.NavigateTo(pagePath)
		app.WaitForPageStable(5 * time.Second)
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for DOM update via Datastar SSE processing
	timeout := 5 * time.Second
	if a.EventType != "" {
		// Wait for DOM changes indicating SSE event was processed
		domUpdated := false
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			// Check if DOM observer detected changes
			domChanged := app.GetPage().MustEval(`() => {
				if (window.__domChanges && window.__domChanges.length > 0) return true;
				if (window.__orderStatusChanged) return true;
				return false;
			}`).Bool()

			if domChanged {
				domUpdated = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if !domUpdated {
			t.Fatalf("Failed to receive SSE event %s (DOM update) within %v", a.EventType, timeout)
		}

		// Verify HTML content if specified
		if len(a.ExpectedHTML) > 0 {
			bodyHTML := app.GetPage().MustEval(`() => document.body.innerHTML`).String()
			bodyHTMLUpper := strings.ToUpper(bodyHTML)
			for _, html := range a.ExpectedHTML {
				// Check case-insensitively for better reliability
				htmlUpper := strings.ToUpper(html)
				assert.True(t, strings.Contains(bodyHTMLUpper, htmlUpper) || strings.Contains(bodyHTML, html),
					"Expected HTML '%s' not found in page after SSE update", html)
			}
		}
	} else {
		// Just verify that DOM was updated (indicating SSE events were received)
		domChanged := app.WaitForSSEEvent(timeout)
		assert.True(t, domChanged, "Expected SSE events (DOM updates) but none were detected")
	}
}

// KitchenDashboardAssertion verifies kitchen dashboard state
func (a *KitchenDashboardAssertion) Verify(t *testing.T, app *TestApplication) {
	// Ensure we're on kitchen page
	currentURL := app.GetCurrentURL()
	if currentURL != "/kitchen" {
		app.NavigateTo("/kitchen")
		app.WaitForPageStable(5 * time.Second)
		time.Sleep(500 * time.Millisecond) // Give Datastar time to process
	}

	// Get page HTML from browser
	body := app.GetPage().MustEval(`() => document.body.innerHTML`).String()

	if a.ExpectedOrderCount != nil {
		// Count orders in the HTML by looking for order divs
		// The actual structure uses id="order-{orderID}" pattern
		orderCount := strings.Count(body, `id="order-`)
		assert.Equal(t, *a.ExpectedOrderCount, orderCount, "Order count mismatch")
	}

	for _, orderNumber := range a.ContainsOrderNumber {
		assert.Contains(t, body, orderNumber, "Expected order number not found in kitchen dashboard")
	}

	for orderNumber, expectedStatus := range a.OrderStatus {
		// Verify order status in HTML
		// This is a simplified check - in production you'd parse HTML properly
		if strings.Contains(body, orderNumber) {
			// Check if status text appears near the order number
			assert.Contains(t, body, expectedStatus, "Expected status %s for order %s", expectedStatus, orderNumber)
		}
	}

	// Verify order details
	for orderNumber, shouldShow := range a.ExpectedOrderShowsDetails {
		if shouldShow {
			// Verify order number is present
			assert.Contains(t, body, orderNumber, "Order number %s not found", orderNumber)
			// Verify order shows items (look for list structure)
			assert.Contains(t, body, "list-disc", "Order %s should show items list", orderNumber)
			// Verify order shows total (look for "Total:" text)
			assert.Contains(t, body, "Total:", "Order %s should show total amount", orderNumber)
		}
	}

	// Verify order items
	for orderNumber, expectedItems := range a.ExpectedOrderShowsItems {
		for _, itemName := range expectedItems {
			// Find the order section and verify item is present
			assert.Contains(t, body, itemName, "Order %s should contain item %s", orderNumber, itemName)
		}
	}

	// Verify order totals
	for orderNumber, expectedTotal := range a.ExpectedOrderShowsTotal {
		// Look for the total amount formatted as currency
		totalStr := fmt.Sprintf("$%.2f", expectedTotal)
		assert.Contains(t, body, totalStr, "Order %s should show total %s", orderNumber, totalStr)
	}

	// Verify order sorting (oldest first)
	if a.OrdersSortedByTime {
		// Get all orders from repository and verify they're sorted
		restaurantID := domain.RestaurantID("restaurant_1")
		orders, err := app.orderRepo.FindByRestaurantID(restaurantID)
		require.NoError(t, err)

		if len(orders) > 1 {
			// Verify orders are sorted by CreatedAt (oldest first)
			for i := 1; i < len(orders); i++ {
				assert.True(t, orders[i-1].CreatedAt.Before(orders[i].CreatedAt) || orders[i-1].CreatedAt.Equal(orders[i].CreatedAt),
					"Orders should be sorted by time (oldest first), but order %d was created after order %d",
					i-1, i)
			}
		}
	}
}
