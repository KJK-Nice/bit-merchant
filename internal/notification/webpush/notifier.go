package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	webpushlib "github.com/SherClockHolmes/webpush-go"

	"bitmerchant/internal/common"
	"bitmerchant/internal/notification"
)

// VAPIDConfig holds the VAPID key pair and subject required to send Web Push messages.
type VAPIDConfig struct {
	PublicKey  string
	PrivateKey string
	Subject    string // mailto: or https: URL identifying the sender
}

// SendFunc is the signature used to actually deliver a push message.
// Defaults to webpushlib.SendNotification; override in tests.
type SendFunc func(message []byte, s *webpushlib.Subscription, options *webpushlib.Options) (*http.Response, error)

// Notifier implements notification.Notifier using the Web Push protocol.
type Notifier struct {
	repo   Repository
	vapid  VAPIDConfig
	send   SendFunc
	logger *slog.Logger
}

func NewNotifier(repo Repository, vapid VAPIDConfig, logger *slog.Logger) *Notifier {
	if logger == nil {
		logger = slog.Default()
	}
	return &Notifier{repo: repo, vapid: vapid, send: webpushlib.SendNotification, logger: logger}
}

func (n *Notifier) Name() string { return "web-push" }

func (n *Notifier) Send(ctx context.Context, notif notification.Notification) error {
	subs, err := n.subscriptionsFor(notif)
	if err != nil {
		return fmt.Errorf("query subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return nil
	}

	payload, err := json.Marshal(map[string]string{
		"title": notif.Title,
		"body":  notif.Body,
		"url":   notif.URL,
	})
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	// Per-subscription delivery is best-effort: a failure on one endpoint
	// (network blip, expired credentials) must not block delivery to siblings.
	for _, sub := range subs {
		if sendErr := n.sendOne(sub, payload); sendErr != nil {
			n.logger.Warn("web push delivery failed for endpoint",
				"endpoint", sub.Endpoint,
				"role", sub.Role,
				"error", sendErr,
			)
		}
	}
	return nil
}

func (n *Notifier) sendOne(sub *Subscription, payload []byte) error {
	wpSub := &webpushlib.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpushlib.Keys{
			Auth:   sub.AuthKey,
			P256dh: sub.P256DHKey,
		},
	}

	resp, err := n.send(payload, wpSub, &webpushlib.Options{
		VAPIDPublicKey:  n.vapid.PublicKey,
		VAPIDPrivateKey: n.vapid.PrivateKey,
		Subscriber:      n.vapid.Subject,
		TTL:             30,
	})
	if err != nil {
		return fmt.Errorf("send push to %s: %w", sub.Endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		if delErr := n.repo.DeleteByEndpoint(sub.Endpoint); delErr != nil {
			n.logger.Warn("failed to delete expired push subscription",
				"endpoint", sub.Endpoint,
				"error", delErr,
			)
		}
	}
	return nil
}

// WithSendFunc returns a copy of the Notifier with a custom send function (for testing).
func (n *Notifier) WithSendFunc(fn SendFunc) *Notifier {
	return &Notifier{repo: n.repo, vapid: n.vapid, send: fn, logger: n.logger}
}

func (n *Notifier) subscriptionsFor(notif notification.Notification) ([]*Subscription, error) {
	role := notif.Metadata["role"]
	switch role {
	case "customer":
		orderNumber := notif.Metadata["order_number"]
		return n.repo.FindByOrderNumber(orderNumber)
	case "kitchen":
		restaurantID := notif.Metadata["restaurant_id"]
		return n.repo.FindByRestaurantID(common.RestaurantID(restaurantID))
	default:
		return nil, nil
	}
}
