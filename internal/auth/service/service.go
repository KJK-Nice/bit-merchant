package service

import (
	"log/slog"

	authInfra "bitmerchant/internal/auth/adapters"
	authapp "bitmerchant/internal/auth/app"
	authhttp "bitmerchant/internal/auth/ports/http"
	"bitmerchant/internal/common/http/middleware"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/wiring"
)

// Auth bundles the auth application layer and HTTP adapter.
type Auth struct {
	Application *authapp.Application
	HTTP        *authhttp.AuthHandler
}

// New wires auth bounded-context application and HTTP port.
func New(
	repos wiring.Repositories,
	webauthnSvc *authInfra.WebAuthnService,
	logger *slog.Logger,
	sessionOpts middleware.SessionOptions,
	createRestaurant restaurantCmd.CreateRestaurantHandler,
	resetBaseURL string,
) *Auth {
	if logger == nil {
		logger = slog.Default()
	}
	hasher := authInfra.NewBcryptPasswordHasher()
	mailer := authInfra.NewLoggingMailer(logger)
	app := authapp.NewApplication(repos.User, repos.Membership, repos.Invitation, repos.Session, repos.Restaurant, repos.PasswordResetToken, mailer, resetBaseURL, createRestaurant, hasher, logger, nil)
	return &Auth{
		Application: app,
		HTTP:        authhttp.NewAuthHandler(webauthnSvc, app, logger, sessionOpts),
	}
}
