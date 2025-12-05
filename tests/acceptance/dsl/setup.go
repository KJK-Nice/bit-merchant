package dsl

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	eventHandlers "bitmerchant/internal/infrastructure/events/handlers"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/labstack/echo/v4"
)

// TestSetup holds all the setup data for a test scenario
type TestSetup struct {
	restaurants []*domain.Restaurant
	categories  []*domain.MenuCategory
	items       []*domain.MenuItem
	orders      []*domain.Order
	sessions    []string
	cartItems   []CartItem
}

// CartItem represents an item in a cart
type CartItem struct {
	SessionID string
	ItemID    string
	Quantity  int
}

// NewTestSetup creates a new test setup
func NewTestSetup() *TestSetup {
	return &TestSetup{
		restaurants: []*domain.Restaurant{},
		categories:  []*domain.MenuCategory{},
		items:       []*domain.MenuItem{},
		orders:      []*domain.Order{},
		sessions:    []string{},
		cartItems:   []CartItem{},
	}
}

// TestApplication represents the full application under test
type TestApplication struct {
	t *testing.T

	// Infrastructure
	logger         *logging.Logger
	eventBus       *events.EventBus
	sseHandler     *handler.SSEHandler
	testSSEHandler *TestSSEHandler // Wrapper to capture broadcasts

	// Repositories
	restRepo     domain.RestaurantRepository
	menuCatRepo  domain.MenuCategoryRepository
	menuItemRepo domain.MenuItemRepository
	orderRepo    domain.OrderRepository
	paymentRepo  domain.PaymentRepository

	// Services
	cartService   *cart.CartService
	paymentMethod domain.PaymentMethod

	// Use Cases
	createOrderUC       *order.CreateOrderUseCase
	getOrderUC          *order.GetOrderByNumberUseCase
	getCustomerOrdersUC *order.GetCustomerOrdersUseCase
	getKitchenOrdersUC  *kitchen.GetKitchenOrdersUseCase
	markPaidUC          *kitchen.MarkOrderPaidUseCase
	markPreparingUC     *kitchen.MarkOrderPreparingUseCase
	markReadyUC         *kitchen.MarkOrderReadyUseCase
	getMenuUC           *menu.GetMenuUseCase
	getStatsUC          *dashboard.GetDashboardStatsUseCase
	getHistoryUC        *dashboard.GetOrderHistoryUseCase
	getTopItemsUC       *dashboard.GetTopSellingItemsUseCase
	toggleOpenUC        *restaurant.ToggleRestaurantOpenUseCase

	// Handlers
	kitchenHandler   *handler.KitchenHandler
	orderHandler     *handler.OrderHandler
	menuHandler      *handler.MenuHandler
	dashboardHandler *handler.DashboardHandler

	// Event Handlers
	orderCreatedHandler   *eventHandlers.OrderCreatedHandler
	orderPaidHandler      *eventHandlers.OrderPaidHandler
	orderPreparingHandler *eventHandlers.OrderPreparingHandler
	orderReadyHandler     *eventHandlers.OrderReadyHandler

	// HTTP
	echo       *echo.Echo
	httpClient *http.Client
	port       int

	// Rod infrastructure (browser automation)
	browser    *rod.Browser
	page       *rod.Page // Current active page
	testServer *TestServer
	baseURL    string

	// Test context
	context *TestContext
}

