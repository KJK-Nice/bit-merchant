package order_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
)

func TestCreateOrderUseCase_Execute(t *testing.T) {
	orderRepo := &mockOrderRepository{
		orders: make(map[domain.OrderID]*domain.Order),
	}
	paymentRepo := &mockPaymentRepository{
		payments: map[domain.PaymentID]*domain.Payment{
			"pay_001": {
				ID:             "pay_001",
				RestaurantID:   "rest_001",
				Status:         domain.PaymentStatusPaid,
				InvoiceID:      "inv_001",
				Invoice:        "lnbc123...",
				AmountSatoshis: 1000,
				AmountFiat:     10.00,
				ExchangeRate:   100000,
			},
		},
	}
	itemRepo := &mockMenuItemRepository{
		items: map[domain.ItemID]*domain.MenuItem{
			"item_001": {
				ID:          "item_001",
				Name:        "Test Item",
				Price:       10.00,
				IsAvailable: true,
			},
		},
	}
	eventBus := &mockEventBus{}

	useCase := order.NewCreateOrderUseCase(orderRepo, paymentRepo, itemRepo, eventBus)

	req := order.CreateOrderRequest{
		PaymentID: "pay_001",
		Items: []order.OrderItemRequest{
			{MenuItemID: "item_001", Quantity: 1},
		},
	}

	result, err := useCase.Execute(req)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if result.OrderID == "" {
		t.Error("CreateOrder() returned empty OrderID")
	}
	if result.OrderNumber == "" {
		t.Error("CreateOrder() returned empty OrderNumber")
	}

	// Verify order was saved
	order, err := orderRepo.FindByID(result.OrderID)
	if err != nil {
		t.Fatalf("Order not found in repository: %v", err)
	}
	if order.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("Order PaymentStatus = %v, want %v", order.PaymentStatus, domain.PaymentStatusPaid)
	}
	if order.FulfillmentStatus != domain.FulfillmentStatusPaid {
		t.Errorf("Order FulfillmentStatus = %v, want %v", order.FulfillmentStatus, domain.FulfillmentStatusPaid)
	}
}

func TestCreateOrderUseCase_PaymentNotPaid(t *testing.T) {
	orderRepo := &mockOrderRepository{}
	paymentRepo := &mockPaymentRepository{
		payments: map[domain.PaymentID]*domain.Payment{
			"pay_001": {
				ID:     "pay_001",
				Status: domain.PaymentStatusPending,
			},
		},
	}
	itemRepo := &mockMenuItemRepository{}
	eventBus := &mockEventBus{}

	useCase := order.NewCreateOrderUseCase(orderRepo, paymentRepo, itemRepo, eventBus)

	req := order.CreateOrderRequest{
		PaymentID: "pay_001",
		Items:     []order.OrderItemRequest{},
	}

	_, err := useCase.Execute(req)
	if err == nil {
		t.Error("CreateOrder() with unpaid payment should return error")
	}
}

func TestCreateOrderUseCase_EmptyItems(t *testing.T) {
	orderRepo := &mockOrderRepository{}
	paymentRepo := &mockPaymentRepository{
		payments: map[domain.PaymentID]*domain.Payment{
			"pay_001": {
				ID:     "pay_001",
				Status: domain.PaymentStatusPaid,
			},
		},
	}
	itemRepo := &mockMenuItemRepository{}
	eventBus := &mockEventBus{}

	useCase := order.NewCreateOrderUseCase(orderRepo, paymentRepo, itemRepo, eventBus)

	req := order.CreateOrderRequest{
		PaymentID: "pay_001",
		Items:     []order.OrderItemRequest{},
	}

	_, err := useCase.Execute(req)
	if err == nil {
		t.Error("CreateOrder() with empty items should return error")
	}
}

