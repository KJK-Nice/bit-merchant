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
	RegisterWithPassword        command.RegisterWithPasswordHandler
	LoginWithPassword           command.LoginWithPasswordHandler
}

// Queries are read-side handlers.
type Queries struct {
	InvitationForToken query.InvitationForTokenHandler
	MembershipsForUser query.MembershipsForUserHandler
}

// PasswordHasher abstracts bcrypt so the app layer stays free of crypto deps.
type PasswordHasher interface {
	command.PasswordHasher
	command.PasswordVerifier
}

// NewApplication wires auth application handlers with optional logging/metrics decorators.
func NewApplication(
	userRepo user.Repository,
	memRepo membership.Repository,
	invRepo invitation.Repository,
	sessRepo session.Repository,
	restRepo restaurantdomain.Repository,
	createRestaurant restaurantCmd.CreateRestaurantHandler,
	hasher PasswordHasher,
	log *slog.Logger,
	metrics decorator.MetricsClient,
) *Application {
	acceptInv := command.NewAcceptInvitationForUserHandler(invRepo, memRepo, log, metrics)
	completeSignup := command.NewCompleteSignupNewRestaurantHandler(memRepo, createRestaurant, log, metrics)
	return &Application{
		User:       userRepo,
		Membership: memRepo,
		Invitation: invRepo,
		Session:    sessRepo,
		Restaurant: restRepo,
		Commands: Commands{
			AcceptInvitation:            acceptInv,
			CompleteSignupNewRestaurant: completeSignup,
			CreateRestaurantUnderOwner:  command.NewCreateRestaurantUnderOwnerHandler(memRepo, createRestaurant, log, metrics),
			SwitchActiveRestaurant:      command.NewSwitchActiveRestaurantHandler(memRepo, sessRepo, log, metrics),
			IssueKitchenStaffInvitation: command.NewIssueKitchenStaffInvitationHandler(memRepo, invRepo, log, metrics),
			EndCustomerSession:          command.NewEndCustomerSessionHandler(sessRepo, log, metrics),
			RegisterWithPassword:        command.NewRegisterWithPasswordHandler(userRepo, hasher, acceptInv, completeSignup, log, metrics),
			LoginWithPassword:           command.NewLoginWithPasswordHandler(userRepo, hasher, log, metrics),
		},
		Queries: Queries{
			InvitationForToken: query.NewInvitationForTokenHandler(invRepo, log, metrics),
			MembershipsForUser: query.NewMembershipsForUserHandler(memRepo, log, metrics),
		},
	}
}
