package domain

import "time"

// SessionRestaurantVisit records that a browser session opened a restaurant menu (customer "stamp").
type SessionRestaurantVisit struct {
	SessionID      string
	RestaurantID   RestaurantID
	FirstVisitedAt time.Time
	LastVisitedAt  time.Time
}
