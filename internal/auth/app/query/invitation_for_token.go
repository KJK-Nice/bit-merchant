package query

import (
	"context"
	"log/slog"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/common/decorator"
)

// InvitationForToken loads an invitation by its shareable token.
type InvitationForToken struct {
	Token string
}

type InvitationForTokenHandler decorator.QueryHandler[InvitationForToken, *invitation.Invitation]

type invitationForTokenHandler struct {
	repo invitation.Repository
}

func NewInvitationForTokenHandler(repo invitation.Repository, log *slog.Logger, metrics decorator.MetricsClient) InvitationForTokenHandler {
	if repo == nil {
		panic("nil invitation.Repository")
	}
	h := invitationForTokenHandler{repo: repo}
	return decorator.ApplyQueryDecorators[InvitationForToken, *invitation.Invitation](h, log, metrics)
}

func (h invitationForTokenHandler) Handle(ctx context.Context, q InvitationForToken) (*invitation.Invitation, error) {
	_ = ctx
	return h.repo.FindByToken(q.Token)
}
