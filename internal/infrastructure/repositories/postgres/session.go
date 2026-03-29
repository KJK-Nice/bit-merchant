package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// SessionRepository implements domain.SessionRepository for PostgreSQL.
type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Save(session *domain.Session) error {
	_, err := r.db.Exec(
		`INSERT INTO auth_sessions (id, user_id, restaurant_id, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE
		 SET user_id = EXCLUDED.user_id,
		     restaurant_id = EXCLUDED.restaurant_id,
		     created_at = EXCLUDED.created_at,
		     expires_at = EXCLUDED.expires_at`,
		session.ID,
		userIDString(session.UserID),
		restaurantIDString(session.RestaurantID),
		session.CreatedAt,
		session.ExpiresAt,
	)
	return err
}

func (r *SessionRepository) Get(id string) (*domain.Session, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, created_at, expires_at
		   FROM auth_sessions
		  WHERE id = $1`,
		id,
	)

	var (
		sessionID    string
		userID       sql.NullString
		restaurantID sql.NullString
		createdAt    sql.NullTime
		expiresAt    sql.NullTime
	)
	if err := row.Scan(&sessionID, &userID, &restaurantID, &createdAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	var userIDPtr *domain.UserID
	if userID.Valid && userID.String != "" {
		uid := domain.UserID(userID.String)
		userIDPtr = &uid
	}

	var restaurantIDPtr *domain.RestaurantID
	if restaurantID.Valid && restaurantID.String != "" {
		rid := domain.RestaurantID(restaurantID.String)
		restaurantIDPtr = &rid
	}

	return &domain.Session{
		ID:           sessionID,
		UserID:       userIDPtr,
		RestaurantID: restaurantIDPtr,
		CreatedAt:    createdAt.Time,
		ExpiresAt:    expiresAt.Time,
	}, nil
}

func (r *SessionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM auth_sessions WHERE id = $1`, id)
	return err
}

func (r *SessionRepository) DeleteByUserID(userID domain.UserID) error {
	_, err := r.db.Exec(`DELETE FROM auth_sessions WHERE user_id = $1`, string(userID))
	return err
}

func userIDString(userID *domain.UserID) *string {
	if userID == nil {
		return nil
	}
	s := string(*userID)
	return &s
}

func restaurantIDString(restaurantID *domain.RestaurantID) *string {
	if restaurantID == nil {
		return nil
	}
	s := string(*restaurantID)
	return &s
}
