package domain

import (
	"errors"
	"time"
)

// MemberRole controls permissions inside a restaurant organization.
type MemberRole string

const (
	RoleOwner        MemberRole = "owner"
	RoleKitchenStaff MemberRole = "kitchen_staff"
	RoleCustomer     MemberRole = "customer"
)

// MembershipID represents a unique membership identifier.
type MembershipID string

// Membership associates a user to a restaurant with a role.
type Membership struct {
	ID           MembershipID
	UserID       UserID
	RestaurantID RestaurantID
	Role         MemberRole
	CreatedAt    time.Time
}

// NewMembership creates a role assignment.
func NewMembership(id MembershipID, userID UserID, restaurantID RestaurantID, role MemberRole) (*Membership, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if restaurantID == "" {
		return nil, errors.New("restaurant ID is required")
	}
	if role != RoleOwner && role != RoleKitchenStaff && role != RoleCustomer {
		return nil, errors.New("invalid member role")
	}

	return &Membership{
		ID:           id,
		UserID:       userID,
		RestaurantID: restaurantID,
		Role:         role,
		CreatedAt:    time.Now(),
	}, nil
}
