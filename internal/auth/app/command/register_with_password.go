package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"

	"github.com/google/uuid"
)

var ErrEmailAlreadyTaken = errors.New("email already taken")

// PasswordHasher hashes a plain-text password.
type PasswordHasher interface {
	Hash(plain string) (string, error)
}

// RegisterWithPassword creates a new user via email+password, then establishes
// restaurant context via invitation or new-restaurant creation.
type RegisterWithPassword struct {
	Email           string
	Password        string
	DisplayName     string
	RestaurantName  string
	InvitationToken string
}

type RegisterWithPasswordHandler decorator.CommandResultHandler[RegisterWithPassword, RegistrationOutcome]

type registerWithPasswordHandler struct {
	userRepo         user.Repository
	hasher           PasswordHasher
	acceptInvitation AcceptInvitationForUserHandler
	completeSignup   CompleteSignupNewRestaurantHandler
}

func NewRegisterWithPasswordHandler(
	userRepo user.Repository,
	hasher PasswordHasher,
	acceptInvitation AcceptInvitationForUserHandler,
	completeSignup CompleteSignupNewRestaurantHandler,
	log *slog.Logger,
	metrics decorator.MetricsClient,
) RegisterWithPasswordHandler {
	if userRepo == nil || acceptInvitation == nil || completeSignup == nil {
		panic("nil dependency")
	}
	h := registerWithPasswordHandler{
		userRepo:         userRepo,
		hasher:           hasher,
		acceptInvitation: acceptInvitation,
		completeSignup:   completeSignup,
	}
	return decorator.ApplyCommandResultDecorators[RegisterWithPassword, RegistrationOutcome](h, log, metrics)
}

func (h registerWithPasswordHandler) Handle(ctx context.Context, cmd RegisterWithPassword) (RegistrationOutcome, error) {
	if err := validateRegisterWithPassword(cmd); err != nil {
		return RegistrationOutcome{}, err
	}

	if _, err := h.userRepo.FindByEmail(cmd.Email); err == nil {
		return RegistrationOutcome{}, ErrEmailAlreadyTaken
	}

	hash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		return RegistrationOutcome{}, err
	}

	u, err := user.NewUserWithPassword(common.UserID(uuid.NewString()), cmd.DisplayName, cmd.Email, hash)
	if err != nil {
		return RegistrationOutcome{}, err
	}
	if err := h.userRepo.Save(u); err != nil {
		return RegistrationOutcome{}, err
	}

	if cmd.InvitationToken != "" {
		return h.acceptInvitation.Handle(ctx, AcceptInvitationForUser{
			NewUserID:       u.ID,
			InvitationToken: cmd.InvitationToken,
		})
	}
	return h.completeSignup.Handle(ctx, CompleteSignupNewRestaurant{
		OwnerUserID:    u.ID,
		RestaurantName: cmd.RestaurantName,
	})
}

func validateRegisterWithPassword(cmd RegisterWithPassword) error {
	if cmd.DisplayName == "" {
		return errors.New("display name is required")
	}
	if cmd.Email == "" || !strings.Contains(cmd.Email, "@") {
		return errors.New("valid email is required")
	}
	if len(cmd.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if cmd.InvitationToken == "" && cmd.RestaurantName == "" {
		return errors.New("restaurant name is required for owner signup")
	}
	return nil
}
