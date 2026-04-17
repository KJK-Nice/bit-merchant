package app

import (
	"log/slog"

	"bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/app/query"
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common/decorator"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	restaurantdomain "bitmerchant/internal/restaurant/domain/restaurant"
)

// Application bundles auth bounded-context handlers and repositories used by HTTP adapters.
type Application struct {
	User       user.Repository
	Membership membership.Repository
	Invitation invitation.Repository
	Session    session.Repository
	Restaurant restaurantdomain.Repository

	Commands Commands
	Queries  Queries
}

// Commands are write-side handlers.
type Commands struct {
	AcceptInvitation            command.AcceptInvitationForUserHandler
	CompleteSignupNewRestaurant command.CompleteSignupNewRestaurantHandler
	CreateRestaurantUnderOwner  command.CreateRestaurantUnderOwnerHandler
	SwitchActiveRestaurant      command.SwitchActiveRestaurantHandler
	IssueKitchenStaffInvitation command.IssueKitchenStaffInvitationHandler
	EndCustomerSession          command.EndCustomerSessionHandler
}

// Queries are read-side handlers.
type Queries struct {
	InvitationForToken query.InvitationForTokenHandler
	MembershipsForUser query.MembershipsForUserHandler
}

// NewApplication wires auth application handlers with optional logging/metrics decorators.
func NewApplication(
	userRepo user.Repository,
	memRepo membership.Repository,
	invRepo invitation.Repository,
	sessRepo session.Repository,
	restRepo restaurantdomain.Repository,
	createRestaurant restaurantCmd.CreateRestaurantHandler,
	log *slog.Logger,
	metrics decorator.MetricsClient,
) *Application {
	return &Application{
		User:       userRepo,
		Membership: memRepo,
		Invitation: invRepo,
		Session:    sessRepo,
		Restaurant: restRepo,
		Commands: Commands{
			AcceptInvitation:            command.NewAcceptInvitationForUserHandler(invRepo, memRepo, log, metrics),
			CompleteSignupNewRestaurant: command.NewCompleteSignupNewRestaurantHandler(memRepo, createRestaurant, log, metrics),
			CreateRestaurantUnderOwner:  command.NewCreateRestaurantUnderOwnerHandler(memRepo, createRestaurant, log, metrics),
			SwitchActiveRestaurant:      command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, log, metrics),
			IssueKitchenStaffInvitation: command.NewIssueKitchenStaffInvitationHandler(memRepo, invRepo, log, metrics),
			EndCustomerSession:          command.NewEndCustomerSessionHandler(sessRepo, log, metrics),
		},
		Queries: Queries{
			InvitationForToken: query.NewInvitationForTokenHandler(invRepo, log, metrics),
			MembershipsForUser: query.NewMembershipsForUserHandler(memRepo, log, metrics),
		},
	}
}
