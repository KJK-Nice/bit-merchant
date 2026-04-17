package query

import (
	"context"
	"log/slog"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
)

// MembershipsForUser lists all restaurant memberships for a user.
type MembershipsForUser struct {
	UserID common.UserID
}

type MembershipsForUserHandler decorator.QueryHandler[MembershipsForUser, []*membership.Membership]

type membershipsForUserHandler struct {
	repo membership.Repository
}

func NewMembershipsForUserHandler(repo membership.Repository, log *slog.Logger, metrics decorator.MetricsClient) MembershipsForUserHandler {
	if repo == nil {
		panic("nil membership.Repository")
	}
	h := membershipsForUserHandler{repo: repo}
	return decorator.ApplyQueryDecorators[MembershipsForUser, []*membership.Membership](h, log, metrics)
}

func (h membershipsForUserHandler) Handle(ctx context.Context, q MembershipsForUser) ([]*membership.Membership, error) {
	_ = ctx
	return h.repo.FindByUserID(q.UserID)
}
