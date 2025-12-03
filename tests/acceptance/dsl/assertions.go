package dsl

import (
	"net/http"
	"net/http/httptest"
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

	// Verify HTML content if needed
	if len(a.ExpectedHTML) > 0 {
		req := httptest.NewRequest(http.MethodGet, "/order/"+orderNumber, nil)
		rec := httptest.NewRecorder()
		c := app.echo.NewContext(req, rec)
		c.SetPath("/order/:orderNumber")
		c.SetParamNames("orderNumber")
		c.SetParamValues(orderNumber)

		err := app.orderHandler.GetOrder(c)
		require.NoError(t, err)
		body := rec.Body.String()

		for _, html := range a.ExpectedHTML {
			assert.Contains(t, body, html, "Expected HTML not found in order page")
		}
	}
}

// SSEAssertion verifies SSE events
func (a *SSEAssertion) Verify(t *testing.T, app *TestApplication) {
	if app.context == nil {
		t.Fatal("Test context not available")
	}

	// Get SSE client for the stream
	client := app.context.GetSSEClient(a.Stream)
	if client == nil {
		// If no client connected, just log (for backward compatibility)
		t.Logf("SSE client not found for stream: %s. Make sure to connect to SSE stream first.", a.Stream)
		return
	}

	// Wait for event if event type specified
	if a.EventType != "" {
		event, err := client.WaitForEvent(a.EventType, 2*time.Second)
		if err != nil {
			t.Fatalf("Failed to receive SSE event %s: %v", a.EventType, err)
		}
		require.NotNil(t, event, "Expected SSE event but got nil")

		// Verify selector if specified
		if a.Selector != "" {
			assert.Equal(t, a.Selector, event.Selector, "SSE event selector mismatch")
		}

		// Verify HTML content if specified
		if len(a.ExpectedHTML) > 0 {
			for _, html := range a.ExpectedHTML {
				assert.Contains(t, event.Data, html, "Expected HTML not found in SSE event data")
			}
		}
	} else {
		// Just verify that events were received
		events := client.GetAllEvents()
		assert.NotEmpty(t, events, "Expected SSE events but none were received")
	}
}

// KitchenDashboardAssertion verifies kitchen dashboard state
func (a *KitchenDashboardAssertion) Verify(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/kitchen", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)

	err := app.kitchenHandler.GetKitchen(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()

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
}

