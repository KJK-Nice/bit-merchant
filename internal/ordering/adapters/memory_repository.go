package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type MemoryOrderRepository struct {
	mu       sync.RWMutex
	orders   map[common.OrderID]*order.Order
	counters map[common.RestaurantID]int
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders:   make(map[common.OrderID]*order.Order),
		counters: make(map[common.RestaurantID]int),
	}
}

// NextOrderNumber mirrors the Postgres implementation's contract: monotonic
// per restaurant, no duplicates under concurrent callers. Held under the same
// write mutex that guards orders, so a Save() following NextOrderNumber()
// observes a consistent state in tests that exercise both.
func (r *MemoryOrderRepository) NextOrderNumber(restaurantID common.RestaurantID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[restaurantID]++
	return r.counters[restaurantID], nil
}

func (r *MemoryOrderRepository) Save(o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID] = o
	return nil
}

func (r *MemoryOrderRepository) FindByID(id common.OrderID) (*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func (r *MemoryOrderRepository) FindByOrderNumber(restaurantID common.RestaurantID, orderNumber string) (*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.orders {
		if o.RestaurantID == restaurantID && string(o.OrderNumber) == orderNumber {
			return o, nil
		}
	}
	return nil, errors.New("order not found")
}

func (r *MemoryOrderRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*order.Order
	for _, o := range r.orders {
		if o.RestaurantID == restaurantID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *MemoryOrderRepository) FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*order.Order
	for _, o := range r.orders {
		if o.RestaurantID == restaurantID {
			s := o.FulfillmentStatus
			if s == common.FulfillmentStatusPaid || s == common.FulfillmentStatusPreparing || s == common.FulfillmentStatusReady {
				result = append(result, o)
			}
		}
	}
	return result, nil
}

func (r *MemoryOrderRepository) FindBySessionID(sessionID string) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*order.Order
	for _, o := range r.orders {
		if o.SessionID == sessionID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *MemoryOrderRepository) Update(o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[o.ID]; !exists {
		return errors.New("order not found")
	}
	r.orders[o.ID] = o
	return nil
}