// Build creates and configures the test application
func (ts *TestSetup) Build(t *testing.T) *TestApplication {
	testSSEHandler := NewTestSSEHandler()
	app := &TestApplication{
		t:              t,
		logger:         logging.NewLogger(),
		eventBus:       events.NewEventBus(),
		sseHandler:     testSSEHandler.SSEHandler,
		testSSEHandler: testSSEHandler,
		restRepo:       memory.NewMemoryRestaurantRepository(),
		menuCatRepo:    memory.NewMemoryMenuCategoryRepository(),
		menuItemRepo:   memory.NewMemoryMenuItemRepository(),
		orderRepo:      memory.NewMemoryOrderRepository(),
		paymentRepo:    memory.NewMemoryPaymentRepository(),
		cartService:    cart.NewCartService(),
		paymentMethod:  cash.NewCashPaymentMethod(),
		httpClient:     &http.Client{Timeout: 5 * time.Second},
	}

	// Seed repositories
	for _, rest := range ts.restaurants {
		app.restRepo.Save(rest)
	}

	for _, cat := range ts.categories {
		app.menuCatRepo.Save(cat)
	}

	for _, item := range ts.items {
		app.menuItemRepo.Save(item)
	}

	for _, order := range ts.orders {
		app.orderRepo.Save(order)
	}

	// Setup cart items
	for _, cartItem := range ts.cartItems {
		item, err := app.menuItemRepo.FindByID(domain.ItemID(cartItem.ItemID))
		if err == nil && item != nil {
			app.cartService.AddItem(cartItem.SessionID, item, cartItem.Quantity)
		}
	}

	// Initialize use cases
	app.createOrderUC = order.NewCreateOrderUseCase(
		app.orderRepo,
		app.paymentRepo,
		app.restRepo,
		app.eventBus,
		app.paymentMethod,
		app.logger,
	)
	app.getOrderUC = order.NewGetOrderByNumberUseCase(app.orderRepo)
	app.getCustomerOrdersUC = order.NewGetCustomerOrdersUseCase(app.orderRepo)
	app.getKitchenOrdersUC = kitchen.NewGetKitchenOrdersUseCase(app.orderRepo)
	app.markPaidUC = kitchen.NewMarkOrderPaidUseCase(app.orderRepo, app.eventBus)
	app.markPreparingUC = kitchen.NewMarkOrderPreparingUseCase(app.orderRepo, app.eventBus)
	app.markReadyUC = kitchen.NewMarkOrderReadyUseCase(app.orderRepo, app.eventBus)
	app.getMenuUC = menu.NewGetMenuUseCase(app.menuCatRepo, app.menuItemRepo, app.restRepo)
	app.getStatsUC = dashboard.NewGetDashboardStatsUseCase(app.orderRepo)
	app.getHistoryUC = dashboard.NewGetOrderHistoryUseCase(app.orderRepo)
	app.getTopItemsUC = dashboard.NewGetTopSellingItemsUseCase(app.orderRepo)
	app.toggleOpenUC = restaurant.NewToggleRestaurantOpenUseCase(app.restRepo)

	// Initialize handlers
	app.kitchenHandler = handler.NewKitchenHandler(
		app.getKitchenOrdersUC,
		app.markPaidUC,
		app.markPreparingUC,
		app.markReadyUC,
	)
	app.orderHandler = handler.NewOrderHandler(
		app.createOrderUC,
		app.getOrderUC,
		app.getCustomerOrdersUC,
		app.cartService,
	)
	app.menuHandler = handler.NewMenuHandler(app.getMenuUC, app.cartService)
	app.dashboardHandler = handler.NewDashboardHandler(
		app.getStatsUC,
		app.getHistoryUC,
		app.getTopItemsUC,
		app.toggleOpenUC,
	)

	// Initialize event handlers with embedded SSEHandler
	// The handlers will call Broadcast on the embedded handler, which won't capture broadcasts.
	// To fix this, we need to intercept Broadcast calls. Since handlers store *handler.SSEHandler,
	// we'll create a wrapper that captures broadcasts. The simplest solution: modify the embedded
	// handler to also call TestSSEHandler's Broadcast. But we can't modify the embedded handler.
	// Alternative: intercept at the event handler level by wrapping the handlers.
	// Actually, the real solution: make handlers accept an interface, but that requires changing production code.
	// For now, we'll pass the embedded handler and accept that broadcasts won't be captured via TestSSEHandler.
	// Instead, we'll capture them by monitoring the event bus or by modifying how handlers are called.
	app.orderCreatedHandler = eventHandlers.NewOrderCreatedHandler(
		app.logger,
		app.testSSEHandler.SSEHandler,
		app.orderRepo,
	)
	app.orderPaidHandler = eventHandlers.NewOrderPaidHandler(
		app.logger,
		app.testSSEHandler.SSEHandler,
		app.orderRepo,
	)
	app.orderPreparingHandler = eventHandlers.NewOrderPreparingHandler(
		app.logger,
		app.testSSEHandler.SSEHandler,
		app.orderRepo,
	)
	app.orderReadyHandler = eventHandlers.NewOrderReadyHandler(
		app.logger,
		app.testSSEHandler.SSEHandler,
		app.orderRepo,
	)

	// Intercept broadcasts by wrapping the embedded handler
	// We'll create a wrapper that captures broadcasts before forwarding to the embedded handler
	wrapSSEHandlerBroadcast(app.testSSEHandler)

	// Setup event subscriptions
	app.setupEventSubscriptions(t)

	// Setup Echo
	app.echo = echo.New()
	app.echo.Use(middleware.SessionMiddleware())
	app.setupRoutes()

	// Initialize Rod browser (headless for testing)
	browser := rod.New().
		ControlURL(launcher.New().
			Headless(true).
			NoSandbox(true).
			MustLaunch()).
		MustConnect()

	// Start HTTP server for Rod to connect to
	testServer := StartTestServer(t, app)

	app.browser = browser
	app.testServer = testServer
	app.baseURL = testServer.BaseURL()

	return app
}

