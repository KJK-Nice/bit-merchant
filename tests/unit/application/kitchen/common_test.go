package kitchen_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"

	"context"
	"time"
)

type mockOrderRepo struct {
	saveFn                     func(order *order.Order) error
	findByIDFn                 func(id common.OrderID) (*order.Order, error)
	findByOrderNumberFn        func(restaurantID common.RestaurantID, orderNumber string) (*order.Order, error)
	findByRestaurantIDFn       func(restaurantID common.RestaurantID) ([]*order.Order, error)
	findActiveByRestaurantIDFn func(restaurantID common.RestaurantID) ([]*order.Order, error)
	findBySessionIDFn          func(sessionID string) ([]*order.Order, error)
	updateFn                   func(order *order.Order) error
}

func (m *mockOrderRepo) Save(order *order.Order) error {
	if m.saveFn != nil {
		return m.saveFn(order)
	}
	return nil
}

func (m *mockOrderRepo) FindByID(id common.OrderID) (*order.Order, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindByOrderNumber(restaurantID common.RestaurantID, orderNumber string) (*order.Order, error) {
	if m.findByOrderNumberFn != nil {
		return m.findByOrderNumberFn(restaurantID, orderNumber)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	if m.findByRestaurantIDFn != nil {
		return m.findByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	if m.findActiveByRestaurantIDFn != nil {
		return m.findActiveByRestaurantIDFn(restaurantID)
	}
	return nil, nil
}

func (m *mockOrderRepo) FindBySessionID(sessionID string) ([]*order.Order, error) {
	if m.findBySessionIDFn != nil {
		return m.findBySessionIDFn(sessionID)
	}
	return nil, nil
}

func (m *mockOrderRepo) Update(order *order.Order) error {
	if m.updateFn != nil {
		return m.updateFn(order)
	}
	return nil
}

func (m *mockOrderRepo) NextOrderNumber(restaurantID common.RestaurantID) (int, error) {
	return 1, nil
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

// Helper to create a valid order
func createTestOrder(id string, status common.FulfillmentStatus, paymentStatus common.PaymentStatus) *order.Order {
	return &order.Order{
		ID:                common.OrderID(id),
		OrderNumber:       common.OrderNumber("1234"),
		RestaurantID:      common.RestaurantID("rest-1"),
		TotalAmount:       1000,
		PaymentStatus:     paymentStatus,
		FulfillmentStatus: status,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}
