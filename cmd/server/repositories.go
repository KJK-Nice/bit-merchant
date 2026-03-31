package main

import (
	"database/sql"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/payment/domain/payment"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"

	authAdapters "bitmerchant/internal/auth/adapters"
	menuAdapters "bitmerchant/internal/menu/adapters"
	orderAdapters "bitmerchant/internal/ordering/adapters"
	payAdapters "bitmerchant/internal/payment/adapters"
	placesAdapters "bitmerchant/internal/places/adapters"
	restAdapters "bitmerchant/internal/restaurant/adapters"
)

type repositories struct {
	Restaurant              restaurant.Repository
	MenuCategory            menu.CategoryRepository
	MenuItem                menu.ItemRepository
	Order                   order.Repository
	Payment                 payment.Repository
	User                    user.Repository
	Membership              membership.Repository
	Invitation              invitation.Repository
	Session                 session.Repository
	SessionRestaurantVisits visit.Repository
}

func newMemoryRepositories() repositories {
	return repositories{
		Restaurant:              restAdapters.NewMemoryRestaurantRepository(),
		MenuCategory:            menuAdapters.NewMemoryCategoryRepository(),
		MenuItem:                menuAdapters.NewMemoryItemRepository(),
		Order:                   orderAdapters.NewMemoryOrderRepository(),
		Payment:                 payAdapters.NewMemoryPaymentRepository(),
		User:                    authAdapters.NewMemoryUserRepository(),
		Membership:              authAdapters.NewMemoryMembershipRepository(),
		Invitation:              authAdapters.NewMemoryInvitationRepository(),
		Session:                 authAdapters.NewMemorySessionRepository(),
		SessionRestaurantVisits: placesAdapters.NewMemoryVisitRepository(),
	}
}

func newPostgresRepositories(db *sql.DB) repositories {
	return repositories{
		Restaurant:              restAdapters.NewPostgresRestaurantRepository(db),
		MenuCategory:            menuAdapters.NewPostgresCategoryRepository(db),
		MenuItem:                menuAdapters.NewPostgresItemRepository(db),
		Order:                   orderAdapters.NewPostgresOrderRepository(db),
		Payment:                 payAdapters.NewPostgresPaymentRepository(db),
		User:                    authAdapters.NewPostgresUserRepository(db),
		Membership:              authAdapters.NewPostgresMembershipRepository(db),
		Invitation:              authAdapters.NewPostgresInvitationRepository(db),
		Session:                 authAdapters.NewPostgresSessionRepository(db),
		SessionRestaurantVisits: placesAdapters.NewPostgresVisitRepository(db),
	}
}
