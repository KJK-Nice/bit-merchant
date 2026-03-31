package membership

import "bitmerchant/internal/common"

// Repository defines operations for Membership persistence.
type Repository interface {
	Save(membership *Membership) error
	FindByUserID(userID common.UserID) ([]*Membership, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Membership, error)
	FindByUserAndRestaurant(userID common.UserID, restaurantID common.RestaurantID) (*Membership, error)
	Delete(id common.MembershipID) error
}
