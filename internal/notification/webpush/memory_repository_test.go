package webpush_test

import (
	"testing"

	"bitmerchant/internal/common"
	"bitmerchant/internal/notification/webpush"
)

// The headline UX win of the per-device + scopes model: a single browser
// endpoint enabled once can collect notifications for multiple orders. This
// test pins the contract — adding a second order's scope must not create a
// new subscription, and a lookup for either order must return the device.
func TestMemoryRepository_OneDeviceManyOrders(t *testing.T) {
	repo := webpush.NewMemoryRepository()
	sub := &webpush.Subscription{
		Role:      "customer",
		Endpoint:  "https://web.push.apple.com/abc",
		AuthKey:   "auth",
		P256DHKey: "p256dh",
	}
	if err := repo.Upsert(sub); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	first := sub.ID
	if first == "" {
		t.Fatal("upsert must populate sub.ID")
	}

	// Subscribe again for the same device (idempotent re-subscribe). The ID
	// must not change — that's what proves "one row per device".
	again := &webpush.Subscription{
		Role:      "customer",
		Endpoint:  sub.Endpoint,
		AuthKey:   "auth-rotated",
		P256DHKey: "p256dh-rotated",
	}
	if err := repo.Upsert(again); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if again.ID != first {
		t.Fatalf("re-upsert must return same id: got %q want %q", again.ID, first)
	}

	for _, order := range []string{"0215", "0216"} {
		if err := repo.AddScope(first, webpush.ScopeTypeOrder, order); err != nil {
			t.Fatalf("add scope %s: %v", order, err)
		}
	}

	for _, order := range []string{"0215", "0216"} {
		subs, err := repo.FindByOrderNumber(order)
		if err != nil {
			t.Fatalf("find %s: %v", order, err)
		}
		if len(subs) != 1 {
			t.Fatalf("find %s: want 1 sub, got %d", order, len(subs))
		}
		if subs[0].Endpoint != sub.Endpoint {
			t.Fatalf("find %s: wrong endpoint %q", order, subs[0].Endpoint)
		}
	}
}

// Customer/kitchen on the same physical device must coexist as separate rows
// (different role) — needed for staff who use one phone for both screens.
func TestMemoryRepository_RoleSeparation(t *testing.T) {
	repo := webpush.NewMemoryRepository()
	const endpoint = "https://example.com/push/shared"

	customer := &webpush.Subscription{Role: "customer", Endpoint: endpoint, AuthKey: "a", P256DHKey: "p"}
	kitchen := &webpush.Subscription{Role: "kitchen", Endpoint: endpoint, AuthKey: "a", P256DHKey: "p"}
	if err := repo.Upsert(customer); err != nil {
		t.Fatalf("customer upsert: %v", err)
	}
	if err := repo.Upsert(kitchen); err != nil {
		t.Fatalf("kitchen upsert: %v", err)
	}
	if customer.ID == "" || kitchen.ID == "" || customer.ID == kitchen.ID {
		t.Fatalf("customer and kitchen rows must be distinct: customer=%q kitchen=%q", customer.ID, kitchen.ID)
	}
	if err := repo.AddScope(customer.ID, webpush.ScopeTypeOrder, "0215"); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddScope(kitchen.ID, webpush.ScopeTypeRestaurant, "rest-1"); err != nil {
		t.Fatal(err)
	}

	custResults, _ := repo.FindByOrderNumber("0215")
	kitResults, _ := repo.FindByRestaurantID(common.RestaurantID("rest-1"))
	if len(custResults) != 1 || custResults[0].Role != "customer" {
		t.Fatalf("FindByOrderNumber must return only the customer row, got %+v", custResults)
	}
	if len(kitResults) != 1 || kitResults[0].Role != "kitchen" {
		t.Fatalf("FindByRestaurantID must return only the kitchen row, got %+v", kitResults)
	}
}

// 410 Gone cleanup must drop the subscription and all of its scopes — mirrors
// Postgres ON DELETE CASCADE.
func TestMemoryRepository_DeleteByEndpointCascadesScopes(t *testing.T) {
	repo := webpush.NewMemoryRepository()
	sub := &webpush.Subscription{Role: "customer", Endpoint: "https://example.com/dead", AuthKey: "a", P256DHKey: "p"}
	if err := repo.Upsert(sub); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddScope(sub.ID, webpush.ScopeTypeOrder, "0215"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByEndpoint(sub.Endpoint); err != nil {
		t.Fatal(err)
	}
	results, _ := repo.FindByOrderNumber("0215")
	if len(results) != 0 {
		t.Fatalf("after delete, find must return zero subs, got %d", len(results))
	}
}
