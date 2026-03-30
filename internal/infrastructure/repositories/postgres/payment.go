package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// PaymentRepository implements domain.PaymentRepository for PostgreSQL.
type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Save(payment *domain.Payment) error {
	_, err := r.db.Exec(
		`INSERT INTO payments (id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (id) DO UPDATE
		 SET order_id = EXCLUDED.order_id,
		     restaurant_id = EXCLUDED.restaurant_id,
		     method = EXCLUDED.method,
		     amount = EXCLUDED.amount,
		     status = EXCLUDED.status,
		     paid_at = EXCLUDED.paid_at,
		     failed_at = EXCLUDED.failed_at,
		     failure_reason = EXCLUDED.failure_reason`,
		string(payment.ID),
		string(payment.OrderID),
		string(payment.RestaurantID),
		string(payment.Method),
		payment.Amount,
		string(payment.Status),
		payment.CreatedAt,
		payment.PaidAt,
		payment.FailedAt,
		payment.FailureReason,
	)
	return err
}

func (r *PaymentRepository) FindByID(id domain.PaymentID) (*domain.Payment, error) {
	row := r.db.QueryRow(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		   FROM payments
		  WHERE id = $1`,
		string(id),
	)
	return scanPaymentRow(row)
}

func (r *PaymentRepository) FindByOrderID(orderID domain.OrderID) (*domain.Payment, error) {
	row := r.db.QueryRow(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		   FROM payments
		  WHERE order_id = $1
		  LIMIT 1`,
		string(orderID),
	)
	return scanPaymentRow(row)
}

func (r *PaymentRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Payment, error) {
	rows, err := r.db.Query(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		   FROM payments
		  WHERE restaurant_id = $1
		  ORDER BY created_at DESC`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment, scanErr := scanPayment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

func (r *PaymentRepository) Update(payment *domain.Payment) error {
	result, err := r.db.Exec(
		`UPDATE payments
		    SET order_id = $2,
		        restaurant_id = $3,
		        method = $4,
		        amount = $5,
		        status = $6,
		        paid_at = $7,
		        failed_at = $8,
		        failure_reason = $9
		  WHERE id = $1`,
		string(payment.ID),
		string(payment.OrderID),
		string(payment.RestaurantID),
		string(payment.Method),
		payment.Amount,
		string(payment.Status),
		payment.PaidAt,
		payment.FailedAt,
		payment.FailureReason,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func scanPayment(rows *sql.Rows) (*domain.Payment, error) {
	var (
		id            string
		orderID       string
		restaurantID  string
		method        string
		amount        float64
		status        string
		createdAt     sql.NullTime
		paidAt        sql.NullTime
		failedAt      sql.NullTime
		failureReason string
	)
	if err := rows.Scan(&id, &orderID, &restaurantID, &method, &amount, &status, &createdAt, &paidAt, &failedAt, &failureReason); err != nil {
		return nil, err
	}

	return &domain.Payment{
		ID:            domain.PaymentID(id),
		OrderID:       domain.OrderID(orderID),
		RestaurantID:  domain.RestaurantID(restaurantID),
		Method:        domain.PaymentMethodType(method),
		Amount:        amount,
		Status:        domain.PaymentStatus(status),
		CreatedAt:     createdAt.Time,
		PaidAt:        nullTimePtr(paidAt),
		FailedAt:      nullTimePtr(failedAt),
		FailureReason: failureReason,
	}, nil
}

func scanPaymentRow(row *sql.Row) (*domain.Payment, error) {
	var (
		id            string
		orderID       string
		restaurantID  string
		method        string
		amount        float64
		status        string
		createdAt     sql.NullTime
		paidAt        sql.NullTime
		failedAt      sql.NullTime
		failureReason string
	)
	if err := row.Scan(&id, &orderID, &restaurantID, &method, &amount, &status, &createdAt, &paidAt, &failedAt, &failureReason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	return &domain.Payment{
		ID:            domain.PaymentID(id),
		OrderID:       domain.OrderID(orderID),
		RestaurantID:  domain.RestaurantID(restaurantID),
		Method:        domain.PaymentMethodType(method),
		Amount:        amount,
		Status:        domain.PaymentStatus(status),
		CreatedAt:     createdAt.Time,
		PaidAt:        nullTimePtr(paidAt),
		FailedAt:      nullTimePtr(failedAt),
		FailureReason: failureReason,
	}, nil
}
