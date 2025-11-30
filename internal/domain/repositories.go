package domain

// RestaurantRepository defines operations for Restaurant persistence
type RestaurantRepository interface {
	Save(restaurant *Restaurant) error
	FindByID(id RestaurantID) (*Restaurant, error)
	Update(restaurant *Restaurant) error
}

// MenuCategoryRepository defines operations for MenuCategory persistence
type MenuCategoryRepository interface {
	Save(category *MenuCategory) error
	FindByID(id CategoryID) (*MenuCategory, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*MenuCategory, error)
	Update(category *MenuCategory) error
	Delete(id CategoryID) error
}

// MenuItemRepository defines operations for MenuItem persistence
type MenuItemRepository interface {
	Save(item *MenuItem) error
	FindByID(id ItemID) (*MenuItem, error)
	FindByCategoryID(categoryID CategoryID) ([]*MenuItem, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*MenuItem, error)
	FindAvailableByRestaurantID(restaurantID RestaurantID) ([]*MenuItem, error)
	Update(item *MenuItem) error
	Delete(id ItemID) error
	CountByRestaurantID(restaurantID RestaurantID) (int, error)
}

// OrderRepository defines operations for Order persistence
type OrderRepository interface {
	Save(order *Order) error
	FindByID(id OrderID) (*Order, error)
	FindByOrderNumber(restaurantID RestaurantID, orderNumber string) (*Order, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*Order, error)
	FindActiveByRestaurantID(restaurantID RestaurantID) ([]*Order, error)
	FindBySessionID(sessionID string) ([]*Order, error)
	Update(order *Order) error
}

// PaymentRepository defines operations for Payment persistence
type PaymentRepository interface {
	Save(payment *Payment) error
	FindByID(id PaymentID) (*Payment, error)
	FindByOrderID(orderID OrderID) (*Payment, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*Payment, error)
	Update(payment *Payment) error
}
