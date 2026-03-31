package payment

import "bitmerchant/internal/common"

// Repository defines operations for Payment persistence.
type Repository interface {
	Save(payment *Payment) error
	FindByID(id common.PaymentID) (*Payment, error)
	FindByOrderID(orderID common.OrderID) (*Payment, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Payment, error)
	Update(payment *Payment) error
}
