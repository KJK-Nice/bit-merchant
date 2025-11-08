package domain

import (
	"errors"
	"time"
)

// PaymentID represents a unique payment identifier
type PaymentID string

// SettlementStatus represents settlement status
type SettlementStatus string

const (
	SettlementStatusPending SettlementStatus = "pending"
	SettlementStatusSettled SettlementStatus = "settled"
)

// Payment represents a Lightning Network transaction record
type Payment struct {
	ID                      PaymentID
	RestaurantID            RestaurantID
	OrderID                 *OrderID
	InvoiceID               string
	Invoice                 string
	AmountSatoshis          int64
	AmountFiat              float64
	ExchangeRate            float64
	Status                  PaymentStatus
	SettlementStatus        SettlementStatus
	SettledAt               *time.Time
	SettlementTransactionID string
	CreatedAt               time.Time
	PaidAt                  *time.Time
	FailedAt                *time.Time
	FailureReason           string
}

// NewPayment creates a new Payment with validation
func NewPayment(id PaymentID, restaurantID RestaurantID, orderID *OrderID, invoiceID, invoice string, amountSatoshis int64, amountFiat, exchangeRate float64) (*Payment, error) {
	if invoiceID == "" {
		return nil, errors.New("invoice ID is required")
	}
	if invoice == "" {
		return nil, errors.New("invoice is required")
	}
	if amountSatoshis <= 0 {
		return nil, errors.New("amount in satoshis must be greater than 0")
	}
	if amountFiat <= 0 {
		return nil, errors.New("amount in fiat must be greater than 0")
	}
	if exchangeRate <= 0 {
		return nil, errors.New("exchange rate must be greater than 0")
	}

	now := time.Now()
	return &Payment{
		ID:               id,
		RestaurantID:     restaurantID,
		OrderID:          orderID,
		InvoiceID:        invoiceID,
		Invoice:          invoice,
		AmountSatoshis:   amountSatoshis,
		AmountFiat:       amountFiat,
		ExchangeRate:     exchangeRate,
		Status:           PaymentStatusPending,
		SettlementStatus: SettlementStatusPending,
		CreatedAt:        now,
	}, nil
}

// MarkAsPaid marks payment as paid
func (p *Payment) MarkAsPaid() {
	p.Status = PaymentStatusPaid
	now := time.Now()
	p.PaidAt = &now
}

// MarkPaid marks payment as paid with order ID
func (p *Payment) MarkPaid(orderID OrderID) error {
	if p.Status != PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}

	p.Status = PaymentStatusPaid
	p.OrderID = &orderID
	now := time.Now()
	p.PaidAt = &now
	return nil
}

// MarkFailed marks payment as failed
func (p *Payment) MarkFailed(reason string) error {
	if p.Status != PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}

	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.FailedAt = &now
	return nil
}

// MarkAsExpired marks payment as expired
func (p *Payment) MarkAsExpired(reason string) {
	p.Status = PaymentStatusExpired
	p.FailureReason = reason
	now := time.Now()
	p.FailedAt = &now
}

// MarkExpired marks payment as expired
func (p *Payment) MarkExpired() error {
	if p.Status != PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}

	p.Status = PaymentStatusExpired
	return nil
}

// MarkSettled marks payment as settled
func (p *Payment) MarkSettled(transactionID string) error {
	if p.Status != PaymentStatusPaid {
		return errors.New("payment must be paid before settlement")
	}
	if p.SettlementStatus != SettlementStatusPending {
		return errors.New("payment is already settled")
	}

	p.SettlementStatus = SettlementStatusSettled
	p.SettlementTransactionID = transactionID
	now := time.Now()
	p.SettledAt = &now
	return nil
}
