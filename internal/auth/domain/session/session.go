package session

import (
	"time"

	"bitmerchant/internal/common"
)

// Session represents browser/server auth state.
type Session struct {
	ID           string
	UserID       *common.UserID
	RestaurantID *common.RestaurantID
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func (s *Session) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt)
}
