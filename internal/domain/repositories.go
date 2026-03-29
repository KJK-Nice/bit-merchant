package domain

import "github.com/go-webauthn/webauthn/webauthn"

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

// UserRepository defines operations for User persistence.
type UserRepository interface {
	Save(user *User) error
	FindByID(id UserID) (*User, error)
	FindByCredentialID(credentialID []byte) (*User, *webauthn.Credential, error)
	Update(user *User) error
}

// MembershipRepository defines operations for Membership persistence.
type MembershipRepository interface {
	Save(membership *Membership) error
	FindByUserID(userID UserID) ([]*Membership, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*Membership, error)
	FindByUserAndRestaurant(userID UserID, restaurantID RestaurantID) (*Membership, error)
	Delete(id MembershipID) error
}

// InvitationRepository defines operations for Invitation persistence.
type InvitationRepository interface {
	Save(invitation *Invitation) error
	FindByToken(token string) (*Invitation, error)
	FindByRestaurantID(restaurantID RestaurantID) ([]*Invitation, error)
	Update(invitation *Invitation) error
}

// SessionRepository defines operations for Session persistence.
type SessionRepository interface {
	Save(session *Session) error
	Get(id string) (*Session, error)
	Delete(id string) error
	DeleteByUserID(userID UserID) error
}
