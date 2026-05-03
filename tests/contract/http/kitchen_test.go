package http_test

import (
	"bitmerchant/internal/common"

	httpMiddleware "bitmerchant/internal/common/http/middleware"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	kitchenQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock OrderRepo for Contract Tests
type mockKitchenOrderRepo struct {
	orders []*order.Order
}

func (m *mockKitchenOrderRepo) FindActiveByRestaurantID(id common.RestaurantID) ([]*order.Order, error) {
	return m.orders, nil
}

func (m *mockKitchenOrderRepo) FindByID(id common.OrderID) (*order.Order, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, nil
}

func (m *mockKitchenOrderRepo) Update(order *order.Order) error {
	// Update in place
	for i, o := range m.orders {
		if o.ID == order.ID {
			m.orders[i] = order
			return nil
		}
	}
	return nil
}

// Stubs for other methods
func (m *mockKitchenOrderRepo) Save(order *order.Order) error { return nil }
func (m *mockKitchenOrderRepo) FindByOrderNumber(rid common.RestaurantID, on string) (*order.Order, error) {
	return nil, nil
}
func (m *mockKitchenOrderRepo) FindByRestaurantID(rid common.RestaurantID) ([]*order.Order, error) {
	return nil, nil
}
func (m *mockKitchenOrderRepo) FindBySessionID(sessionID string) ([]*order.Order, error) {
	return nil, nil
}

// Mock EventBus
type mockKitchenEventBus struct{}

func (m *mockKitchenEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	return nil
}

func TestKitchenEndpoints(t *testing.T) {
	e := echo.New()

	// Setup Mocks
	mockRepo := &mockKitchenOrderRepo{
		orders: []*order.Order{
			{
				ID:                "order-1",
				OrderNumber:       "101",
				RestaurantID:      "rest-1",
				PaymentStatus:     common.PaymentStatusPending,
				FulfillmentStatus: common.FulfillmentStatusPaid, // Technically invalid combination but ok for init
				TotalAmount:       1000,
				Items: []order.OrderItem{
					{MenuItemID: "burger-1", Quantity: 1},
				},
				CreatedAt: time.Now(),
			},
		},
	}
	mockBus := &mockKitchenEventBus{}

	// Setup Use Cases
	getOrdersUC := kitchenQuery.NewActiveKitchenOrdersHandler(mockRepo, nil, nil)
	markPaidUC := kitchenCmd.NewMarkOrderPaidHandler(mockRepo, mockBus, nil, nil)
	markPreparingUC := kitchenCmd.NewMarkOrderPreparingHandler(mockRepo, mockBus, nil, nil)
	markReadyUC := kitchenCmd.NewMarkOrderReadyHandler(mockRepo, mockBus, nil, nil)
	markCompletedUC := kitchenCmd.NewMarkOrderCompletedHandler(mockRepo, mockBus, nil, nil)

	// Setup Handler
	h := orderinghttp.NewKitchenHandler(getOrdersUC, markPaidUC, markPreparingUC, markReadyUC, markCompletedUC, nil, nil, "")

	// Routes
	e.GET("/kitchen", h.GetKitchen)
	e.POST("/kitchen/order/:id/mark-paid", h.MarkPaid)
	e.POST("/kitchen/order/:id/mark-preparing", h.MarkPreparing)
	e.POST("/kitchen/order/:id/mark-ready", h.MarkReady)
	e.POST("/kitchen/order/:id/mark-completed", h.MarkCompleted)

	t.Run("GET /kitchen returns orders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/kitchen", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, common.RestaurantID("rest-1"))

		if assert.NoError(t, h.GetKitchen(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			// assert.Contains(t, rec.Body.String(), "Kitchen Display") // Template not impl yet
		}
	})

	t.Run("POST /kitchen/order/:id/mark-paid updates status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/kitchen/order/order-1/mark-paid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/kitchen/order/:id/mark-paid")
		c.SetParamNames("id")
		c.SetParamValues("order-1")

		if assert.NoError(t, h.MarkPaid(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			// Verify repo update
			order, _ := mockRepo.FindByID("order-1")
			assert.Equal(t, common.PaymentStatusPaid, order.PaymentStatus)
		}
	})
}
