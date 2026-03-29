package domain

import "time"

// Session represents browser/server auth state.
type Session struct {
	ID           string
	UserID       *UserID
	RestaurantID *RestaurantID
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// IsExpired checks whether the session is no longer valid.
func (s *Session) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt)
}
