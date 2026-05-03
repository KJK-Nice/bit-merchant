package webpush

import (
	"sync"

	"bitmerchant/internal/common"
)

// MemoryRepository keys subscriptions by (endpoint, role, order_number, restaurant_id)
// so a single browser endpoint can hold concurrent subscriptions for multiple
// in-flight orders or scopes — matching the postgres unique index.
type MemoryRepository struct {
	mu   sync.RWMutex
	subs map[memKey]*Subscription
}

type memKey struct {
	endpoint     string
	role         string
	orderNumber  string
	restaurantID common.RestaurantID
}

func keyFor(sub *Subscription) memKey {
	return memKey{
		endpoint:     sub.Endpoint,
		role:         sub.Role,
		orderNumber:  sub.OrderNumber,
		restaurantID: sub.RestaurantID,
	}
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{subs: make(map[memKey]*Subscription)}
}

func (r *MemoryRepository) Upsert(sub *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[keyFor(sub)] = sub
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

// DeleteByEndpoint removes every scope tied to an endpoint — used when the push
// service returns 410 Gone, which means the endpoint itself is dead so all
// scopes attached to it become unreachable.
func (r *MemoryRepository) DeleteByEndpoint(endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k := range r.subs {
		if k.endpoint == endpoint {
			delete(r.subs, k)
		}
	}
	return nil
}
