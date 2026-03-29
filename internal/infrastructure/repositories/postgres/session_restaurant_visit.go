package postgres

import (
	"context"
	"database/sql"
	"time"

	"bitmerchant/internal/domain"
)

// SessionRestaurantVisitRepository implements SessionRestaurantVisitRepository for PostgreSQL.
type SessionRestaurantVisitRepository struct {
	db *sql.DB
}

// NewSessionRestaurantVisitRepository constructs the repository.
func NewSessionRestaurantVisitRepository(db *sql.DB) *SessionRestaurantVisitRepository {
	return &SessionRestaurantVisitRepository{db: db}
}

// Upsert inserts or bumps last_visited_at.
func (r *SessionRestaurantVisitRepository) Upsert(visit *domain.SessionRestaurantVisit) error {
	if visit == nil || visit.SessionID == "" || visit.RestaurantID == "" {
		return nil
	}
	now := time.Now()
	first := visit.FirstVisitedAt
	if first.IsZero() {
		first = now
	}
	last := visit.LastVisitedAt
	if last.IsZero() {
		last = now
	}
	_, err := r.db.ExecContext(context.Background(),
		`INSERT INTO session_restaurant_visits (session_id, restaurant_id, first_visited_at, last_visited_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (session_id, restaurant_id) DO UPDATE
		 SET last_visited_at = EXCLUDED.last_visited_at`,
		visit.SessionID, string(visit.RestaurantID), first, last,
	)
	return err
}

// FindBySessionID returns visits ordered by last_visited_at descending.
func (r *SessionRestaurantVisitRepository) FindBySessionID(sessionID string) ([]*domain.SessionRestaurantVisit, error) {
	rows, err := r.db.QueryContext(context.Background(),
		`SELECT session_id, restaurant_id, first_visited_at, last_visited_at
		   FROM session_restaurant_visits
		  WHERE session_id = $1
		  ORDER BY last_visited_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.SessionRestaurantVisit
	for rows.Next() {
		var (
			sid, rid string
			first    time.Time
			last     time.Time
		)
		if err := rows.Scan(&sid, &rid, &first, &last); err != nil {
			return nil, err
		}
		out = append(out, &domain.SessionRestaurantVisit{
			SessionID:      sid,
			RestaurantID:   domain.RestaurantID(rid),
			FirstVisitedAt: first,
			LastVisitedAt:  last,
		})
	}
	return out, rows.Err()
}
