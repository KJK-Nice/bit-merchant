package kitchen_test

import (
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	menuhttp "bitmerchant/internal/menu/ports/http"
	"bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderevent "bitmerchant/internal/ordering/app/event"
	orderQuery "bitmerchant/internal/ordering/app/query"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	ordersse "bitmerchant/internal/ordering/ports/sse"
	placesCmd "bitmerchant/internal/places/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestKitchenWorkflow(t *testing.T) {

	logger := logging.NewLogger()
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	restRepo := memory.NewMemoryRestaurantRepository()
	menuCatRepo := memory.NewMemoryMenuCategoryRepository()
	menuItemRepo := memory.NewMemoryMenuItemRepository()
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()

	// Services
	cartService := cart.NewCartService()
	paymentMethod := cash.NewCashPaymentMethod()
	sseHandler := commonhttp.NewSSEHandler()

	// Seed Data
	restaurantID := common.RestaurantID("restaurant_1")
	rSeed, _ := restaurant.NewRestaurant(restaurantID, "Test Cafe")
	_ = restRepo.Save(rSeed)

	cat1, _ := menu.NewMenuCategory("cat_1", restaurantID, "Mains", 1)
	_ = menuCatRepo.Save(cat1)

	item1, _ := menu.NewMenuItem("item_1", "cat_1", restaurantID, "Burger", 10.00)
	_ = menuItemRepo.Save(item1)

	// Use Cases
	_ = paymentRepo
	_ = paymentMethod
	createOrderUC := orderCmd.NewCreateOrderHandler(orderRepo, restRepo, eventBus, logger.Logger, nil)
	getCustomerOrderUC := orderQuery.NewCustomerOrderByLookupHandler(orderRepo, nil, nil)
	getCustomerOrdersUC := orderQuery.NewCustomerOrdersForSessionHandler(orderRepo, nil, nil)
	getKitchenOrdersUC := orderQuery.NewActiveKitchenOrdersHandler(orderRepo, nil, nil)
	markPaidUC := orderCmd.NewMarkOrderPaidHandler(orderRepo, eventBus, logger.Logger, nil)
	markPreparingUC := orderCmd.NewMarkOrderPreparingHandler(orderRepo, eventBus, logger.Logger, nil)
	markReadyUC := orderCmd.NewMarkOrderReadyHandler(orderRepo, eventBus, logger.Logger, nil)
	markCompletedUC := orderCmd.NewMarkOrderCompletedHandler(orderRepo, eventBus, logger.Logger, nil)
	toggleItemPrepUC := orderCmd.NewToggleOrderItemPrepHandler(orderRepo, eventBus, logger.Logger, nil)
	getMenuUC := menuQuery.NewMenuForCustomerHandler(menuCatRepo, menuItemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{}, nil, nil)

	getUnpaidServerUC := orderQuery.NewUnpaidServerOrdersHandler(orderRepo, nil, nil)

	// Handlers
	kitchenHandler := orderinghttp.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC, markCompletedUC, toggleItemPrepUC, nil, nil, "")
	serverHandler := orderinghttp.NewServerHandler(getUnpaidServerUC, markPaidUC, nil, nil)
	orderHandler := orderinghttp.NewOrderHandler(createOrderUC, getCustomerOrderUC, getCustomerOrdersUC, orderRepo, restRepo, cartService, "")
	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	recordVisitUC := placesCmd.NewRecordMenuVisitHandler(restRepo, visitRepo, nil, nil)
	_ = menuhttp.NewMenuHandler(getMenuUC, cartService, recordVisitUC)

	// Event Handlers
	orderCreatedHandler := ordersse.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := ordersse.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := ordersse.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := ordersse.NewOrderReadyHandler(logger, sseHandler, orderRepo)
	orderCompletedHandler := ordersse.NewOrderCompletedHandler(logger, sseHandler, orderRepo)

	// Subscriptions
	subscribe(t, eventBus, common.EventOrderCreated, func(msg []byte) {
		var event orderevent.OrderCreated
		require.NoError(t, json.Unmarshal(msg, &event))
		require.NoError(t, orderCreatedHandler.Handle(context.Background(), event))
	})
	subscribe(t, eventBus, common.EventOrderPaid, func(msg []byte) {
		var event orderevent.OrderPaid
		require.NoError(t, json.Unmarshal(msg, &event))
		require.NoError(t, orderPaidHandler.Handle(context.Background(), event))
	})
	subscribe(t, eventBus, common.EventOrderPreparing, func(msg []byte) {
		var event orderevent.OrderPreparing
		require.NoError(t, json.Unmarshal(msg, &event))
		require.NoError(t, orderPreparingHandler.Handle(context.Background(), event))
	})
	subscribe(t, eventBus, common.EventOrderReady, func(msg []byte) {
		var event orderevent.OrderReady
		require.NoError(t, json.Unmarshal(msg, &event))
		require.NoError(t, orderReadyHandler.Handle(context.Background(), event))
	})
	subscribe(t, eventBus, common.EventOrderCompleted, func(msg []byte) {
		var event orderevent.OrderCompleted
		require.NoError(t, json.Unmarshal(msg, &event))
		require.NoError(t, orderCompletedHandler.Handle(context.Background(), event))
	})

	// Echo Setup
	e := echo.New()
	e.Use(middleware.SessionMiddleware())

	// Routes
	e.POST("/order/create", orderHandler.CreateOrder)
	e.GET("/kitchen", kitchenHandler.GetKitchen)
	e.POST("/kitchen/order/:id/mark-preparing", kitchenHandler.MarkPreparing)
	e.POST("/kitchen/order/:id/mark-ready", kitchenHandler.MarkReady)
	e.POST("/kitchen/order/:id/mark-completed", kitchenHandler.MarkCompleted)
	e.POST("/server/order/:id/mark-paid", serverHandler.MarkPaid)

	// --- Test Execution ---

	// 1. Customer Creates Order
	sessionID := "session-1"
	_ = cartService.AddItem(sessionID, item1, 1)

	form := "paymentMethod=cash&restaurantID=" + string(restaurantID) + "&customerName=Maya&tipPercent=15"
	req := httptest.NewRequest(http.MethodPost, "/order/create", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.AddCookie(&http.Cookie{Name: "bitmerchant_session", Value: sessionID})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", sessionID)

	err := orderHandler.CreateOrder(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, rec.Code)

	orders, _ := orderRepo.FindByRestaurantID(restaurantID)
	require.Len(t, orders, 1)
	orderID := orders[0].ID
	orderNumber := orders[0].OrderNumber

	time.Sleep(100 * time.Millisecond)

	// 2. Kitchen Views Orders
	// Note: Kitchen handler uses hardcoded "rest-1" in GetKitchen method in previous implementation?
	// Let's check the implementation of GetKitchen.
	// I updated main.go to use "restaurant-1" but did I update kitchen handler logic?
	// The handler implementation in internal/ordering/ports/http/kitchen.go had: restaurantID := common.RestaurantID("rest-1")
	// This mismatch ("rest-1" vs "restaurant-1" in seed/tests) is likely the cause of empty list!

	req = httptest.NewRequest(http.MethodGet, "/kitchen", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(middleware.ContextRestaurantID, restaurantID)

	err = kitchenHandler.GetKitchen(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	if !strings.Contains(rec.Body.String(), string(orderNumber)) {
		t.Logf("Body: %s", rec.Body.String())
		t.Logf("Expected Order Number: %s", orderNumber)

		activeOrders, _ := orderRepo.FindActiveByRestaurantID(restaurantID)
		t.Logf("Active Orders in Repo for %s: %d", restaurantID, len(activeOrders))
	}

	assert.Contains(t, rec.Body.String(), string(orderNumber))
	assert.Contains(t, rec.Body.String(), "UNPAID")
	assert.NotContains(t, rec.Body.String(), "Mark Paid", "kitchen view must not expose Mark Paid")

	// 3. FOH (server) Marks Paid
	req = httptest.NewRequest(http.MethodPost, "/server/order/"+string(orderID)+"/mark-paid", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/server/order/:id/mark-paid")
	c.SetParamNames("id")
	c.SetParamValues(string(orderID))

	err = serverHandler.MarkPaid(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedOrder, _ := orderRepo.FindByID(orderID)
	assert.Equal(t, common.PaymentStatusPaid, updatedOrder.PaymentStatus)

	time.Sleep(100 * time.Millisecond)

	// 4. Kitchen Marks Preparing
	req = httptest.NewRequest(http.MethodPost, "/kitchen/order/"+string(orderID)+"/mark-preparing", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-preparing")
	c.SetParamNames("id")
	c.SetParamValues(string(orderID))

	err = kitchenHandler.MarkPreparing(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedOrder, _ = orderRepo.FindByID(orderID)
	assert.Equal(t, common.FulfillmentStatusPreparing, updatedOrder.FulfillmentStatus)

	// 5. Kitchen Marks Ready
	req = httptest.NewRequest(http.MethodPost, "/kitchen/order/"+string(orderID)+"/mark-ready", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-ready")
	c.SetParamNames("id")
	c.SetParamValues(string(orderID))

	err = kitchenHandler.MarkReady(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedOrder, _ = orderRepo.FindByID(orderID)
	assert.Equal(t, common.FulfillmentStatusReady, updatedOrder.FulfillmentStatus)

	// 6. Kitchen Marks Completed (settled/cleared from active queue)
	req = httptest.NewRequest(http.MethodPost, "/kitchen/order/"+string(orderID)+"/mark-completed", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/kitchen/order/:id/mark-completed")
	c.SetParamNames("id")
	c.SetParamValues(string(orderID))

	err = kitchenHandler.MarkCompleted(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedOrder, _ = orderRepo.FindByID(orderID)
	assert.Equal(t, common.FulfillmentStatusCompleted, updatedOrder.FulfillmentStatus)
}

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
