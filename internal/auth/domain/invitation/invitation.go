package invitation

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// Invitation grants role-based access to a restaurant.
type Invitation struct {
	ID           common.InvitationID
	RestaurantID common.RestaurantID
	Role         common.MemberRole
	Token        string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	UsedByUserID *common.UserID
	CreatedAt    time.Time
}

func NewInvitation(id common.InvitationID, restaurantID common.RestaurantID, role common.MemberRole, token string, expiresAt time.Time) (*Invitation, error) {
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
		ID: id, RestaurantID: restaurantID, Role: role,
		Token: token, ExpiresAt: expiresAt, CreatedAt: time.Now(),
	}, nil
}

func (i *Invitation) IsExpired(now time.Time) bool { return now.After(i.ExpiresAt) }
func (i *Invitation) IsUsed() bool                 { return i.UsedAt != nil }
func (i *Invitation) MarkUsed(userID common.UserID, now time.Time) {
	i.UsedAt = &now
	i.UsedByUserID = &userID
}
