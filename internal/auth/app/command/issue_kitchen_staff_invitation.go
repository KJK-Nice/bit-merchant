package command

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"time"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"

	"github.com/google/uuid"
)

// IssueKitchenStaffInvitationResult carries the token for the invite URL.
type IssueKitchenStaffInvitationResult struct {
	Token string
}

// IssueKitchenStaffInvitation creates a kitchen-staff invitation for a restaurant.
type IssueKitchenStaffInvitation struct {
	OwnerUserID  common.UserID
	RestaurantID common.RestaurantID
}

type IssueKitchenStaffInvitationHandler decorator.CommandResultHandler[IssueKitchenStaffInvitation, IssueKitchenStaffInvitationResult]

type issueKitchenStaffInvitationHandler struct {
	memRepo membership.Repository
	invRepo invitation.Repository
}

func NewIssueKitchenStaffInvitationHandler(memRepo membership.Repository, invRepo invitation.Repository, log *slog.Logger, metrics decorator.MetricsClient) IssueKitchenStaffInvitationHandler {
	if memRepo == nil || invRepo == nil {
		panic("nil repository")
	}
	h := issueKitchenStaffInvitationHandler{memRepo: memRepo, invRepo: invRepo}
	return decorator.ApplyCommandResultDecorators[IssueKitchenStaffInvitation, IssueKitchenStaffInvitationResult](h, log, metrics)
}

func (h issueKitchenStaffInvitationHandler) Handle(ctx context.Context, cmd IssueKitchenStaffInvitation) (IssueKitchenStaffInvitationResult, error) {
	_ = ctx
	mem, err := h.memRepo.FindByUserAndRestaurant(cmd.OwnerUserID, cmd.RestaurantID)
	if err != nil || mem == nil || mem.Role != common.RoleOwner {
		return IssueKitchenStaffInvitationResult{}, ErrNotRestaurantOwner
	}

	token, err := randomInviteToken()
	if err != nil {
		return IssueKitchenStaffInvitationResult{}, err
	}

	inv, err := invitation.NewInvitation(
		common.InvitationID(uuid.NewString()),
		cmd.RestaurantID,
		common.RoleKitchenStaff,
		token,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return IssueKitchenStaffInvitationResult{}, err
	}
	if err := h.invRepo.Save(inv); err != nil {
		return IssueKitchenStaffInvitationResult{}, err
	}
	return IssueKitchenStaffInvitationResult{Token: token}, nil
}

func randomInviteToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