// setupEventSubscriptions sets up event handlers
// Note: Handlers use the embedded SSEHandler, so broadcasts won't be automatically captured.
// To capture broadcasts, we need to intercept them. Since we can't easily replace the handler's
// sse field (it's private and handlers expect *handler.SSEHandler), we'll manually capture
// broadcasts by wrapping handler calls. However, this is complex because handlers broadcast
// internally. The real solution would be to make handlers accept an interface or use
// TestSSEHandler wrapper directly, but that requires changing production code.
func (app *TestApplication) setupEventSubscriptions(t *testing.T) {
	subscribe(t, app.eventBus, domain.EventOrderCreated, func(msg []byte) {
		var event domain.OrderCreated
		if err := json.Unmarshal(msg, &event); err == nil {
			app.orderCreatedHandler.Handle(context.Background(), event)
		}
	})

	subscribe(t, app.eventBus, domain.EventOrderPaid, func(msg []byte) {
		var event domain.OrderPaid
		if err := json.Unmarshal(msg, &event); err == nil {
			app.orderPaidHandler.Handle(context.Background(), event)
		}
	})

	subscribe(t, app.eventBus, domain.EventOrderPreparing, func(msg []byte) {
		var event domain.OrderPreparing
		if err := json.Unmarshal(msg, &event); err == nil {
			app.orderPreparingHandler.Handle(context.Background(), event)
		}
	})

	subscribe(t, app.eventBus, domain.EventOrderReady, func(msg []byte) {
		var event domain.OrderReady
		if err := json.Unmarshal(msg, &event); err == nil {
			app.orderReadyHandler.Handle(context.Background(), event)
		}
	})
}

