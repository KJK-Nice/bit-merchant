package invitation

import "bitmerchant/internal/common"

// Repository defines operations for Invitation persistence.
type Repository interface {
	Save(invitation *Invitation) error
	FindByToken(token string) (*Invitation, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*Invitation, error)
	Update(invitation *Invitation) error
}
