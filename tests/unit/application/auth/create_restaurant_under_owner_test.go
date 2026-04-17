package auth_test

import (
	"context"
	"errors"
	"testing"

	"bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/adapters"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newRealCreateRestaurantHandler() restaurantCmd.CreateRestaurantHandler {
	return restaurantCmd.NewCreateRestaurantHandler(adapters.NewMemoryRestaurantRepository(), nil, nil)
}

func TestCreateRestaurantUnderOwner_NotOwner(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	createRest := newRealCreateRestaurantHandler()

	staffMem, _ := membership.NewMembership("mem-k", "u-staff", "rest-1", common.RoleKitchenStaff)
	memRepo.On("FindByUserAndRestaurant", common.UserID("u-staff"), common.RestaurantID("rest-1")).
		Return(staffMem, nil)

	h := command.NewCreateRestaurantUnderOwnerHandler(memRepo, createRest, nil, nil)
	result, err := h.Handle(context.Background(), command.CreateRestaurantUnderOwner{
		OwnerUserID:              "u-staff",
		OwnerContextRestaurantID: "rest-1",
		Name:                     "New Location",
	})

	assert.ErrorIs(t, err, command.ErrNotRestaurantOwner)
	assert.Nil(t, result)
}

func TestCreateRestaurantUnderOwner_MembershipNotFound(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	createRest := newRealCreateRestaurantHandler()

	memRepo.On("FindByUserAndRestaurant", common.UserID("u-ghost"), common.RestaurantID("rest-1")).
		Return(nil, errors.New("not found"))

	h := command.NewCreateRestaurantUnderOwnerHandler(memRepo, createRest, nil, nil)
	result, err := h.Handle(context.Background(), command.CreateRestaurantUnderOwner{
		OwnerUserID:              "u-ghost",
		OwnerContextRestaurantID: "rest-1",
		Name:                     "New Location",
	})

	assert.ErrorIs(t, err, command.ErrNotRestaurantOwner)
	assert.Nil(t, result)
}

func TestCreateRestaurantUnderOwner_OwnerCreatesNewRestaurant(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	createRest := newRealCreateRestaurantHandler()

	ownerMem, _ := membership.NewMembership("mem-o", "u-owner", "rest-1", common.RoleOwner)
	memRepo.On("FindByUserAndRestaurant", common.UserID("u-owner"), common.RestaurantID("rest-1")).
		Return(ownerMem, nil)
	memRepo.On("Save", mock.AnythingOfType("*membership.Membership")).Return(nil)

	h := command.NewCreateRestaurantUnderOwnerHandler(memRepo, createRest, nil, nil)
	result, err := h.Handle(context.Background(), command.CreateRestaurantUnderOwner{
		OwnerUserID:              "u-owner",
		OwnerContextRestaurantID: "rest-1",
		Name:                     "Second Location",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Second Location", result.Name)
	assert.NotEmpty(t, result.ID)
	memRepo.AssertExpectations(t)
}

func TestCompleteSignupNewRestaurant_CreatesRestaurantAndOwnerMembership(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	createRest := newRealCreateRestaurantHandler()

	memRepo.On("Save", mock.MatchedBy(func(m *membership.Membership) bool {
		return m.UserID == "u-new" && m.Role == common.RoleOwner
	})).Return(nil)

	h := command.NewCompleteSignupNewRestaurantHandler(memRepo, createRest, nil, nil)
	outcome, err := h.Handle(context.Background(), command.CompleteSignupNewRestaurant{
		OwnerUserID:    "u-new",
		RestaurantName: "Grand Bistro",
	})

	assert.NoError(t, err)
	assert.Equal(t, "/dashboard", outcome.Redirect)
	assert.NotNil(t, outcome.RestaurantID)
	memRepo.AssertExpectations(t)
}

func TestCompleteSignupNewRestaurant_MembershipSaveError(t *testing.T) {
	memRepo := new(mockMembershipRepo)
	createRest := newRealCreateRestaurantHandler()

	memRepo.On("Save", mock.Anything).Return(errors.New("db error"))

	h := command.NewCompleteSignupNewRestaurantHandler(memRepo, createRest, nil, nil)
	outcome, err := h.Handle(context.Background(), command.CompleteSignupNewRestaurant{
		OwnerUserID:    "u-new",
		RestaurantName: "Grand Bistro",
	})

	assert.Error(t, err)
	assert.Nil(t, outcome.RestaurantID)
}
