package dsl

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/require"
)

// Step represents an executable step in a test scenario
type Step interface {
	Execute(t *testing.T, app *TestApplication)
}

// AddToCartStep represents adding an item to cart
type AddToCartStep struct {
	SessionID string
	ItemID    string
	Quantity  int
}

func (s *AddToCartStep) Execute(t *testing.T, app *TestApplication) {
	item, err := app.menuItemRepo.FindByID(domain.ItemID(s.ItemID))
	require.NoError(t, err)
	require.NotNil(t, item)

	err = app.cartService.AddItem(s.SessionID, item, s.Quantity)
	require.NoError(t, err)
}

// CreateOrderStep represents creating an order
type CreateOrderStep struct {
	SessionID string
}

func (s *CreateOrderStep) Execute(t *testing.T, app *TestApplication) {
	// Create order via HTTP client (POST request)
	req, err := http.NewRequest("POST", app.baseURL+"/order/create", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "bitmerchant_session",
		Value: s.SessionID,
	})

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, resp.StatusCode)
	resp.Body.Close()

	// Extract order ID and number from repository
	restaurantID := domain.RestaurantID("restaurant_1")
	orders, err := app.orderRepo.FindByRestaurantID(restaurantID)
	require.NoError(t, err)
	require.NotEmpty(t, orders, "Order should be created")

	// Store the most recent order (last one)
	latestOrder := orders[len(orders)-1]
	if app.context != nil {
		app.context.SetCreatedOrder(latestOrder.ID, latestOrder.OrderNumber)
	}

	// Give time for event processing and DOM updates via Datastar
	time.Sleep(500 * time.Millisecond)
}

// ViewOrderStep represents viewing an order
type ViewOrderStep struct {
	SessionID   string
	OrderNumber string
}

func (s *ViewOrderStep) Execute(t *testing.T, app *TestApplication) {
	orderNumber := s.OrderNumber
	if orderNumber == "" && app.context != nil {
		orderNumber = string(app.context.GetCreatedOrderNumber())
	}
	require.NotEmpty(t, orderNumber, "OrderNumber must be provided")

	// Navigate to order page via browser
	app.NavigateTo("/order/" + orderNumber)

	// Set session cookie
	app.SetCookie("bitmerchant_session", s.SessionID)

	// Reload to apply cookie
	app.ReloadPage()
}

// ViewKitchenDashboardStep represents viewing kitchen dashboard
type ViewKitchenDashboardStep struct{}

func (s *ViewKitchenDashboardStep) Execute(t *testing.T, app *TestApplication) {
	// Navigate to kitchen dashboard
	app.NavigateTo("/kitchen")

	// Wait for Datastar to initialize and connect to SSE
	app.WaitForPageStable(10 * time.Second)
	time.Sleep(1 * time.Second) // Give Datastar time to initialize

	// Setup DOM monitoring for SSE updates
	app.SetupDOMObserver("orders-list")
}

// MarkOrderPaidStep represents marking an order as paid
type MarkOrderPaidStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderPaidStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	// Make POST request via HTTP client
	req, err := http.NewRequest("POST", app.baseURL+"/kitchen/order/"+orderID+"/mark-paid", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Give time for event processing and DOM updates
	time.Sleep(500 * time.Millisecond)
}

// MarkOrderPreparingStep represents marking an order as preparing
type MarkOrderPreparingStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderPreparingStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	// Make POST request via HTTP client
	req, err := http.NewRequest("POST", app.baseURL+"/kitchen/order/"+orderID+"/mark-preparing", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Give time for event processing and DOM updates
	time.Sleep(500 * time.Millisecond)
}

// MarkOrderReadyStep represents marking an order as ready
type MarkOrderReadyStep struct {
	OrderID string // If empty, uses context's created order ID
}

func (s *MarkOrderReadyStep) Execute(t *testing.T, app *TestApplication) {
	orderID := s.OrderID
	if orderID == "" && app.context != nil {
		orderID = string(app.context.GetCreatedOrderID())
	}
	require.NotEmpty(t, orderID, "OrderID must be provided or order must be created first")

	// Make POST request via HTTP client
	req, err := http.NewRequest("POST", app.baseURL+"/kitchen/order/"+orderID+"/mark-ready", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Give time for event processing and DOM updates
	time.Sleep(500 * time.Millisecond)
}

// PublishEventStep represents publishing a domain event
type PublishEventStep struct {
	Event interface{}
}

func (s *PublishEventStep) Execute(t *testing.T, app *TestApplication) {
	ctx := context.Background()
	var topic string

	switch s.Event.(type) {
	case domain.OrderCreated:
		topic = domain.EventOrderCreated
	case domain.OrderPaid:
		topic = domain.EventOrderPaid
	case domain.OrderPreparing:
		topic = domain.EventOrderPreparing
	case domain.OrderReady:
		topic = domain.EventOrderReady
	default:
		t.Fatalf("Unknown event type: %T", s.Event)
	}

	err := app.eventBus.Publish(ctx, topic, s.Event)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

// SSEStep represents connecting to an SSE stream
// With Rod, SSE connections are handled automatically by Datastar
// We navigate to the page and Datastar connects to the SSE stream via data-init
type SSEStep struct {
	Path string
}

func (s *SSEStep) Execute(t *testing.T, app *TestApplication) {
	// Resolve order number placeholder if path contains "/order/" and context has order
	resolvedPath := s.Path
	if app.context != nil && s.Path != "" {
		// If path contains order number placeholder, resolve it
		if app.context.GetCreatedOrderNumber() != "" && (s.Path == "/order/0001/stream" || strings.Contains(s.Path, "0001")) {
			// Replace placeholder with actual order number
			resolvedPath = "/order/" + string(app.context.GetCreatedOrderNumber()) + "/stream"
		}
	}

	// Determine the page to navigate to based on the SSE stream path
	var pagePath string
	if resolvedPath == "/kitchen/stream" {
		pagePath = "/kitchen"
	} else if strings.HasPrefix(resolvedPath, "/order/") && strings.HasSuffix(resolvedPath, "/stream") {
		// Extract order number from path
		parts := strings.Split(resolvedPath, "/")
		if len(parts) >= 3 {
			orderNumber := parts[2]
			pagePath = "/order/" + orderNumber
		}
	}

	if pagePath != "" {
		// Navigate to the page - Datastar will automatically connect to SSE via data-init
		app.NavigateTo(pagePath)

		// Wait for Datastar to initialize and connect to SSE
		app.WaitForPageStable(10 * time.Second)
		time.Sleep(1 * time.Second) // Give Datastar time to initialize

		// Setup DOM observer for SSE updates
		if pagePath == "/kitchen" {
			app.SetupDOMObserver("orders-list")
		} else if strings.HasPrefix(pagePath, "/order/") {
			// Setup observer for order status updates
			app.GetPage().MustEval(`
				() => {
					window.__orderStatusUpdates = [];
					window.__orderStatusChanged = false;
					
					const observer = new MutationObserver((mutations) => {
						window.__orderStatusChanged = true;
						window.__orderStatusUpdates.push({
							timestamp: Date.now()
						});
					});
					
					const body = document.body;
					if (body) {
						observer.observe(body, { childList: true, subtree: true, characterData: true });
					}
				}
			`)
		}
	}
}
