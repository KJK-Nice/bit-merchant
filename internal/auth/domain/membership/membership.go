package membership

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// Membership associates a user to a restaurant with a role.
type Membership struct {
	ID           common.MembershipID
	UserID       common.UserID
	RestaurantID common.RestaurantID
	Role         common.MemberRole
	CreatedAt    time.Time
}

func NewMembership(id common.MembershipID, userID common.UserID, restaurantID common.RestaurantID, role common.MemberRole) (*Membership, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if restaurantID == "" {
		return nil, errors.New("restaurant ID is required")
	}
	if role != common.RoleOwner && role != common.RoleKitchenStaff && role != common.RoleCustomer {
		return nil, errors.New("invalid member role")
	}
	return &Membership{
		ID: id, UserID: userID, RestaurantID: restaurantID,
		Role: role, CreatedAt: time.Now(),
	}, nil
}
