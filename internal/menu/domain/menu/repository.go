package menu

import "bitmerchant/internal/common"

// CategoryRepository defines operations for MenuCategory persistence.
type CategoryRepository interface {
	Save(category *MenuCategory) error
	FindByID(id common.CategoryID) (*MenuCategory, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*MenuCategory, error)
	Update(category *MenuCategory) error
	Delete(id common.CategoryID) error
}

// ItemRepository defines operations for MenuItem persistence.
type ItemRepository interface {
	Save(item *MenuItem) error
	FindByID(id common.ItemID) (*MenuItem, error)
	FindByCategoryID(categoryID common.CategoryID) ([]*MenuItem, error)
	FindByRestaurantID(restaurantID common.RestaurantID) ([]*MenuItem, error)
	FindAvailableByRestaurantID(restaurantID common.RestaurantID) ([]*MenuItem, error)
	Update(item *MenuItem) error
	Delete(id common.ItemID) error
	CountByRestaurantID(restaurantID common.RestaurantID) (int, error)
}
