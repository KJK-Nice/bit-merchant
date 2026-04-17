package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"

	"github.com/google/uuid"
)

var (
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationNotUsable = errors.New("invitation expired or already used")
)

// AcceptInvitationForUser consumes an invitation and creates membership for a new user.
type AcceptInvitationForUser struct {
	NewUserID       common.UserID
	InvitationToken string
}

type AcceptInvitationForUserHandler decorator.CommandResultHandler[AcceptInvitationForUser, RegistrationOutcome]

type acceptInvitationForUserHandler struct {
	invRepo invitation.Repository
	memRepo membership.Repository
}

func NewAcceptInvitationForUserHandler(invRepo invitation.Repository, memRepo membership.Repository, log *slog.Logger, metrics decorator.MetricsClient) AcceptInvitationForUserHandler {
	if invRepo == nil || memRepo == nil {
		panic("nil repository")
	}
	h := acceptInvitationForUserHandler{invRepo: invRepo, memRepo: memRepo}
	return decorator.ApplyCommandResultDecorators[AcceptInvitationForUser, RegistrationOutcome](h, log, metrics)
}

func (h acceptInvitationForUserHandler) Handle(ctx context.Context, cmd AcceptInvitationForUser) (RegistrationOutcome, error) {
	_ = ctx
	inv, err := h.invRepo.FindByToken(cmd.InvitationToken)
	if err != nil {
		return RegistrationOutcome{}, ErrInvitationNotFound
	}
	if inv.IsExpired(time.Now()) || inv.IsUsed() {
		return RegistrationOutcome{}, ErrInvitationNotUsable
	}

	mem, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		cmd.NewUserID,
		inv.RestaurantID,
		inv.Role,
	)
	if err != nil {
		return RegistrationOutcome{}, err
	}
	if err := h.memRepo.Save(mem); err != nil {
		return RegistrationOutcome{}, err
	}
	inv.MarkUsed(cmd.NewUserID, time.Now())
	if err := h.invRepo.Update(inv); err != nil {
		return RegistrationOutcome{}, err
	}

	redirect := "/dashboard"
	if inv.Role == common.RoleKitchenStaff {
		redirect = "/kitchen"
	}
	rid := inv.RestaurantID
	return RegistrationOutcome{RestaurantID: &rid, Redirect: redirect}, nil
}
