package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"bitmerchant/internal/auth/domain/passwordreset"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common/decorator"
)

// ErrInvalidResetToken is returned for an unknown, expired, or already-used link.
var ErrInvalidResetToken = errors.New("this reset link is invalid or has expired")

// PasswordHashStorer hashes a new password for storage.
type PasswordHashStorer interface {
	Hash(plain string) (string, error)
}

// ResetPassword redeems a reset token and sets a new password.
type ResetPassword struct {
	Token       string
	NewPassword string
}

type ResetPasswordHandler decorator.CommandHandler[ResetPassword]

type resetPasswordHandler struct {
	userRepo  user.Repository
	tokenRepo passwordreset.Repository
	hasher    PasswordHashStorer
}

func NewResetPasswordHandler(userRepo user.Repository, tokenRepo passwordreset.Repository, hasher PasswordHashStorer, log *slog.Logger, metrics decorator.MetricsClient) ResetPasswordHandler {
	if userRepo == nil {
		panic("nil user repository")
	}
	h := resetPasswordHandler{userRepo: userRepo, tokenRepo: tokenRepo, hasher: hasher}
	return decorator.ApplyCommandDecorators[ResetPassword](h, log, metrics)
}

func (h resetPasswordHandler) Handle(_ context.Context, cmd ResetPassword) error {
	if len(cmd.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	token, err := h.tokenRepo.FindByHash(HashResetToken(cmd.Token))
	if err != nil || !token.IsUsable(time.Now()) {
		return ErrInvalidResetToken
	}
	u, err := h.userRepo.FindByID(token.UserID)
	if err != nil || u == nil {
		return ErrInvalidResetToken
	}
	hash, err := h.hasher.Hash(cmd.NewPassword)
	if err != nil {
		return err
	}
	u.SetPassword(hash)
	if err := h.userRepo.Update(u); err != nil {
		return err
	}
	token.MarkUsed(time.Now())
	return h.tokenRepo.Update(token)
}
