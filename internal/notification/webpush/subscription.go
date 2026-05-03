package webpush

import "bitmerchant/internal/common"

// Subscription represents a browser push subscription stored in the DB.
type Subscription struct {
	ID           string
	Role         string              // "customer" or "kitchen"
	OrderNumber  string              // set for customer subscriptions
	RestaurantID common.RestaurantID // set for kitchen subscriptions
	Endpoint     string
	AuthKey      string
	P256DHKey    string
}

// Repository persists and queries push subscriptions.
type Repository interface {
	// Upsert inserts or updates a subscription by endpoint (idempotent re-subscribe).
	Upsert(sub *Subscription) error
	FindByOrderNumber(orderNumber string) ([]*Subscription, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Subscription, error)
	// DeleteByEndpoint removes a subscription, used when the push service returns 410 Gone.
	DeleteByEndpoint(endpoint string) error
}
