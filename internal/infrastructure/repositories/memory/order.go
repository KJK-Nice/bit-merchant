package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryOrderRepository implements OrderRepository with in-memory storage
type MemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[domain.OrderID]*domain.Order
}

// NewMemoryOrderRepository creates a new in-memory order repository
func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders: make(map[domain.OrderID]*domain.Order),
	}
}

// Save saves an order
func (r *MemoryOrderRepository) Save(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

// FindByID finds an order by ID
func (r *MemoryOrderRepository) FindByID(id domain.OrderID) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// FindByOrderNumber finds an order by order number
func (r *MemoryOrderRepository) FindByOrderNumber(restaurantID domain.RestaurantID, orderNumber string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, order := range r.orders {
		if order.RestaurantID == restaurantID && string(order.OrderNumber) == orderNumber {
			return order, nil
		}
	}
	return nil, errors.New("order not found")
}

// FindByRestaurantID finds all orders for a restaurant
func (r *MemoryOrderRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Order
	for _, order := range r.orders {
		if order.RestaurantID == restaurantID {
			result = append(result, order)
		}
	}
	return result, nil
}

// FindActiveByRestaurantID finds active orders for a restaurant
func (r *MemoryOrderRepository) FindActiveByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Order
	for _, order := range r.orders {
		if order.RestaurantID == restaurantID {
			status := order.FulfillmentStatus
			if status == domain.FulfillmentStatusPaid ||
				status == domain.FulfillmentStatusPreparing ||
				status == domain.FulfillmentStatusReady {
				result = append(result, order)
			}
		}
	}
	return result, nil
}

// FindBySessionID finds all orders for a session
func (r *MemoryOrderRepository) FindBySessionID(sessionID string) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Order
	for _, order := range r.orders {
		if order.SessionID == sessionID {
			result = append(result, order)
		}
	}
	return result, nil
}

// Update updates an order
func (r *MemoryOrderRepository) Update(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("order not found")
	}
	r.orders[order.ID] = order
	return nil
}
