package passwordreset

import (
	"time"

	"bitmerchant/internal/common"
)

// TTL is how long a password-reset link stays valid.
const TTL = time.Hour

// Token is a single-use password-reset grant. Only the SHA-256 hash of the raw
// token is ever stored; the raw token lives only in the emailed link.
type Token struct {
	ID        string
	UserID    common.UserID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsUsable reports whether the token can still be redeemed at the given instant.
func (t *Token) IsUsable(now time.Time) bool {
	return t != nil && t.UsedAt == nil && now.Before(t.ExpiresAt)
}

// MarkUsed records redemption so the token cannot be replayed.
func (t *Token) MarkUsed(now time.Time) {
	t.UsedAt = &now
}

// Repository persists password-reset tokens, keyed by their hash.
type Repository interface {
	Save(token *Token) error
	FindByHash(tokenHash string) (*Token, error)
	Update(token *Token) error
}
