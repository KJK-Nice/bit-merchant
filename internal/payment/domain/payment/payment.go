package payment

import (
	"context"
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// Payment represents a payment transaction.
type Payment struct {
	ID            common.PaymentID
	OrderID       common.OrderID
	RestaurantID  common.RestaurantID
	Method        common.PaymentMethodType
	Amount        float64
	Status        common.PaymentStatus
	CreatedAt     time.Time
	PaidAt        *time.Time
	FailedAt      *time.Time
	FailureReason string
}

// PaymentMethod defines the interface for payment processing.
type PaymentMethod interface {
	ProcessPayment(ctx context.Context, orderID common.OrderID, restaurantID common.RestaurantID, fiatAmount float64) (*Payment, error)
	ValidatePayment(ctx context.Context, orderID common.OrderID) error
	GetPaymentMethodType() common.PaymentMethodType
}

func NewPayment(id common.PaymentID, orderID common.OrderID, restaurantID common.RestaurantID, method common.PaymentMethodType, amount float64) (*Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	now := time.Now()
	return &Payment{
		ID: id, OrderID: orderID, RestaurantID: restaurantID,
		Method: method, Amount: amount, Status: common.PaymentStatusPending,
		CreatedAt: now,
	}, nil
}

func (p *Payment) MarkAsPaid() {
	p.Status = common.PaymentStatusPaid
	now := time.Now()
	p.PaidAt = &now
}

func (p *Payment) MarkPaid(orderID common.OrderID) error {
	if p.Status != common.PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}
	p.Status = common.PaymentStatusPaid
	p.OrderID = orderID
	now := time.Now()
	p.PaidAt = &now
	return nil
}

func (p *Payment) MarkFailed(reason string) error {
	if p.Status != common.PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}
	p.Status = common.PaymentStatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.FailedAt = &now
	return nil
}

func (p *Payment) MarkExpired() error {
	if p.Status != common.PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}
	p.Status = common.PaymentStatusExpired
	return nil
}
