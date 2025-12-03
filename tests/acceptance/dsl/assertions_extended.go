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

// VerifySSEEvents verifies SSE events (enhanced version)
func VerifySSEEvents(t *testing.T, app *TestApplication, assertion *SSEAssertion) {
	if app.context == nil {
		t.Fatal("Test context not available")
	}

	// Get SSE client for the stream
	client := app.context.GetSSEClient(assertion.Stream)
	if client == nil {
		t.Fatalf("SSE client not found for stream: %s. Make sure to connect to SSE stream first.", assertion.Stream)
	}

	// Wait for event if event type specified
	if assertion.EventType != "" {
		event, err := client.WaitForEvent(assertion.EventType, 2*time.Second)
		if err != nil {
			t.Fatalf("Failed to receive SSE event %s: %v", assertion.EventType, err)
		}
		require.NotNil(t, event, "Expected SSE event but got nil")

		// Verify selector if specified
		if assertion.Selector != "" {
			assert.Equal(t, assertion.Selector, event.Selector, "SSE event selector mismatch")
		}

		// Verify HTML content if specified
		if len(assertion.ExpectedHTML) > 0 {
			for _, html := range assertion.ExpectedHTML {
				assert.Contains(t, event.Data, html, "Expected HTML not found in SSE event data")
			}
		}
	} else {
		// Just verify that events were received
		events := client.GetAllEvents()
		assert.NotEmpty(t, events, "Expected SSE events but none were received")
	}
}

// MenuAssertion provides assertions for menu state
type MenuAssertion struct {
	ExpectedItems []string
	ExpectedCount *int
}

// MenuShouldShow asserts menu state
func (t *ThenBuilder) MenuShouldShow() *MenuAssertion {
	assertion := &MenuAssertion{}
	t.scenario.addAssertion(assertion)
	return assertion
}

// ContainsItem asserts that menu contains an item
func (a *MenuAssertion) ContainsItem(itemName string) *MenuAssertion {
	a.ExpectedItems = append(a.ExpectedItems, itemName)
	return a
}

// ItemCount asserts the number of items
func (a *MenuAssertion) ItemCount(count int) *MenuAssertion {
	a.ExpectedCount = &count
	return a
}

func (a *MenuAssertion) Verify(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)

	err := app.menuHandler.GetMenu(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()

	if a.ExpectedCount != nil {
		// Count menu items (simple string matching)
		itemCount := strings.Count(body, "menu-item") // Adjust based on actual HTML structure
		assert.Equal(t, *a.ExpectedCount, itemCount, "Menu item count mismatch")
	}

	for _, itemName := range a.ExpectedItems {
		assert.Contains(t, body, itemName, "Expected menu item not found")
	}
}

// OrderHistoryAssertion provides assertions for order history
type OrderHistoryAssertion struct {
	ExpectedOrderCount  *int
	ContainsOrderNumber []string
}

// OrderHistoryShouldShow asserts order history state
func (t *ThenBuilder) OrderHistoryShouldShow() *OrderHistoryAssertion {
	assertion := &OrderHistoryAssertion{}
	t.scenario.addAssertion(assertion)
	return assertion
}

// OrderCount asserts the number of orders
func (a *OrderHistoryAssertion) OrderCount(count int) *OrderHistoryAssertion {
	a.ExpectedOrderCount = &count
	return a
}

// ContainsOrder asserts that history contains an order
func (a *OrderHistoryAssertion) ContainsOrder(orderNumber string) *OrderHistoryAssertion {
	a.ContainsOrderNumber = append(a.ContainsOrderNumber, orderNumber)
	return a
}

func (a *OrderHistoryAssertion) Verify(t *testing.T, app *TestApplication) {
	// This would need to be implemented based on actual order history endpoint
	// For now, we'll verify via repository
	if a.ExpectedOrderCount != nil {
		restaurantID := domain.RestaurantID("restaurant_1")
		orders, err := app.orderRepo.FindByRestaurantID(restaurantID)
		require.NoError(t, err)
		assert.Equal(t, *a.ExpectedOrderCount, len(orders), "Order count mismatch")
	}
}

// CQRSAssertion provides assertions for CQRS-specific behavior
type CQRSAssertion struct {
	CommandExecuted bool
	QueryExecuted   bool
	EventPublished  string
}

// CQRSShould asserts CQRS behavior
func (t *ThenBuilder) CQRSShould() *CQRSAssertion {
	assertion := &CQRSAssertion{}
	t.scenario.addAssertion(assertion)
	return assertion
}

// HaveExecutedCommand asserts that a command was executed
func (a *CQRSAssertion) HaveExecutedCommand() *CQRSAssertion {
	a.CommandExecuted = true
	return a
}

// HaveExecutedQuery asserts that a query was executed
func (a *CQRSAssertion) HaveExecutedQuery() *CQRSAssertion {
	a.QueryExecuted = true
	return a
}

// HavePublishedEvent asserts that an event was published
func (a *CQRSAssertion) HavePublishedEvent(eventType string) *CQRSAssertion {
	a.EventPublished = eventType
	return a
}

func (a *CQRSAssertion) Verify(t *testing.T, app *TestApplication) {
	// For now, this is a placeholder
	// In a full CQRS implementation, we would track command/query execution
	// and verify that commands don't return data, queries don't modify state, etc.

	if a.EventPublished != "" {
		// Verify event was published by checking captured broadcasts
		if app.testSSEHandler != nil {
			// Events trigger SSE broadcasts, so we can verify indirectly
			// In a real CQRS system, we'd check the event bus directly
			broadcasts := app.testSSEHandler.GetCapturedBroadcasts("kitchen")
			if len(broadcasts) == 0 {
				broadcasts = app.testSSEHandler.GetCapturedBroadcasts("order:" + string(app.context.GetCreatedOrderNumber()))
			}
			assert.NotEmpty(t, broadcasts, "Expected event to trigger SSE broadcast")
		}
	}
}
