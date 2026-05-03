package notification

import (
	"context"

	"bitmerchant/internal/infrastructure/logging"
)

// Service fan-outs a Notification to every registered Notifier.
// Errors from individual channels are logged but never stop delivery to others.
type Service struct {
	notifiers []Notifier
	logger    *logging.Logger
}

func NewService(logger *logging.Logger, notifiers ...Notifier) *Service {
	return &Service{notifiers: notifiers, logger: logger}
}

func (s *Service) Send(ctx context.Context, n Notification) {
	for _, notifier := range s.notifiers {
		if err := notifier.Send(ctx, n); err != nil {
			s.logger.Error("notification delivery failed",
				"channel", notifier.Name(),
				"error", err,
			)
		}
	}
}
