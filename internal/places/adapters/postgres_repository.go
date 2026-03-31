package adapters

import (
	"database/sql"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/places/domain/visit"
)

type PostgresVisitRepository struct {
	db *sql.DB
}

func NewPostgresVisitRepository(db *sql.DB) *PostgresVisitRepository {
	return &PostgresVisitRepository{db: db}
}

func (r *PostgresVisitRepository) Upsert(v *visit.SessionRestaurantVisit) error {
	_, err := r.db.Exec(
		`INSERT INTO session_restaurant_visits (session_id, restaurant_id, first_visited_at, last_visited_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (session_id, restaurant_id) DO UPDATE SET last_visited_at = EXCLUDED.last_visited_at`,
		v.SessionID, string(v.RestaurantID), v.FirstVisitedAt, v.LastVisitedAt)
	return err
}

func (r *PostgresVisitRepository) FindBySessionID(sessionID string) ([]*visit.SessionRestaurantVisit, error) {
	rows, err := r.db.Query(
		`SELECT session_id, restaurant_id, first_visited_at, last_visited_at
		 FROM session_restaurant_visits WHERE session_id = $1 ORDER BY last_visited_at DESC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*visit.SessionRestaurantVisit
	for rows.Next() {
		var (
			sid, rid                  string
			firstVisited, lastVisited time.Time
		)
		if err := rows.Scan(&sid, &rid, &firstVisited, &lastVisited); err != nil {
			return nil, err
		}
		result = append(result, &visit.SessionRestaurantVisit{
			SessionID: sid, RestaurantID: common.RestaurantID(rid),
			FirstVisitedAt: firstVisited, LastVisitedAt: lastVisited,
		})
	}
	return result, rows.Err()
}
