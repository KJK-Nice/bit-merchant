package webpush_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"bitmerchant/internal/common"
	"bitmerchant/internal/notification"
	"bitmerchant/internal/notification/webpush"

	webpushlib "github.com/SherClockHolmes/webpush-go"
)

// stubRepo is an in-memory Repository that records calls for assertions.
type stubRepo struct {
	byOrderNumber map[string][]*webpush.Subscription
	byRestaurant  map[common.RestaurantID][]*webpush.Subscription
	deleted       []string
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		byOrderNumber: make(map[string][]*webpush.Subscription),
		byRestaurant:  make(map[common.RestaurantID][]*webpush.Subscription),
	}
}

func (r *stubRepo) Upsert(sub *webpush.Subscription) error { return nil }
func (r *stubRepo) FindByOrderNumber(n string) ([]*webpush.Subscription, error) {
	return r.byOrderNumber[n], nil
}
func (r *stubRepo) FindByRestaurantID(id common.RestaurantID) ([]*webpush.Subscription, error) {
	return r.byRestaurant[id], nil
}
func (r *stubRepo) DeleteByEndpoint(endpoint string) error {
	r.deleted = append(r.deleted, endpoint)
	return nil
}

func fakeSend(status int) (webpush.SendFunc, *int) {
	calls := 0
	fn := func(_ []byte, _ *webpushlib.Subscription, _ *webpushlib.Options) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	return fn, &calls
}

func makeNotif(role, id string) notification.Notification {
	meta := map[string]string{"role": role}
	if role == "customer" {
		meta["order_number"] = id
	} else {
		meta["restaurant_id"] = id
	}
	return notification.Notification{Title: "T", Body: "B", URL: "/", Metadata: meta}
}

func TestNotifier_Send_CustomerSubscriptionsReceivePush(t *testing.T) {
	repo := newStubRepo()
	repo.byOrderNumber["ORD-1"] = []*webpush.Subscription{
		{Endpoint: "https://example.com/push/1", AuthKey: "auth1", P256DHKey: "p256dh1"},
		{Endpoint: "https://example.com/push/2", AuthKey: "auth2", P256DHKey: "p256dh2"},
	}

	fn, calls := fakeSend(http.StatusCreated)
	n := webpush.NewNotifier(repo, webpush.VAPIDConfig{}).WithSendFunc(fn)
	if err := n.Send(context.Background(), makeNotif("customer", "ORD-1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *calls != 2 {
		t.Fatalf("expected 2 push sends, got %d", *calls)
	}
}

func TestNotifier_Send_GoneResponseRemovesSubscription(t *testing.T) {
	repo := newStubRepo()
	const endpoint = "https://example.com/push/gone"
	repo.byOrderNumber["ORD-2"] = []*webpush.Subscription{
		{Endpoint: endpoint, AuthKey: "auth", P256DHKey: "p256dh"},
	}

	fn, _ := fakeSend(http.StatusGone)
	n := webpush.NewNotifier(repo, webpush.VAPIDConfig{}).WithSendFunc(fn)
	_ = n.Send(context.Background(), makeNotif("customer", "ORD-2"))

	if len(repo.deleted) != 1 || repo.deleted[0] != endpoint {
		t.Fatalf("expected endpoint to be deleted on 410 Gone, got deleted=%v", repo.deleted)
	}
}

func TestNotifier_Send_NoSubscriptions_NoError(t *testing.T) {
	repo := newStubRepo()
	fn, calls := fakeSend(http.StatusCreated)
	n := webpush.NewNotifier(repo, webpush.VAPIDConfig{}).WithSendFunc(fn)
	if err := n.Send(context.Background(), makeNotif("customer", "ORD-NONE")); err != nil {
		t.Fatalf("unexpected error when no subscriptions: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("expected 0 sends when no subscriptions, got %d", *calls)
	}
}

func TestNotifier_Send_KitchenSubscriptions(t *testing.T) {
	repo := newStubRepo()
	repo.byRestaurant["rest-1"] = []*webpush.Subscription{
		{Endpoint: "https://example.com/push/k", AuthKey: "auth", P256DHKey: "p256dh"},
	}

	fn, calls := fakeSend(http.StatusCreated)
	n := webpush.NewNotifier(repo, webpush.VAPIDConfig{}).WithSendFunc(fn)
	if err := n.Send(context.Background(), makeNotif("kitchen", "rest-1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *calls != 1 {
		t.Fatalf("expected 1 push send for kitchen, got %d", *calls)
	}
}