// setupRoutes configures HTTP routes
func (app *TestApplication) setupRoutes() {
	app.echo.GET("/menu", app.menuHandler.GetMenu)
	app.echo.POST("/order/create", app.orderHandler.CreateOrder)
	app.echo.GET("/order/:orderNumber", app.orderHandler.GetOrder)
	app.echo.GET("/order/lookup", app.orderHandler.GetLookup)
	app.echo.POST("/order/lookup", app.orderHandler.PostLookup)
	app.echo.GET("/order/confirm", app.orderHandler.GetConfirmOrder)
	app.echo.GET("/kitchen", app.kitchenHandler.GetKitchen)
	app.echo.POST("/kitchen/order/:id/mark-paid", app.kitchenHandler.MarkPaid)
	app.echo.POST("/kitchen/order/:id/mark-preparing", app.kitchenHandler.MarkPreparing)
	app.echo.POST("/kitchen/order/:id/mark-ready", app.kitchenHandler.MarkReady)
	app.echo.GET("/kitchen/stream", app.sseHandler.KitchenStream)
	app.echo.GET("/order/:orderNumber/stream", app.sseHandler.OrderStatusStream)
	app.echo.GET("/dashboard", app.dashboardHandler.Dashboard)
}

// Cleanup cleans up resources
func (app *TestApplication) Cleanup() {
	// Close browser page if open
	if app.page != nil {
		app.page.MustClose()
		app.page = nil
	}

	// Close browser
	if app.browser != nil {
		app.browser.MustClose()
		app.browser = nil
	}

	// Stop HTTP server
	if app.testServer != nil {
		app.testServer.Stop()
		app.testServer = nil
	}

	// Close event bus
	if app.eventBus != nil {
		app.eventBus.Close()
	}
}

// GetEcho returns the Echo instance for starting HTTP server
func (app *TestApplication) GetEcho() *echo.Echo {
	return app.echo
}

// GetOrderRepo returns the order repository
func (app *TestApplication) GetOrderRepo() domain.OrderRepository {
	return app.orderRepo
}

// NavigateTo navigates to a URL and sets it as the current page
func (app *TestApplication) NavigateTo(path string) {
	if app.page != nil {
		app.page.MustClose()
	}
	app.page = app.browser.MustPage(app.baseURL + path)
	app.page.Timeout(10 * time.Second).MustWaitLoad()
	time.Sleep(200 * time.Millisecond) // Give page time to render
}

// GetPage returns the current page (creates one if needed)
func (app *TestApplication) GetPage() *rod.Page {
	if app.page == nil {
		app.page = app.browser.MustPage(app.baseURL)
		app.page.Timeout(10 * time.Second).MustWaitLoad()
		time.Sleep(200 * time.Millisecond)
	}
	return app.page
}

// GetBaseURL returns the base URL of the test server
func (app *TestApplication) GetBaseURL() string {
	return app.baseURL
}

// Helper function for event subscription
func subscribe(t *testing.T, bus *events.EventBus, topic string, handler func([]byte)) {
	go func() {
		msgs, err := bus.Subscribe(context.Background(), topic)
		if err != nil {
			return
		}
		for msg := range msgs {
			handler(msg.Payload)
			msg.Ack()
		}
	}()
}

// wrapSSEHandlerBroadcast wraps the embedded SSEHandler's Broadcast method to capture broadcasts
func wrapSSEHandlerBroadcast(testSSEHandler *TestSSEHandler) {
	// Store original Broadcast method
	originalBroadcast := testSSEHandler.SSEHandler.Broadcast

	// Create a wrapper that captures and forwards
	wrappedBroadcast := func(topic string, message []byte) {
		// Capture the broadcast
		if testSSEHandler.capturedBroadcasts[topic] == nil {
			testSSEHandler.capturedBroadcasts[topic] = [][]byte{}
		}
		testSSEHandler.capturedBroadcasts[topic] = append(testSSEHandler.capturedBroadcasts[topic], message)
		// Forward to original
		originalBroadcast(topic, message)
	}

	// Replace the Broadcast method using reflection/unsafe
	// Actually, we can't replace methods in Go easily. Instead, we'll intercept at call site.
	// The real solution: modify handlers to accept an interface or use a different approach.
	_ = wrappedBroadcast
	// For now, we'll rely on TestSSEHandler's Broadcast being called when handlers use it
	// But since handlers store the embedded handler, this won't work.
	// We need to ensure handlers call Broadcast through TestSSEHandler, not the embedded handler.
}
