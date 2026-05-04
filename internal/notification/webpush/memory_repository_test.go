package webpush_test

import (
	"testing"

	"bitmerchant/internal/common"
	"bitmerchant/internal/notification/webpush"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, repo.Upsert(sub))
	first := sub.ID
	require.NotEmpty(t, first, "upsert must populate sub.ID")

	// Subscribe again for the same device (idempotent re-subscribe). The ID
	// must not change — that's what proves "one row per device".
	again := &webpush.Subscription{
		Role:      "customer",
		Endpoint:  sub.Endpoint,
		AuthKey:   "auth-rotated",
		P256DHKey: "p256dh-rotated",
	}
	require.NoError(t, repo.Upsert(again))
	assert.Equal(t, first, again.ID, "re-upsert must return same id")

	for _, order := range []string{"0215", "0216"} {
		require.NoError(t, repo.AddScope(first, webpush.ScopeTypeOrder, order))
	}

	for _, order := range []string{"0215", "0216"} {
		subs, err := repo.FindByOrderNumber(order)
		require.NoError(t, err)
		require.Len(t, subs, 1, "find %s", order)
		assert.Equal(t, sub.Endpoint, subs[0].Endpoint, "find %s endpoint", order)
	}
}

// Customer/kitchen on the same physical device must coexist as separate rows
// (different role) — needed for staff who use one phone for both screens.
func TestMemoryRepository_RoleSeparation(t *testing.T) {
	repo := webpush.NewMemoryRepository()
	const endpoint = "https://example.com/push/shared"

	customer := &webpush.Subscription{Role: "customer", Endpoint: endpoint, AuthKey: "a", P256DHKey: "p"}
	kitchen := &webpush.Subscription{Role: "kitchen", Endpoint: endpoint, AuthKey: "a", P256DHKey: "p"}
	require.NoError(t, repo.Upsert(customer))
	require.NoError(t, repo.Upsert(kitchen))
	require.NotEmpty(t, customer.ID)
	require.NotEmpty(t, kitchen.ID)
	assert.NotEqual(t, customer.ID, kitchen.ID, "customer and kitchen rows must be distinct")

	require.NoError(t, repo.AddScope(customer.ID, webpush.ScopeTypeOrder, "0215"))
	require.NoError(t, repo.AddScope(kitchen.ID, webpush.ScopeTypeRestaurant, "rest-1"))

	custResults, err := repo.FindByOrderNumber("0215")
	require.NoError(t, err)
	require.Len(t, custResults, 1)
	assert.Equal(t, "customer", custResults[0].Role)

	kitResults, err := repo.FindByRestaurantID(common.RestaurantID("rest-1"))
	require.NoError(t, err)
	require.Len(t, kitResults, 1)
	assert.Equal(t, "kitchen", kitResults[0].Role)
}

// 410 Gone cleanup must drop the subscription and all of its scopes — mirrors
// Postgres ON DELETE CASCADE.
func TestMemoryRepository_DeleteByEndpointCascadesScopes(t *testing.T) {
	repo := webpush.NewMemoryRepository()
	sub := &webpush.Subscription{Role: "customer", Endpoint: "https://example.com/dead", AuthKey: "a", P256DHKey: "p"}
	require.NoError(t, repo.Upsert(sub))
	require.NoError(t, repo.AddScope(sub.ID, webpush.ScopeTypeOrder, "0215"))
	require.NoError(t, repo.DeleteByEndpoint(sub.Endpoint))

	results, err := repo.FindByOrderNumber("0215")
	require.NoError(t, err)
	assert.Empty(t, results, "after delete, find must return zero subs")
}
