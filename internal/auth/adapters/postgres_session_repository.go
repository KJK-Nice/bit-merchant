package adapters

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common"
)

type PostgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Save(s *session.Session) error {
	var userID, restID *string
	if s.UserID != nil {
		v := string(*s.UserID)
		userID = &v
	}
	if s.RestaurantID != nil {
		v := string(*s.RestaurantID)
		restID = &v
	}
	_, err := r.db.Exec(
		`INSERT INTO auth_sessions (id, user_id, restaurant_id, created_at, expires_at)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (id) DO UPDATE SET user_id=EXCLUDED.user_id, restaurant_id=EXCLUDED.restaurant_id, expires_at=EXCLUDED.expires_at`,
		s.ID, userID, restID, s.CreatedAt, s.ExpiresAt)
	return err
}

func (r *PostgresSessionRepository) Get(id string) (*session.Session, error) {
	var sid string
	var userID, restID sql.NullString
	var createdAt, expiresAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, created_at, expires_at FROM auth_sessions WHERE id = $1`, id).
		Scan(&sid, &userID, &restID, &createdAt, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	s := &session.Session{ID: sid, CreatedAt: createdAt.Time, ExpiresAt: expiresAt.Time}
	if userID.Valid {
		uid := common.UserID(userID.String)
		s.UserID = &uid
	}
	if restID.Valid {
		rid := common.RestaurantID(restID.String)
		s.RestaurantID = &rid
	}
	return s, nil
}

func (r *PostgresSessionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM auth_sessions WHERE id = $1`, id)
	return err
}

func (r *PostgresSessionRepository) DeleteByUserID(userID common.UserID) error {
	_, err := r.db.Exec(`DELETE FROM auth_sessions WHERE user_id = $1`, string(userID))
	return err
}
