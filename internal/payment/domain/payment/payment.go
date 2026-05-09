package payment

import (
	"context"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

// Payment represents a payment transaction.
type Payment struct {
	ID            common.PaymentID
	OrderID       common.OrderID
	RestaurantID  common.RestaurantID
	Method        common.PaymentMethodType
	Amount        float64
	Currency      money.Currency
	Status        common.PaymentStatus
	CreatedAt     time.Time
	PaidAt        *time.Time
	FailedAt      *time.Time
	FailureReason string
}

// Money returns the payment amount as money.Money. Falls back to USD when
// the row predates currency support.
func (p *Payment) Money() money.Money {
	c := p.Currency
	if c.IsZero() {
		c = money.USD
	}
	return money.FromMajor(p.Amount, c)
}

// PaymentMethod defines the interface for payment processing. The amount is
// passed as money.Money so each adapter sees the unit it expects (sats for
// Lightning, fiat for Cash).
type PaymentMethod interface {
	ProcessPayment(ctx context.Context, orderID common.OrderID, restaurantID common.RestaurantID, amount money.Money) (*Payment, error)
	ValidatePayment(ctx context.Context, orderID common.OrderID) error
	GetPaymentMethodType() common.PaymentMethodType
}

// NewPayment creates a Payment in the legacy USD-only path.
func NewPayment(id common.PaymentID, orderID common.OrderID, restaurantID common.RestaurantID, method common.PaymentMethodType, amount float64) (*Payment, error) {
	return NewPaymentWithCurrency(id, orderID, restaurantID, method, amount, money.USD)
}

// NewPaymentWithCurrency creates a Payment in the given currency.
func NewPaymentWithCurrency(id common.PaymentID, orderID common.OrderID, restaurantID common.RestaurantID, method common.PaymentMethodType, amount float64, currency money.Currency) (*Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if currency.IsZero() {
		currency = money.USD
	}
	now := time.Now()
	return &Payment{
		ID: id, OrderID: orderID, RestaurantID: restaurantID,
		Method: method, Amount: amount, Currency: currency,
		Status:    common.PaymentStatusPending,
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
