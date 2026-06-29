package command

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"bitmerchant/internal/auth/domain/passwordreset"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common/decorator"

	"github.com/google/uuid"
)

// Mailer delivers the password-reset link. Implemented by the dev LoggingMailer
// (logs the link) or a real provider in production.
type Mailer interface {
	SendPasswordReset(ctx context.Context, email, resetURL string) error
}

// HashResetToken returns the stored hash for a raw reset token. Only the hash is
// persisted; the raw token travels in the emailed link.
func HashResetToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RequestPasswordReset starts the forgot-password flow for an email address.
type RequestPasswordReset struct {
	Email string
}

type RequestPasswordResetHandler decorator.CommandHandler[RequestPasswordReset]

type requestPasswordResetHandler struct {
	userRepo  user.Repository
	tokenRepo passwordreset.Repository
	mailer    Mailer
	baseURL   string
	logger    *slog.Logger
}

func NewRequestPasswordResetHandler(userRepo user.Repository, tokenRepo passwordreset.Repository, mailer Mailer, baseURL string, log *slog.Logger, metrics decorator.MetricsClient) RequestPasswordResetHandler {
	if userRepo == nil {
		panic("nil user repository")
	}
	if log == nil {
		log = slog.Default()
	}
	h := requestPasswordResetHandler{userRepo: userRepo, tokenRepo: tokenRepo, mailer: mailer, baseURL: strings.TrimRight(baseURL, "/"), logger: log}
	return decorator.ApplyCommandDecorators[RequestPasswordReset](h, log, metrics)
}

// Handle always succeeds (returns nil) whether or not the email matches an
// account — this prevents account enumeration. A reset link is only generated,
// stored, and sent when a password user actually exists for the address.
func (h requestPasswordResetHandler) Handle(ctx context.Context, cmd RequestPasswordReset) error {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	u, err := h.userRepo.FindByEmail(email)
	if err != nil || u == nil || u.PasswordHash == "" {
		// Unknown email or passkey-only account: respond as success, do nothing.
		h.logger.Info("password reset requested for unknown/ineligible account", "email", email)
		return nil
	}

	rawToken, terr := randomToken()
	if terr != nil {
		return terr
	}
	now := time.Now()
	token := &passwordreset.Token{
		ID:        uuid.New().String(),
		UserID:    u.ID,
		TokenHash: HashResetToken(rawToken),
		ExpiresAt: now.Add(passwordreset.TTL),
		CreatedAt: now,
	}
	if err := h.tokenRepo.Save(token); err != nil {
		return err
	}

	resetURL := h.baseURL + "/auth/reset-password/" + rawToken
	if err := h.mailer.SendPasswordReset(ctx, email, resetURL); err != nil {
		h.logger.Error("failed to send password reset email", "error", err, "userID", u.ID)
		return err
	}
	return nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
