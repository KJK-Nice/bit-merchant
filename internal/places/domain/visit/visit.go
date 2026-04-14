package visit

import (
	"time"

	"bitmerchant/internal/common"
)

// SessionRestaurantVisit records that a browser session opened a restaurant menu.
type SessionRestaurantVisit struct {
	sessionID      string
	restaurantID   common.RestaurantID
	firstVisitedAt time.Time
	lastVisitedAt  time.Time
}

func NewSessionRestaurantVisit(
	sessionID string,
	restaurantID common.RestaurantID,
	firstVisitedAt time.Time,
	lastVisitedAt time.Time,
) *SessionRestaurantVisit {
	return &SessionRestaurantVisit{
		sessionID:      sessionID,
		restaurantID:   restaurantID,
		firstVisitedAt: firstVisitedAt,
		lastVisitedAt:  lastVisitedAt,
	}
}

func (v *SessionRestaurantVisit) SessionID() string {
	return v.sessionID
}

func (v *SessionRestaurantVisit) RestaurantID() common.RestaurantID {
	return v.restaurantID
}

func (v *SessionRestaurantVisit) FirstVisitedAt() time.Time {
	return v.firstVisitedAt
}

func (v *SessionRestaurantVisit) LastVisitedAt() time.Time {
	return v.lastVisitedAt
}

func (v *SessionRestaurantVisit) Touch(at time.Time) {
	if at.IsZero() {
		return
	}
	if v.firstVisitedAt.IsZero() {
		v.firstVisitedAt = at
	}
	v.lastVisitedAt = at
}

func (v *SessionRestaurantVisit) Clone() *SessionRestaurantVisit {
	if v == nil {
		return nil
	}
	return &SessionRestaurantVisit{
		sessionID:      v.sessionID,
		restaurantID:   v.restaurantID,
		firstVisitedAt: v.firstVisitedAt,
		lastVisitedAt:  v.lastVisitedAt,
	}
}
