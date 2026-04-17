package command

import (
	"context"
	"errors"
	"log/slog"

	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common/decorator"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// PasswordVerifier checks a bcrypt hash against a plain-text password.
type PasswordVerifier interface {
	Verify(hash, plain string) error
}

// LoginWithPassword authenticates via email+password and returns the matched user.
type LoginWithPassword struct {
	Email    string
	Password string
}

// LoginWithPasswordResult holds the authenticated user.
type LoginWithPasswordResult struct {
	User *user.User
}

type LoginWithPasswordHandler decorator.CommandResultHandler[LoginWithPassword, LoginWithPasswordResult]

type loginWithPasswordHandler struct {
	userRepo user.Repository
	hasher   PasswordVerifier
}

func NewLoginWithPasswordHandler(userRepo user.Repository, hasher PasswordVerifier, log *slog.Logger, metrics decorator.MetricsClient) LoginWithPasswordHandler {
	if userRepo == nil {
		panic("nil dependency")
	}
	h := loginWithPasswordHandler{userRepo: userRepo, hasher: hasher}
	return decorator.ApplyCommandResultDecorators[LoginWithPassword, LoginWithPasswordResult](h, log, metrics)
}

func (h loginWithPasswordHandler) Handle(_ context.Context, cmd LoginWithPassword) (LoginWithPasswordResult, error) {
	u, err := h.userRepo.FindByEmail(cmd.Email)
	if err != nil {
		// Return same error regardless of whether user exists (no enumeration).
		return LoginWithPasswordResult{}, ErrInvalidCredentials
	}
	if u.PasswordHash == "" {
		return LoginWithPasswordResult{}, ErrInvalidCredentials
	}
	if err := h.hasher.Verify(u.PasswordHash, cmd.Password); err != nil {
		return LoginWithPasswordResult{}, ErrInvalidCredentials
	}
	return LoginWithPasswordResult{User: u}, nil
}
