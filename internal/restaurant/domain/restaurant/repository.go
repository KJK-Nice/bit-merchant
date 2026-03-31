package restaurant

import "bitmerchant/internal/common"

// Repository defines operations for Restaurant persistence.
type Repository interface {
	Save(restaurant *Restaurant) error
	FindByID(id common.RestaurantID) (*Restaurant, error)
	Update(restaurant *Restaurant) error
}
