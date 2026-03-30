package domain

import (
	"errors"
	"time"
)

// InvitationID represents a unique invitation identifier.
type InvitationID string

// Invitation grants role-based access to a restaurant.
type Invitation struct {
	ID           InvitationID
	RestaurantID RestaurantID
	Role         MemberRole
	Token        string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	UsedByUserID *UserID
	CreatedAt    time.Time
}

// NewInvitation creates a new invitation.
func NewInvitation(id InvitationID, restaurantID RestaurantID, role MemberRole, token string, expiresAt time.Time) (*Invitation, error) {
	if restaurantID == "" {
		return nil, errors.New("restaurant ID is required")
	}
	if token == "" {
		return nil, errors.New("token is required")
	}
	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expiration must be in the future")
	}

	return &Invitation{
		ID:           id,
		RestaurantID: restaurantID,
		Role:         role,
		Token:        token,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}, nil
}

// IsExpired checks invitation validity.
func (i *Invitation) IsExpired(now time.Time) bool {
	return now.After(i.ExpiresAt)
}

// IsUsed returns true when the invitation was already consumed.
func (i *Invitation) IsUsed() bool {
	return i.UsedAt != nil
}

// MarkUsed marks the invitation as consumed.
func (i *Invitation) MarkUsed(userID UserID, now time.Time) {
	i.UsedAt = &now
	i.UsedByUserID = &userID
}
