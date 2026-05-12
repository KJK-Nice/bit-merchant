package adapters

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type PostgresRestaurantRepository struct {
	db *sql.DB
}

func NewPostgresRestaurantRepository(db *sql.DB) *PostgresRestaurantRepository {
	return &PostgresRestaurantRepository{db: db}
}

func (r *PostgresRestaurantRepository) Save(rest *restaurant.Restaurant) error {
	currency := rest.BaseCurrency
	if currency.IsZero() {
		currency = money.USD
	}
	_, err := r.db.Exec(
		`INSERT INTO restaurants (id, name, base_currency, tax_rate, table_count, is_open, closed_message, reopening_hours, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (id) DO UPDATE
		 SET name = EXCLUDED.name,
		     base_currency = EXCLUDED.base_currency,
		     tax_rate = EXCLUDED.tax_rate,
		     table_count = EXCLUDED.table_count,
		     is_open = EXCLUDED.is_open,
		     closed_message = EXCLUDED.closed_message,
		     reopening_hours = EXCLUDED.reopening_hours,
		     updated_at = EXCLUDED.updated_at`,
		string(rest.ID),
		rest.Name,
		currency.Code,
		rest.TaxRate,
		rest.TableCount,
		rest.IsOpen,
		rest.ClosedMessage,
		rest.ReopeningHours,
		rest.CreatedAt,
		rest.UpdatedAt,
	)
	return err
}

func (r *PostgresRestaurantRepository) FindByID(id common.RestaurantID) (*restaurant.Restaurant, error) {
	row := r.db.QueryRow(
		`SELECT id, name, COALESCE(base_currency, 'USD'), COALESCE(tax_rate, 0.08), table_count, is_open, closed_message, reopening_hours, created_at, updated_at
		 FROM restaurants WHERE id = $1`,
		string(id),
	)

	var (
		rid            string
		name           string
		baseCurrency   string
		taxRate        float64
		tableCount     int
		isOpen         bool
		closedMessage  sql.NullString
		reopeningHours sql.NullString
		createdAt      time.Time
		updatedAt      time.Time
	)

	if err := row.Scan(&rid, &name, &baseCurrency, &taxRate, &tableCount, &isOpen, &closedMessage, &reopeningHours, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("restaurant not found")
		}
		return nil, err
	}

	currency, err := money.Parse(baseCurrency)
	if err != nil {
		currency = money.USD
	}

	return &restaurant.Restaurant{
		ID:             common.RestaurantID(rid),
		Name:           name,
		BaseCurrency:   currency,
		TaxRate:        taxRate,
		TableCount:     tableCount,
		IsOpen:         isOpen,
		ClosedMessage:  closedMessage.String,
		ReopeningHours: reopeningHours.String,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (r *PostgresRestaurantRepository) Update(rest *restaurant.Restaurant) error {
	result, err := r.db.Exec(
		`UPDATE restaurants SET name=$2, tax_rate=$3, table_count=$4, is_open=$5, closed_message=$6, reopening_hours=$7, updated_at=$8 WHERE id=$1`,
		string(rest.ID),
		rest.Name,
		rest.TaxRate,
		rest.TableCount,
		rest.IsOpen,
		rest.ClosedMessage,
		rest.ReopeningHours,
		rest.UpdatedAt,
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
