package domain

import (
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/payment/domain/payment"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

// Repository interfaces aliased from bounded contexts.
// New code should import directly from the bounded context domain packages.
type RestaurantRepository = restaurant.Repository
type MenuCategoryRepository = menu.CategoryRepository
type MenuItemRepository = menu.ItemRepository
type OrderRepository = order.Repository
type PaymentRepository = payment.Repository
type UserRepository = user.Repository
type MembershipRepository = membership.Repository
type InvitationRepository = invitation.Repository
type SessionRepository = session.Repository
type SessionRestaurantVisitRepository = visit.Repository
