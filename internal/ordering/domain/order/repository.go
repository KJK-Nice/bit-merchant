package order

import "bitmerchant/internal/common"

// Repository defines operations for Order persistence.
type Repository interface {
	Save(order *Order) error
	FindByID(id common.OrderID) (*Order, error)
	FindByOrderNumber(restaurantID common.RestaurantID, orderNumber string) (*Order, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Order, error)
	FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*Order, error)
	FindBySessionID(sessionID string) ([]*Order, error)
	Update(order *Order) error
	// NextOrderNumber atomically allocates the next order number for the given
	// restaurant. The returned value is monotonically increasing within a
	// restaurant and is safe to call concurrently — implementations must
	// guarantee no two callers ever receive the same number.
	NextOrderNumber(restaurantID common.RestaurantID) (int, error)
}
