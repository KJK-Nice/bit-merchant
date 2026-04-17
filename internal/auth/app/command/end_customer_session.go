package command

import (
	"context"
	"log/slog"

	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common/decorator"
)

// EndCustomerSession removes the persisted session record.
type EndCustomerSession struct {
	SessionID string
}

type EndCustomerSessionHandler decorator.CommandHandler[EndCustomerSession]

type endCustomerSessionHandler struct {
	sessRepo session.Repository
}

func NewEndCustomerSessionHandler(sessRepo session.Repository, log *slog.Logger, metrics decorator.MetricsClient) EndCustomerSessionHandler {
	if sessRepo == nil {
		panic("nil session.Repository")
	}
	h := endCustomerSessionHandler{sessRepo: sessRepo}
	return decorator.ApplyCommandDecorators[EndCustomerSession](h, log, metrics)
}

func (h endCustomerSessionHandler) Handle(ctx context.Context, cmd EndCustomerSession) error {
	_ = ctx
	if cmd.SessionID == "" {
		return nil
	}
	_ = h.sessRepo.Delete(cmd.SessionID)
	return nil
}
