package http_test

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	handler "bitmerchant/internal/interfaces/http"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Mock OrderRepo for Contract Tests
type mockKitchenOrderRepo struct {
	orders []*domain.Order
}

func (m *mockKitchenOrderRepo) FindActiveByRestaurantID(id domain.RestaurantID) ([]*domain.Order, error) {
	return m.orders, nil
}

func (m *mockKitchenOrderRepo) FindByID(id domain.OrderID) (*domain.Order, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, nil
}

func (m *mockKitchenOrderRepo) Update(order *domain.Order) error {
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
func (m *mockKitchenOrderRepo) Save(order *domain.Order) error { return nil }
func (m *mockKitchenOrderRepo) FindByOrderNumber(rid domain.RestaurantID, on string) (*domain.Order, error) { return nil, nil }
func (m *mockKitchenOrderRepo) FindByRestaurantID(rid domain.RestaurantID) ([]*domain.Order, error) { return nil, nil }
func (m *mockKitchenOrderRepo) FindBySessionID(sessionID string) ([]*domain.Order, error) { return nil, nil }



// Mock EventBus
type mockKitchenEventBus struct{}
func (m *mockKitchenEventBus) Publish(ctx context.Context, topic string, event interface{}) error { return nil }


func TestKitchenEndpoints(t *testing.T) {
	e := echo.New()
	
	// Setup Mocks
	mockRepo := &mockKitchenOrderRepo{
		orders: []*domain.Order{
			{
				ID:           "order-1",
				OrderNumber:  "101",
				RestaurantID: "rest-1",
				PaymentStatus: domain.PaymentStatusPending,
				FulfillmentStatus: domain.FulfillmentStatusPaid, // Technically invalid combination but ok for init
				TotalAmount: 1000,
				Items: []domain.OrderItem{
					{MenuItemID: "burger-1", Quantity: 1},
				},
				CreatedAt: time.Now(),
			},
		},
	}
	mockBus := &mockKitchenEventBus{}

	// Setup Use Cases
	getOrdersUC := kitchen.NewGetKitchenOrdersUseCase(mockRepo)
	markPaidUC := kitchen.NewMarkOrderPaidUseCase(mockRepo, mockBus)
	markPreparingUC := kitchen.NewMarkOrderPreparingUseCase(mockRepo, mockBus)
	markReadyUC := kitchen.NewMarkOrderReadyUseCase(mockRepo, mockBus)

	// Setup Handler
	h := handler.NewKitchenHandler(getOrdersUC, markPaidUC, markPreparingUC, markReadyUC)

	// Routes
	e.GET("/kitchen", h.GetKitchen)
	e.POST("/kitchen/order/:id/mark-paid", h.MarkPaid)
	e.POST("/kitchen/order/:id/mark-preparing", h.MarkPreparing)
	e.POST("/kitchen/order/:id/mark-ready", h.MarkReady)

	t.Run("GET /kitchen returns orders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/kitchen", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Mock Session/Context if needed (RestaurantID usually comes from session or config)
		// For now, handler might hardcode or read from env. 
		// We'll assume handler handles it.
		
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
			assert.Equal(t, domain.PaymentStatusPaid, order.PaymentStatus)
		}
	})
}

