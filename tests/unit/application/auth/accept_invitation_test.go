package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock invitation repository ---

type mockInvitationRepo struct{ mock.Mock }

func (m *mockInvitationRepo) Save(inv *invitation.Invitation) error {
	return m.Called(inv).Error(0)
}
func (m *mockInvitationRepo) FindByToken(token string) (*invitation.Invitation, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*invitation.Invitation), args.Error(1)
}
func (m *mockInvitationRepo) FindByRestaurantID(id common.RestaurantID) ([]*invitation.Invitation, error) {
	args := m.Called(id)
	return args.Get(0).([]*invitation.Invitation), args.Error(1)
}
func (m *mockInvitationRepo) Update(inv *invitation.Invitation) error {
	return m.Called(inv).Error(0)
}

// --- mock membership repository ---

type mockMembershipRepo struct{ mock.Mock }

func (m *mockMembershipRepo) Save(mem *membership.Membership) error {
	return m.Called(mem).Error(0)
}
func (m *mockMembershipRepo) FindByUserID(id common.UserID) ([]*membership.Membership, error) {
	args := m.Called(id)
	return args.Get(0).([]*membership.Membership), args.Error(1)
}
func (m *mockMembershipRepo) FindByRestaurantID(id common.RestaurantID) ([]*membership.Membership, error) {
	args := m.Called(id)
	return args.Get(0).([]*membership.Membership), args.Error(1)
}
func (m *mockMembershipRepo) FindByUserAndRestaurant(uid common.UserID, rid common.RestaurantID) (*membership.Membership, error) {
	args := m.Called(uid, rid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*membership.Membership), args.Error(1)
}
func (m *mockMembershipRepo) Delete(id common.MembershipID) error {
	return m.Called(id).Error(0)
}

// --- helpers ---

func newValidInvitation(token string, role common.MemberRole) *invitation.Invitation {
	inv, _ := invitation.NewInvitation(
		common.InvitationID("inv-1"),
		common.RestaurantID("rest-1"),
		role,
		token,
		time.Now().Add(7*24*time.Hour),
	)
	return inv
}

// --- tests ---

func TestAcceptInvitationForUser_TokenNotFound(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	invRepo.On("FindByToken", "bad-token").Return(nil, errors.New("not found"))

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-1",
		InvitationToken: "bad-token",
	})

	assert.ErrorIs(t, err, command.ErrInvitationNotFound)
	memRepo.AssertNotCalled(t, "Save")
}

func TestAcceptInvitationForUser_ExpiredInvitation(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	// Manually build an already-expired invitation (bypass constructor validation).
	expired := &invitation.Invitation{
		ID:           "inv-2",
		RestaurantID: "rest-1",
		Role:         common.RoleKitchenStaff,
		Token:        "expired-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}
	invRepo.On("FindByToken", "expired-token").Return(expired, nil)

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-1",
		InvitationToken: "expired-token",
	})

	assert.ErrorIs(t, err, command.ErrInvitationNotUsable)
	memRepo.AssertNotCalled(t, "Save")
}

func TestAcceptInvitationForUser_AlreadyUsedInvitation(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	inv := newValidInvitation("used-token", common.RoleKitchenStaff)
	uid := common.UserID("previous-user")
	inv.MarkUsed(uid, time.Now())

	invRepo.On("FindByToken", "used-token").Return(inv, nil)

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-2",
		InvitationToken: "used-token",
	})

	assert.ErrorIs(t, err, command.ErrInvitationNotUsable)
	memRepo.AssertNotCalled(t, "Save")
}

func TestAcceptInvitationForUser_KitchenStaffRedirect(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	inv := newValidInvitation("staff-token", common.RoleKitchenStaff)
	invRepo.On("FindByToken", "staff-token").Return(inv, nil)
	memRepo.On("Save", mock.AnythingOfType("*membership.Membership")).Return(nil)
	invRepo.On("Update", mock.AnythingOfType("*invitation.Invitation")).Return(nil)

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	outcome, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-3",
		InvitationToken: "staff-token",
	})

	assert.NoError(t, err)
	assert.Equal(t, "/kitchen", outcome.Redirect)
	assert.NotNil(t, outcome.RestaurantID)
	assert.Equal(t, common.RestaurantID("rest-1"), *outcome.RestaurantID)
}

func TestAcceptInvitationForUser_OwnerRedirect(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	inv := newValidInvitation("owner-token", common.RoleOwner)
	invRepo.On("FindByToken", "owner-token").Return(inv, nil)
	memRepo.On("Save", mock.AnythingOfType("*membership.Membership")).Return(nil)
	invRepo.On("Update", mock.AnythingOfType("*invitation.Invitation")).Return(nil)

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	outcome, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-4",
		InvitationToken: "owner-token",
	})

	assert.NoError(t, err)
	assert.Equal(t, "/dashboard", outcome.Redirect)
}

func TestAcceptInvitationForUser_MarksInvitationUsed(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	memRepo := new(mockMembershipRepo)

	inv := newValidInvitation("mark-token", common.RoleKitchenStaff)
	invRepo.On("FindByToken", "mark-token").Return(inv, nil)
	memRepo.On("Save", mock.Anything).Return(nil)

	var capturedInv *invitation.Invitation
	invRepo.On("Update", mock.AnythingOfType("*invitation.Invitation")).Run(func(args mock.Arguments) {
		capturedInv = args.Get(0).(*invitation.Invitation)
	}).Return(nil)

	h := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.AcceptInvitationForUser{
		NewUserID:       "user-5",
		InvitationToken: "mark-token",
	})

	assert.NoError(t, err)
	assert.True(t, capturedInv.IsUsed())
	assert.Equal(t, common.UserID("user-5"), *capturedInv.UsedByUserID)
}
