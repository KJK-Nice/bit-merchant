package visit

import (
	"time"

	"bitmerchant/internal/common"
)

// SessionRestaurantVisit records that a browser session opened a restaurant menu.
type SessionRestaurantVisit struct {
	SessionID      string
	RestaurantID   common.RestaurantID
	FirstVisitedAt time.Time
	LastVisitedAt  time.Time
}
