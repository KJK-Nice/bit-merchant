package query

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

// OrderReadModel is the read-side dependency for dashboard analytics.
type OrderReadModel interface {
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error)
	FindActiveByRestaurantID(restaurantID common.RestaurantID) ([]*order.Order, error)
}
