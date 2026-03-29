package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// RestaurantRepository implements domain.RestaurantRepository for PostgreSQL.
type RestaurantRepository struct {
	db *sql.DB
}

func NewRestaurantRepository(db *sql.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) Save(restaurant *domain.Restaurant) error {
	_, err := r.db.Exec(
		`INSERT INTO restaurants (id, name, table_count, is_open, closed_message, reopening_hours, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (id) DO UPDATE
		 SET name = EXCLUDED.name,
		     table_count = EXCLUDED.table_count,
		     is_open = EXCLUDED.is_open,
		     closed_message = EXCLUDED.closed_message,
		     reopening_hours = EXCLUDED.reopening_hours,
		     updated_at = EXCLUDED.updated_at`,
		string(restaurant.ID),
		restaurant.Name,
		restaurant.TableCount,
		restaurant.IsOpen,
		restaurant.ClosedMessage,
		restaurant.ReopeningHours,
		restaurant.CreatedAt,
		restaurant.UpdatedAt,
	)
	return err
}

func (r *RestaurantRepository) FindByID(id domain.RestaurantID) (*domain.Restaurant, error) {
	row := r.db.QueryRow(
		`SELECT id, name, table_count, is_open, closed_message, reopening_hours, created_at, updated_at
		   FROM restaurants
		  WHERE id = $1`,
		string(id),
	)

	var (
		restaurantID   string
		name           string
		tableCount     int
		isOpen         bool
		closedMessage  string
		reopeningHours string
		createdAt      sql.NullTime
		updatedAt      sql.NullTime
	)
	if err := row.Scan(&restaurantID, &name, &tableCount, &isOpen, &closedMessage, &reopeningHours, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("restaurant not found")
		}
		return nil, err
	}

	return &domain.Restaurant{
		ID:             domain.RestaurantID(restaurantID),
		Name:           name,
		TableCount:     tableCount,
		IsOpen:         isOpen,
		ClosedMessage:  closedMessage,
		ReopeningHours: reopeningHours,
		CreatedAt:      createdAt.Time,
		UpdatedAt:      updatedAt.Time,
	}, nil
}

func (r *RestaurantRepository) Update(restaurant *domain.Restaurant) error {
	result, err := r.db.Exec(
		`UPDATE restaurants
		    SET name = $2,
		        table_count = $3,
		        is_open = $4,
		        closed_message = $5,
		        reopening_hours = $6,
		        updated_at = $7
		  WHERE id = $1`,
		string(restaurant.ID),
		restaurant.Name,
		restaurant.TableCount,
		restaurant.IsOpen,
		restaurant.ClosedMessage,
		restaurant.ReopeningHours,
		restaurant.UpdatedAt,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("restaurant not found")
	}
	return nil
}
