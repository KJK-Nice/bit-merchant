package wiring

import (
	"context"
	"database/sql"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/passwordreset"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
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

// Repositories bundles concrete repository implementations used across bounded contexts.
type Repositories struct {
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
	PasswordResetToken      passwordreset.Repository
}

// NewMemoryRepositories wires in-memory repositories (tests and local dev without Postgres).
func NewMemoryRepositories() Repositories {
	return Repositories{
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
		PasswordResetToken:      authAdapters.NewMemoryPasswordResetTokenRepository(),
	}
}

// NewPostgresRepositories wires Postgres-backed repositories.
func NewPostgresRepositories(db *sql.DB) Repositories {
	return Repositories{
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
		PasswordResetToken:      authAdapters.NewPostgresPasswordResetTokenRepository(db),
	}
}

// SeedData loads demo menu data for the default restaurant (memory and fresh DB dev flows).
func SeedData(ctx context.Context, repos Repositories) {
	_ = ctx

	restaurantID := common.RestaurantID("restaurant_1")
	restaurantObj, _ := restaurant.NewRestaurant(restaurantID, "BitMerchant Cafe")
	_ = repos.Restaurant.Save(restaurantObj)

	cat1, _ := menu.NewMenuCategory("cat_1", restaurantID, "Appetizers", 1)
	cat2, _ := menu.NewMenuCategory("cat_2", restaurantID, "Mains", 2)
	cat3, _ := menu.NewMenuCategory("cat_3", restaurantID, "Drinks", 3)
	_ = repos.MenuCategory.Save(cat1)
	_ = repos.MenuCategory.Save(cat2)
	_ = repos.MenuCategory.Save(cat3)

	item1, _ := menu.NewMenuItem("item_1", "cat_1", restaurantID, "Bruschetta", 8.50)
	_ = item1.SetDescription("Toasted bread with tomatoes and basil")
	_ = repos.MenuItem.Save(item1)

	item2, _ := menu.NewMenuItem("item_2", "cat_2", restaurantID, "Bitcoin Burger", 15.00)
	_ = item2.SetDescription("Premium beef patty with cheese")
	_ = repos.MenuItem.Save(item2)

	item3, _ := menu.NewMenuItem("item_3", "cat_3", restaurantID, "Satoshi Soda", 3.00)
	_ = repos.MenuItem.Save(item3)
}
