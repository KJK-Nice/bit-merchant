package kitchen_test

import (
	"bitmerchant/internal/domain"
	"context"
	"time"
)

type mockOrderRepo struct {
	saveFn                     func(order *domain.Order) error
	findByIDFn                 func(id domain.OrderID) (*domain.Order, error)
	findByOrderNumberFn        func(restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error)
	findByRestaurantIDFn       func(restaurantID domain.RestaurantID) ([]*domain.Order, error)
	findActiveByRestaurantIDFn func(restaurantID domain.RestaurantID) ([]*domain.Order, error)
	updateFn                   func(order *domain.Order) error
}

func (m *mockOrderRepo) Save(order *domain.Order) error {
	if m.saveFn != nil {
		return m.saveFn(order)
	}
	return nil
}

func (m *mockOrderRepo) FindByID(id domain.OrderID) (*domain.Order, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindByOrderNumber(restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error) {
	if m.findByOrderNumberFn != nil {
		return m.findByOrderNumberFn(restaurantID, orderNumber)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	if m.findByRestaurantIDFn != nil {
		return m.findByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindActiveByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	if m.findActiveByRestaurantIDFn != nil {
		return m.findActiveByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

func (m *mockOrderRepo) Update(order *domain.Order) error {
	if m.updateFn != nil {
		return m.updateFn(order)
	}
	return nil
}

type mockEventBus struct {
	publishFn func(ctx context.Context, topic string, event interface{}) error
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, topic, event)
	}
	return nil
}

type mockPaymentRepo struct {
	saveFn               func(payment *domain.Payment) error
	findByIDFn           func(id domain.PaymentID) (*domain.Payment, error)
	findByOrderIDFn      func(orderID domain.OrderID) (*domain.Payment, error)
	findByRestaurantIDFn func(restaurantID domain.RestaurantID) ([]*domain.Payment, error)
	updateFn             func(payment *domain.Payment) error
}

func (m *mockPaymentRepo) Save(payment *domain.Payment) error {
	if m.saveFn != nil {
		return m.saveFn(payment)
	}
	return nil
}

func (m *mockPaymentRepo) FindByID(id domain.PaymentID) (*domain.Payment, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}

func (m *mockPaymentRepo) FindByOrderID(orderID domain.OrderID) (*domain.Payment, error) {
	if m.findByOrderIDFn != nil {
		return m.findByOrderIDFn(orderID)
	}
	return nil, nil
}

func (m *mockPaymentRepo) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Payment, error) {
	if m.findByRestaurantIDFn != nil {
		return m.findByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

func (m *mockPaymentRepo) Update(payment *domain.Payment) error {
	if m.updateFn != nil {
		return m.updateFn(payment)
	}
	return nil
}

// Helper to create a valid order
func createTestOrder(id string, status domain.FulfillmentStatus, paymentStatus domain.PaymentStatus) *domain.Order {
	return &domain.Order{
		ID:                domain.OrderID(id),
		OrderNumber:       domain.OrderNumber("1234"),
		RestaurantID:      domain.RestaurantID("rest-1"),
		TotalAmount:       1000,
		PaymentStatus:     paymentStatus,
		FulfillmentStatus: status,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}
