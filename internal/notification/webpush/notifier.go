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
	role := notif.Metadata["role"]
	target := notif.Metadata["order_number"]
	if role == "kitchen" {
		target = notif.Metadata["restaurant_id"]
	}

	subs, err := n.subscriptionsFor(notif)
	if err != nil {
		return fmt.Errorf("query subscriptions: %w", err)
	}
	// Always log so the operator can tell whether a fired event reached the
	// notifier, how many subscriptions matched, and whether the no-op was
	// "no subs registered" vs "delivery failed". This is the difference
	// between debugging the publisher pipeline and debugging push delivery.
	n.logger.Info("web push send",
		"role", role,
		"target", target,
		"title", notif.Title,
		"subscriptions", len(subs),
	)
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
		// 1 hour: long enough that a phone backgrounded for a while will
		// still receive the message when it next syncs. The previous 30s
		// limit dropped messages whenever the device was asleep at send
		// time, which on mobile is most of the time.
		TTL: 3600,
	})
	if err != nil {
		return fmt.Errorf("send push to %s: %w", sub.Endpoint, err)
	}
	defer resp.Body.Close()

	// Surface non-2xx responses from the push service (FCM, Mozilla autopush,
	// Apple Push Service). 2xx means the push service accepted the message
	// for delivery; anything else explains why a notification never appeared.
	// Both branches log at Info: when notifications mysteriously don't arrive,
	// the operator needs to see "did the push service accept this?" without
	// reconfiguring log levels — the volume is bounded by status-change events.
	if resp.StatusCode >= 300 {
		n.logger.Warn("push service rejected delivery",
			"endpoint", sub.Endpoint,
			"status", resp.StatusCode,
		)
	} else {
		n.logger.Info("push service accepted delivery",
			"endpoint", sub.Endpoint,
			"status", resp.StatusCode,
		)
	}

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
