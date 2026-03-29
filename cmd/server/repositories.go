package main

import (
	"database/sql"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	postgresRepos "bitmerchant/internal/infrastructure/repositories/postgres"
)

type repositories struct {
	Restaurant   domain.RestaurantRepository
	MenuCategory domain.MenuCategoryRepository
	MenuItem     domain.MenuItemRepository
	Order        domain.OrderRepository
	Payment      domain.PaymentRepository
	User         domain.UserRepository
	Membership   domain.MembershipRepository
	Invitation   domain.InvitationRepository
	Session      domain.SessionRepository
}

func newMemoryRepositories() repositories {
	return repositories{
		Restaurant:   memory.NewMemoryRestaurantRepository(),
		MenuCategory: memory.NewMemoryMenuCategoryRepository(),
		MenuItem:     memory.NewMemoryMenuItemRepository(),
		Order:        memory.NewMemoryOrderRepository(),
		Payment:      memory.NewMemoryPaymentRepository(),
		User:         memory.NewMemoryUserRepository(),
		Membership:   memory.NewMemoryMembershipRepository(),
		Invitation:   memory.NewMemoryInvitationRepository(),
		Session:      memory.NewMemorySessionRepository(),
	}
}

func newPostgresRepositories(db *sql.DB) repositories {
	return repositories{
		Restaurant:   postgresRepos.NewRestaurantRepository(db),
		MenuCategory: postgresRepos.NewMenuCategoryRepository(db),
		MenuItem:     postgresRepos.NewMenuItemRepository(db),
		Order:        postgresRepos.NewOrderRepository(db),
		Payment:      postgresRepos.NewPaymentRepository(db),
		User:         postgresRepos.NewUserRepository(db),
		Membership:   postgresRepos.NewMembershipRepository(db),
		Invitation:   postgresRepos.NewInvitationRepository(db),
		Session:      postgresRepos.NewSessionRepository(db),
	}
}
