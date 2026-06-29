package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	authAdapters "bitmerchant/internal/auth/adapters"
	authCmd "bitmerchant/internal/auth/app/command"
	"bitmerchant/internal/auth/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureMailer struct {
	calls int
	url   string
}

func (m *captureMailer) SendPasswordReset(_ context.Context, _ string, resetURL string) error {
	m.calls++
	m.url = resetURL
	return nil
}

func seedPasswordUser(t *testing.T, repo user.Repository, hasher *authAdapters.BcryptPasswordHasher, email, password string) *user.User {
	t.Helper()
	hash, err := hasher.Hash(password)
	require.NoError(t, err)
	u, err := user.NewUserWithPassword("u1", "Owner", email, hash)
	require.NoError(t, err)
	require.NoError(t, repo.Save(u))
	return u
}

func rawTokenFromURL(t *testing.T, url string) string {
	t.Helper()
	i := strings.LastIndex(url, "/auth/reset-password/")
	require.GreaterOrEqual(t, i, 0, "reset url should contain the token path")
	return url[i+len("/auth/reset-password/"):]
}

func TestPasswordResetFlow(t *testing.T) {
	userRepo := authAdapters.NewMemoryUserRepository()
	tokenRepo := authAdapters.NewMemoryPasswordResetTokenRepository()
	hasher := authAdapters.NewBcryptPasswordHasher()
	mailer := &captureMailer{}

	seedPasswordUser(t, userRepo, hasher, "owner@example.com", "oldpass123")

	requestUC := authCmd.NewRequestPasswordResetHandler(userRepo, tokenRepo, mailer, "https://merchant.test", nil, nil)
	resetUC := authCmd.NewResetPasswordHandler(userRepo, tokenRepo, hasher, nil, nil)

	t.Run("unknown email is a no-op success (no enumeration)", func(t *testing.T) {
		require.NoError(t, requestUC.Handle(context.Background(), authCmd.RequestPasswordReset{Email: "nobody@example.com"}))
		assert.Equal(t, 0, mailer.calls, "no email sent for unknown address")
	})

	t.Run("known email sends a reset link", func(t *testing.T) {
		require.NoError(t, requestUC.Handle(context.Background(), authCmd.RequestPasswordReset{Email: "owner@example.com"}))
		require.Equal(t, 1, mailer.calls)
		assert.True(t, strings.HasPrefix(mailer.url, "https://merchant.test/auth/reset-password/"))
	})

	raw := rawTokenFromURL(t, mailer.url)

	t.Run("only the hash is stored, never the raw token", func(t *testing.T) {
		_, err := tokenRepo.FindByHash(raw)
		assert.Error(t, err, "raw token must not be a key")
		_, err = tokenRepo.FindByHash(authCmd.HashResetToken(raw))
		assert.NoError(t, err, "hashed token is stored")
	})

	t.Run("redeeming sets the new password", func(t *testing.T) {
		require.NoError(t, resetUC.Handle(context.Background(), authCmd.ResetPassword{Token: raw, NewPassword: "brandnew123"}))
		u, err := userRepo.FindByEmail("owner@example.com")
		require.NoError(t, err)
		assert.NoError(t, hasher.Verify(u.PasswordHash, "brandnew123"), "new password verifies")
		assert.Error(t, hasher.Verify(u.PasswordHash, "oldpass123"), "old password no longer works")
	})

	t.Run("token is single-use", func(t *testing.T) {
		err := resetUC.Handle(context.Background(), authCmd.ResetPassword{Token: raw, NewPassword: "another123"})
		assert.ErrorIs(t, err, authCmd.ErrInvalidResetToken)
	})

	t.Run("rejects short passwords", func(t *testing.T) {
		mailer.calls = 0
		require.NoError(t, requestUC.Handle(context.Background(), authCmd.RequestPasswordReset{Email: "owner@example.com"}))
		fresh := rawTokenFromURL(t, mailer.url)
		err := resetUC.Handle(context.Background(), authCmd.ResetPassword{Token: fresh, NewPassword: "short"})
		assert.Error(t, err)
	})

	t.Run("rejects expired tokens", func(t *testing.T) {
		mailer.calls = 0
		require.NoError(t, requestUC.Handle(context.Background(), authCmd.RequestPasswordReset{Email: "owner@example.com"}))
		fresh := rawTokenFromURL(t, mailer.url)
		// Force expiry in the store.
		tok, err := tokenRepo.FindByHash(authCmd.HashResetToken(fresh))
		require.NoError(t, err)
		past := time.Now().Add(-time.Minute)
		tok.ExpiresAt = past
		require.NoError(t, tokenRepo.Update(tok))
		err = resetUC.Handle(context.Background(), authCmd.ResetPassword{Token: fresh, NewPassword: "valid12345"})
		assert.ErrorIs(t, err, authCmd.ErrInvalidResetToken)
	})
}
