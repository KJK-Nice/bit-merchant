package webpush

import "bitmerchant/internal/common"

// Subscription represents a browser push subscription stored in the DB.
// One row per (endpoint, role) — a single device gets one identity, and the
// scopes it's interested in (orders, restaurants) live in a separate table.
type Subscription struct {
	ID        string
	Role      string // "customer" or "kitchen"
	Endpoint  string
	AuthKey   string
	P256DHKey string
}

// ScopeType narrows what kind of thing a subscription wants pings for.
type ScopeType string

const (
	ScopeTypeOrder      ScopeType = "order"
	ScopeTypeRestaurant ScopeType = "restaurant"
)

// Repository persists and queries push subscriptions and their scopes.
type Repository interface {
	// Upsert inserts or updates the subscription identified by (endpoint, role)
	// and populates sub.ID on return. Idempotent — calling it again with the
	// same endpoint refreshes the encryption material in place.
	Upsert(sub *Subscription) error
	// AddScope links the subscription to a scope (e.g. an order number or a
	// restaurant id). Idempotent — re-adding the same scope is a no-op.
	AddScope(subscriptionID string, scopeType ScopeType, scopeID string) error
	FindByOrderNumber(orderNumber string) ([]*Subscription, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error)
	// DeleteByEndpoint removes a subscription, used when the push service
	// returns 410 Gone. CASCADE drops every scope tied to that subscription.
	DeleteByEndpoint(endpoint string) error
}
