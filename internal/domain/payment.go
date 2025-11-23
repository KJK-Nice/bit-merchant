package domain

import (
	"context"
	"errors"
	"time"
)

// PaymentID represents a unique payment identifier
type PaymentID string

// PaymentMethodType represents payment method type
type PaymentMethodType string

const (
	PaymentMethodTypeCash      PaymentMethodType = "cash"
	PaymentMethodTypeLightning PaymentMethodType = "lightning"
)

// Payment represents a payment transaction
type Payment struct {
	ID            PaymentID
	OrderID       OrderID
	RestaurantID  RestaurantID
	Method        PaymentMethodType
	Amount        float64
	Status        PaymentStatus
	CreatedAt     time.Time
	PaidAt        *time.Time
	FailedAt      *time.Time
	FailureReason string
}

// PaymentMethod defines the interface for payment processing
type PaymentMethod interface {
	ProcessPayment(ctx context.Context, order *Order) (*Payment, error)
	ValidatePayment(ctx context.Context, order *Order) error
	GetPaymentMethodType() PaymentMethodType
}

// NewPayment creates a new Payment with validation
func NewPayment(id PaymentID, orderID OrderID, restaurantID RestaurantID, method PaymentMethodType, amount float64) (*Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	now := time.Now()
	return &Payment{
		ID:           id,
		OrderID:      orderID,
		RestaurantID: restaurantID,
		Method:       method,
		Amount:       amount,
		Status:       PaymentStatusPending,
		CreatedAt:    now,
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
	p.OrderID = orderID
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

// MarkExpired marks payment as expired
func (p *Payment) MarkExpired() error {
	if p.Status != PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}

	p.Status = PaymentStatusExpired
	return nil
}
