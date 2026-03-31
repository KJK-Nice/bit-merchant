package adapters

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/payment/domain/payment"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(p *payment.Payment) error {
	_, err := r.db.Exec(
		`INSERT INTO payments (id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (id) DO UPDATE SET
		   order_id=EXCLUDED.order_id, status=EXCLUDED.status, paid_at=EXCLUDED.paid_at,
		   failed_at=EXCLUDED.failed_at, failure_reason=EXCLUDED.failure_reason`,
		string(p.ID), string(p.OrderID), string(p.RestaurantID),
		string(p.Method), p.Amount, string(p.Status),
		p.CreatedAt, p.PaidAt, p.FailedAt, p.FailureReason)
	return err
}

func (r *PostgresPaymentRepository) FindByID(id common.PaymentID) (*payment.Payment, error) {
	row := r.db.QueryRow(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		 FROM payments WHERE id = $1`, string(id))
	return scanPayment(row)
}

func (r *PostgresPaymentRepository) FindByOrderID(orderID common.OrderID) (*payment.Payment, error) {
	row := r.db.QueryRow(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		 FROM payments WHERE order_id = $1 LIMIT 1`, string(orderID))
	return scanPayment(row)
}

func (r *PostgresPaymentRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*payment.Payment, error) {
	rows, err := r.db.Query(
		`SELECT id, order_id, restaurant_id, method, amount, status, created_at, paid_at, failed_at, failure_reason
		 FROM payments WHERE restaurant_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*payment.Payment
	for rows.Next() {
		p, err := scanPaymentRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *PostgresPaymentRepository) Update(p *payment.Payment) error {
	result, err := r.db.Exec(
		`UPDATE payments SET order_id=$2, status=$3, paid_at=$4, failed_at=$5, failure_reason=$6 WHERE id=$1`,
		string(p.ID), string(p.OrderID), string(p.Status), p.PaidAt, p.FailedAt, p.FailureReason)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func scanPayment(row *sql.Row) (*payment.Payment, error) {
	var (
		id, orderID, restID, method, status string
		amount                              float64
		createdAt                           time.Time
		paidAt, failedAt                    sql.NullTime
		failureReason                       sql.NullString
	)
	if err := row.Scan(&id, &orderID, &restID, &method, &amount, &status, &createdAt, &paidAt, &failedAt, &failureReason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return buildPayment(id, orderID, restID, method, amount, status, createdAt, paidAt, failedAt, failureReason), nil
}

func scanPaymentRows(rows *sql.Rows) (*payment.Payment, error) {
	var (
		id, orderID, restID, method, status string
		amount                              float64
		createdAt                           time.Time
		paidAt, failedAt                    sql.NullTime
		failureReason                       sql.NullString
	)
	if err := rows.Scan(&id, &orderID, &restID, &method, &amount, &status, &createdAt, &paidAt, &failedAt, &failureReason); err != nil {
		return nil, err
	}
	return buildPayment(id, orderID, restID, method, amount, status, createdAt, paidAt, failedAt, failureReason), nil
}

func buildPayment(id, orderID, restID, method string, amount float64, status string, createdAt time.Time, paidAt, failedAt sql.NullTime, failureReason sql.NullString) *payment.Payment {
	p := &payment.Payment{
		ID: common.PaymentID(id), OrderID: common.OrderID(orderID),
		RestaurantID: common.RestaurantID(restID),
		Method: common.PaymentMethodType(method), Amount: amount,
		Status: common.PaymentStatus(status), CreatedAt: createdAt,
		FailureReason: failureReason.String,
	}
	if paidAt.Valid {
		t := paidAt.Time
		p.PaidAt = &t
	}
	if failedAt.Valid {
		t := failedAt.Time
		p.FailedAt = &t
	}
	return p
}
