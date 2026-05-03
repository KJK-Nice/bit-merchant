package webpush

import (
	"sync"

	"bitmerchant/internal/common"
)

type MemoryRepository struct {
	mu   sync.RWMutex
	subs map[string]*Subscription // keyed by endpoint
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{subs: make(map[string]*Subscription)}
}

func (r *MemoryRepository) Upsert(sub *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[sub.Endpoint] = sub
	return nil
}

func (r *MemoryRepository) FindByOrderNumber(orderNumber string) ([]*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Subscription
	for _, s := range r.subs {
		if s.Role == "customer" && s.OrderNumber == orderNumber {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *MemoryRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Subscription
	for _, s := range r.subs {
		if s.Role == "kitchen" && s.RestaurantID == restaurantID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *MemoryRepository) DeleteByEndpoint(endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subs, endpoint)
	return nil
}
