package adapters

import (
	"context"
	"log/slog"
)

// LoggingMailer is the development mailer: it logs the reset link instead of
// sending an email, so the flow is fully exercisable without an email provider.
type LoggingMailer struct {
	logger *slog.Logger
}

func NewLoggingMailer(logger *slog.Logger) *LoggingMailer {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingMailer{logger: logger}
}

func (m *LoggingMailer) SendPasswordReset(_ context.Context, email, resetURL string) error {
	// NOTE: dev only. The link is logged, not emailed. Do not enable in prod.
	m.logger.Info("[dev-mailer] password reset link", "email", email, "reset_url", resetURL)
	return nil
}
