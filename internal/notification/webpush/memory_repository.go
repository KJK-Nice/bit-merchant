package webpush

import (
	"strconv"
	"sync"
	"sync/atomic"

	"bitmerchant/internal/common"
)

// MemoryRepository mirrors the Postgres schema for unit tests:
// one entry per (endpoint, role), and a separate set of scopes per
// subscription. Find lookups join the two.
type MemoryRepository struct {
	mu      sync.RWMutex
	nextID  int64
	subs    map[memSubKey]*Subscription
	subByID map[string]*Subscription
	// scopes[subscriptionID] -> set of (type|id) tuples this device wants pings for
	scopes map[string]map[memScope]struct{}
}

type memSubKey struct {
	endpoint string
	role     string
}

type memScope struct {
	scopeType ScopeType
	scopeID   string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		subs:    make(map[memSubKey]*Subscription),
		subByID: make(map[string]*Subscription),
		scopes:  make(map[string]map[memScope]struct{}),
	}
}

func (r *MemoryRepository) Upsert(sub *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := memSubKey{endpoint: sub.Endpoint, role: sub.Role}
	if existing, ok := r.subs[key]; ok {
		existing.AuthKey = sub.AuthKey
		existing.P256DHKey = sub.P256DHKey
		sub.ID = existing.ID
		return nil
	}
	sub.ID = strconv.FormatInt(atomic.AddInt64(&r.nextID, 1), 10)
	stored := *sub
	r.subs[key] = &stored
	r.subByID[stored.ID] = &stored
	return nil
}

func (r *MemoryRepository) AddScope(subscriptionID string, scopeType ScopeType, scopeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	scopes, ok := r.scopes[subscriptionID]
	if !ok {
		scopes = make(map[memScope]struct{})
		r.scopes[subscriptionID] = scopes
	}
	scopes[memScope{scopeType: scopeType, scopeID: scopeID}] = struct{}{}
	return nil
}

func (r *MemoryRepository) FindByOrderNumber(orderNumber string) ([]*Subscription, error) {
	return r.findByScope("customer", ScopeTypeOrder, orderNumber), nil
}

func (r *MemoryRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error) {
	return r.findByScope("kitchen", ScopeTypeRestaurant, string(restaurantID)), nil
}

func (r *MemoryRepository) findByScope(role string, scopeType ScopeType, scopeID string) []*Subscription {
	r.mu.RLock()
	defer r.mu.RUnlock()
	target := memScope{scopeType: scopeType, scopeID: scopeID}
	var out []*Subscription
	for subID, scopes := range r.scopes {
		if _, hit := scopes[target]; !hit {
			continue
		}
		sub, ok := r.subByID[subID]
		if !ok || sub.Role != role {
			continue
		}
		copied := *sub
		out = append(out, &copied)
	}
	return out
}

// DeleteByEndpoint removes the subscription(s) at this endpoint and any of
// their scope rows. Mirrors the Postgres ON DELETE CASCADE behaviour.
func (r *MemoryRepository) DeleteByEndpoint(endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, sub := range r.subs {
		if key.endpoint != endpoint {
			continue
		}
		delete(r.subs, key)
		delete(r.subByID, sub.ID)
		delete(r.scopes, sub.ID)
	}
	return nil
}
