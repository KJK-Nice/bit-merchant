package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Save(s *session.Session) error {
	return m.Called(s).Error(0)
}
func (m *mockSessionRepo) Get(id string) (*session.Session, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}
func (m *mockSessionRepo) Delete(id string) error {
	return m.Called(id).Error(0)
}
func (m *mockSessionRepo) DeleteByUserID(uid common.UserID) error {
	return m.Called(uid).Error(0)
}

func ownerMembership(userID common.UserID, restaurantID common.RestaurantID) *membership.Membership {
	m, _ := membership.NewMembership("mem-1", userID, restaurantID, common.RoleOwner)
	return m
}

func TestSwitchActiveRestaurant_MembershipNotFound(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	sessRepo := new(mockSessionRepo)

	memRepo.On("FindByUserAndRestaurant", common.UserID("u1"), common.RestaurantID("rest-2")).
		Return(nil, errors.New("not found"))

	h := command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.SwitchActiveRestaurant{
		SessionID:    "sess-1",
		UserID:       "u1",
		RestaurantID: "rest-2",
		SessionTTL:   time.Hour,
	})

	assert.ErrorIs(t, err, command.ErrMembershipNotFound)
	sessRepo.AssertNotCalled(t, "Save")
}

func TestSwitchActiveRestaurant_CreatesNewSessionWhenNotFound(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	sessRepo := new(mockSessionRepo)

	uid := common.UserID("u1")
	rid := common.RestaurantID("rest-1")
	memRepo.On("FindByUserAndRestaurant", uid, rid).Return(ownerMembership(uid, rid), nil)
	sessRepo.On("Get", "new-sess").Return(nil, errors.New("not found"))

	var savedSession *session.Session
	sessRepo.On("Save", mock.AnythingOfType("*session.Session")).Run(func(args mock.Arguments) {
		savedSession = args.Get(0).(*session.Session)
	}).Return(nil)

	h := command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, nil, nil)
	result, err := h.Handle(context.Background(), command.SwitchActiveRestaurant{
		SessionID:    "new-sess",
		UserID:       uid,
		RestaurantID: rid,
		SessionTTL:   2 * time.Hour,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uid, *savedSession.UserID)
	assert.Equal(t, rid, *savedSession.RestaurantID)
	assert.True(t, savedSession.ExpiresAt.After(time.Now()))
}

func TestSwitchActiveRestaurant_UpdatesExistingSession(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	sessRepo := new(mockSessionRepo)

	uid := common.UserID("u1")
	rid := common.RestaurantID("rest-2")
	memRepo.On("FindByUserAndRestaurant", uid, rid).Return(ownerMembership(uid, rid), nil)

	existingSession := &session.Session{ID: "sess-existing", CreatedAt: time.Now().Add(-time.Hour)}
	sessRepo.On("Get", "sess-existing").Return(existingSession, nil)
	sessRepo.On("Save", mock.Anything).Return(nil)

	h := command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, nil, nil)
	result, err := h.Handle(context.Background(), command.SwitchActiveRestaurant{
		SessionID:    "sess-existing",
		UserID:       uid,
		RestaurantID: rid,
		SessionTTL:   time.Hour,
	})

	assert.NoError(t, err)
	assert.Equal(t, rid, *result.RestaurantID)
}

func TestSwitchActiveRestaurant_DefaultTTLApplied(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	sessRepo := new(mockSessionRepo)

	uid := common.UserID("u1")
	rid := common.RestaurantID("rest-1")
	memRepo.On("FindByUserAndRestaurant", uid, rid).Return(ownerMembership(uid, rid), nil)
	sessRepo.On("Get", "sess-ttl").Return(nil, errors.New("not found"))

	var savedSession *session.Session
	sessRepo.On("Save", mock.Anything).Run(func(args mock.Arguments) {
		savedSession = args.Get(0).(*session.Session)
	}).Return(nil)

	h := command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, nil, nil)
	_, err := h.Handle(context.Background(), command.SwitchActiveRestaurant{
		SessionID:    "sess-ttl",
		UserID:       uid,
		RestaurantID: rid,
		SessionTTL:   0, // should default to 24h
	})

	assert.NoError(t, err)
	// ExpiresAt should be approximately 24 hours from now
	expectedMin := time.Now().Add(23 * time.Hour)
	assert.True(t, savedSession.ExpiresAt.After(expectedMin), "expected 24h TTL default")
}