func TestGetOrderByNumberUseCase_Execute(t *testing.T) {
	orderRepo := &mockOrderRepository{
		orders: map[domain.OrderID]*domain.Order{
			"ord_001": {
				ID:           "ord_001",
				OrderNumber:  "ORD-001",
				RestaurantID: "rest_001",
			},
		},
		orderNumbers: map[string]*domain.Order{
			"rest_001:ORD-001": {
				ID:           "ord_001",
				OrderNumber:  "ORD-001",
				RestaurantID: "rest_001",
			},
		},
	}

	useCase := order.NewGetOrderByNumberUseCase(orderRepo)

	order, err := useCase.Execute("rest_001", "ORD-001")
	if err != nil {
		t.Fatalf("GetOrderByNumber() error = %v", err)
	}
	if order.OrderNumber != "ORD-001" {
		t.Errorf("GetOrderByNumber() OrderNumber = %v, want 'ORD-001'", order.OrderNumber)
	}
}

func TestGetOrderByNumberUseCase_NotFound(t *testing.T) {
	orderRepo := &mockOrderRepository{
		orderNumbers: make(map[string]*domain.Order),
	}

	useCase := order.NewGetOrderByNumberUseCase(orderRepo)

	_, err := useCase.Execute("rest_001", "ORD-999")
	if err == nil {
		t.Error("GetOrderByNumber() with non-existent order should return error")
	}
}

// Mock repositories for testing
type mockOrderRepository struct {
	orders       map[domain.OrderID]*domain.Order
	orderNumbers map[string]*domain.Order
}

func (m *mockOrderRepository) Save(order *domain.Order) error {
	m.orders[order.ID] = order
	if m.orderNumbers != nil {
		m.orderNumbers[string(order.RestaurantID)+":"+string(order.OrderNumber)] = order
	}
	return nil
}

func (m *mockOrderRepository) FindByID(id domain.OrderID) (*domain.Order, error) {
	order, exists := m.orders[id]
	if !exists {
		return nil, &mockError{msg: "order not found"}
	}
	return order, nil
}

func (m *mockOrderRepository) FindByOrderNumber(restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error) {
	order, exists := m.orderNumbers[string(restaurantID)+":"+orderNumber]
	if !exists {
		return nil, &mockError{msg: "order not found"}
	}
	return order, nil
}

func (m *mockOrderRepository) FindByRestaurantID(domain.RestaurantID) ([]*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepository) FindActiveByRestaurantID(domain.RestaurantID) ([]*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepository) Update(*domain.Order) error { return nil }

type mockPaymentRepository struct {
	payments map[domain.PaymentID]*domain.Payment
}

func (m *mockPaymentRepository) FindByID(id domain.PaymentID) (*domain.Payment, error) {
	payment, exists := m.payments[id]
	if !exists {
		return nil, &mockError{msg: "payment not found"}
	}
	return payment, nil
}

func (m *mockPaymentRepository) Save(*domain.Payment) error                      { return nil }
func (m *mockPaymentRepository) FindByInvoiceID(string) (*domain.Payment, error) { return nil, nil }
func (m *mockPaymentRepository) FindByRestaurantID(domain.RestaurantID) ([]*domain.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepository) FindPendingSettlements(domain.RestaurantID) ([]*domain.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepository) Update(*domain.Payment) error { return nil }

type mockMenuItemRepository struct {
	items map[domain.ItemID]*domain.MenuItem
}

func (m *mockMenuItemRepository) FindByID(id domain.ItemID) (*domain.MenuItem, error) {
	item, exists := m.items[id]
	if !exists {
		return nil, &mockError{msg: "menu item not found"}
	}
	return item, nil
}

func (m *mockMenuItemRepository) Save(*domain.MenuItem) error { return nil }
func (m *mockMenuItemRepository) FindByCategoryID(domain.CategoryID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) FindByRestaurantID(domain.RestaurantID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) FindAvailableByRestaurantID(domain.RestaurantID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) Update(*domain.MenuItem) error                        { return nil }
func (m *mockMenuItemRepository) Delete(domain.ItemID) error                           { return nil }
func (m *mockMenuItemRepository) CountByRestaurantID(domain.RestaurantID) (int, error) { return 0, nil }

type mockEventBus struct {
	publishedEvents []interface{}
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	if m.publishedEvents == nil {
		m.publishedEvents = make([]interface{}, 0)
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
