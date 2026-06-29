package adapters

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/auth/domain/passwordreset"
	"bitmerchant/internal/common"
)

type PostgresPasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPostgresPasswordResetTokenRepository(db *sql.DB) *PostgresPasswordResetTokenRepository {
	return &PostgresPasswordResetTokenRepository{db: db}
}

func (r *PostgresPasswordResetTokenRepository) Save(token *passwordreset.Token) error {
	var usedAt interface{}
	if token.UsedAt != nil {
		usedAt = *token.UsedAt
	}
	_, err := r.db.Exec(
		`INSERT INTO auth_password_reset_tokens (id, user_id, token_hash, expires_at, used_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		token.ID, string(token.UserID), token.TokenHash, token.ExpiresAt, usedAt, token.CreatedAt,
	)
	return err
}

func (r *PostgresPasswordResetTokenRepository) FindByHash(tokenHash string) (*passwordreset.Token, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM auth_password_reset_tokens WHERE token_hash = $1`, tokenHash)
	var (
		id, userID, hash string
		expiresAt        time.Time
		usedAt           sql.NullTime
		createdAt        time.Time
	)
	if err := row.Scan(&id, &userID, &hash, &expiresAt, &usedAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reset token not found")
		}
		return nil, err
	}
	t := &passwordreset.Token{
		ID:        id,
		UserID:    common.UserID(userID),
		TokenHash: hash,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}
	if usedAt.Valid {
		u := usedAt.Time
		t.UsedAt = &u
	}
	return t, nil
}

func (r *PostgresPasswordResetTokenRepository) Update(token *passwordreset.Token) error {
	var usedAt interface{}
	if token.UsedAt != nil {
		usedAt = *token.UsedAt
	}
	_, err := r.db.Exec(
		`UPDATE auth_password_reset_tokens SET used_at = $2 WHERE id = $1`,
		token.ID, usedAt)
	return err
}
