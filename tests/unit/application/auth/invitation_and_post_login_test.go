package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/app/query"
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- IssueKitchenStaffInvitation ----

func TestIssueKitchenStaffInvitation_NotOwner(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	invRepo := new(mockInvitationRepo)

	staffMem, _ := membership.NewMembership("m1", "u-staff", "rest-1", common.RoleKitchenStaff)
	memRepo.On("FindByUserAndRestaurant", common.UserID("u-staff"), common.RestaurantID("rest-1")).
		Return(staffMem, nil)

	h := command.NewIssueKitchenStaffInvitationHandler(memRepo, invRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.IssueKitchenStaffInvitation{
		OwnerUserID:  "u-staff",
		RestaurantID: "rest-1",
	})

	assert.ErrorIs(t, err, command.ErrNotRestaurantOwner)
	invRepo.AssertNotCalled(t, "Save")
}

func TestIssueKitchenStaffInvitation_OwnerIssuesToken(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	invRepo := new(mockInvitationRepo)

	ownerMem, _ := membership.NewMembership("m2", "u-owner", "rest-1", common.RoleOwner)
	memRepo.On("FindByUserAndRestaurant", common.UserID("u-owner"), common.RestaurantID("rest-1")).
		Return(ownerMem, nil)

	var capturedInv *invitation.Invitation
	invRepo.On("Save", mock.AnythingOfType("*invitation.Invitation")).Run(func(args mock.Arguments) {
		capturedInv = args.Get(0).(*invitation.Invitation)
	}).Return(nil)

	h := command.NewIssueKitchenStaffInvitationHandler(memRepo, invRepo, nil, nil)
	result, err := h.Handle(context.Background(), command.IssueKitchenStaffInvitation{
		OwnerUserID:  "u-owner",
		RestaurantID: "rest-1",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, result.Token, capturedInv.Token)
	assert.Equal(t, common.RoleKitchenStaff, capturedInv.Role)
	assert.True(t, capturedInv.ExpiresAt.After(time.Now()), "invitation should have a future expiry")
}

func TestIssueKitchenStaffInvitation_MembershipNotFound(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	invRepo := new(mockInvitationRepo)

	memRepo.On("FindByUserAndRestaurant", common.UserID("u-nobody"), common.RestaurantID("rest-1")).
		Return(nil, errors.New("not found"))

	h := command.NewIssueKitchenStaffInvitationHandler(memRepo, invRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.IssueKitchenStaffInvitation{
		OwnerUserID:  "u-nobody",
		RestaurantID: "rest-1",
	})

	assert.ErrorIs(t, err, command.ErrNotRestaurantOwner)
}

// ---- EndCustomerSession ----

func TestEndCustomerSession_DeletesSession(t *testing.T) {
	sessRepo := new(mockSessionRepo)
	sessRepo.On("Delete", "sess-abc").Return(nil)

	h := command.NewEndCustomerSessionHandler(sessRepo, nil, nil)
	err := h.Handle(context.Background(), command.EndCustomerSession{SessionID: "sess-abc"})

	assert.NoError(t, err)
	sessRepo.AssertCalled(t, "Delete", "sess-abc")
}

func TestEndCustomerSession_EmptySessionIDIsNoop(t *testing.T) {
	sessRepo := new(mockSessionRepo)

	h := command.NewEndCustomerSessionHandler(sessRepo, nil, nil)
	err := h.Handle(context.Background(), command.EndCustomerSession{SessionID: ""})

	assert.NoError(t, err)
	sessRepo.AssertNotCalled(t, "Delete")
}

// ---- InvitationForToken query ----

func TestInvitationForToken_ReturnsInvitation(t *testing.T) {
	invRepo := new(mockInvitationRepo)

	expected, _ := invitation.NewInvitation("inv-q", "rest-1", common.RoleKitchenStaff, "tok-xyz", time.Now().Add(time.Hour))
	invRepo.On("FindByToken", "tok-xyz").Return(expected, nil)

	h := query.NewInvitationForTokenHandler(invRepo, nil, nil)
	result, err := h.Handle(context.Background(), query.InvitationForToken{Token: "tok-xyz"})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestInvitationForToken_NotFound(t *testing.T) {
	invRepo := new(mockInvitationRepo)
	invRepo.On("FindByToken", "missing").Return(nil, errors.New("not found"))

	h := query.NewInvitationForTokenHandler(invRepo, nil, nil)
	result, err := h.Handle(context.Background(), query.InvitationForToken{Token: "missing"})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- MembershipsForUser query ----

func TestMembershipsForUser_ReturnsList(t *testing.T) {
	memRepo := new(mockMembershipRepo)

	m1, _ := membership.NewMembership("m1", "u1", "rest-1", common.RoleOwner)
	m2, _ := membership.NewMembership("m2", "u1", "rest-2", common.RoleOwner)
	memRepo.On("FindByUserID", common.UserID("u1")).Return([]*membership.Membership{m1, m2}, nil)

	h := query.NewMembershipsForUserHandler(memRepo, nil, nil)
	result, err := h.Handle(context.Background(), query.MembershipsForUser{UserID: "u1"})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// ---- PostLoginRestaurantContext ----

func TestPostLoginRestaurantContext_NoMemberships(t *testing.T) {
	rid, redirect := query.PostLoginRestaurantContext(nil)
	assert.Nil(t, rid)
	assert.Equal(t, "/dashboard", redirect)
}

func TestPostLoginRestaurantContext_SingleOwnerMembership(t *testing.T) {
	mem, _ := membership.NewMembership("m1", "u1", "rest-1", common.RoleOwner)
	rid, redirect := query.PostLoginRestaurantContext([]*membership.Membership{mem})

	assert.NotNil(t, rid)
	assert.Equal(t, common.RestaurantID("rest-1"), *rid)
	assert.Equal(t, "/dashboard", redirect)
}

func TestPostLoginRestaurantContext_SingleKitchenStaffMembership(t *testing.T) {
	mem, _ := membership.NewMembership("m1", "u1", "rest-1", common.RoleKitchenStaff)
	rid, redirect := query.PostLoginRestaurantContext([]*membership.Membership{mem})

	assert.NotNil(t, rid)
	assert.Equal(t, "/kitchen", redirect)
}

func TestPostLoginRestaurantContext_MultipleMembershipsGoToSelectRestaurant(t *testing.T) {
	m1, _ := membership.NewMembership("m1", "u1", "rest-1", common.RoleOwner)
	m2, _ := membership.NewMembership("m2", "u1", "rest-2", common.RoleOwner)
	rid, redirect := query.PostLoginRestaurantContext([]*membership.Membership{m1, m2})

	assert.Nil(t, rid)
	assert.Equal(t, "/auth/select-restaurant", redirect)
}

// ---- session.Session helper ----

func TestSession_IsExpired(t *testing.T) {
	s := &session.Session{ExpiresAt: time.Now().Add(-time.Minute)}
	assert.True(t, s.IsExpired(time.Now()))

	s2 := &session.Session{ExpiresAt: time.Now().Add(time.Hour)}
	assert.False(t, s2.IsExpired(time.Now()))
}
