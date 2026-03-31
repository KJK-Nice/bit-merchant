package order

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
)

// Order represents a customer purchase record.
type Order struct {
	ID                common.OrderID
	OrderNumber       common.OrderNumber
	RestaurantID      common.RestaurantID
	SessionID         string
	Items             []OrderItem
	TotalAmount       int64
	FiatAmount        float64
	PaymentMethod     common.PaymentMethodType
	PaymentStatus     common.PaymentStatus
	FulfillmentStatus common.FulfillmentStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PaidAt            *time.Time
	PreparingAt       *time.Time
	ReadyAt           *time.Time
	CompletedAt       *time.Time
}

// NewOrder creates a new Order with validation.
func NewOrder(id common.OrderID, orderNumber common.OrderNumber, restaurantID common.RestaurantID, sessionID string, items []OrderItem, totalAmount int64, paymentMethod common.PaymentMethodType) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}
	if totalAmount <= 0 {
		return nil, errors.New("total amount must be greater than 0")
	}
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	now := time.Now()
	return &Order{
		ID:                id,
		OrderNumber:       orderNumber,
		RestaurantID:      restaurantID,
		SessionID:         sessionID,
		Items:             items,
		TotalAmount:       totalAmount,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     common.PaymentStatusPending,
		FulfillmentStatus: common.FulfillmentStatusPaid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// MarkPaid transitions payment status to paid.
func (o *Order) MarkPaid() {
	o.PaymentStatus = common.PaymentStatusPaid
	now := time.Now()
	o.PaidAt = &now
	o.UpdatedAt = now
}

// StartPreparing transitions fulfillment to preparing.
func (o *Order) StartPreparing() error {
	if o.PaymentStatus != common.PaymentStatusPaid {
		return errors.New("cannot prepare unpaid order")
	}
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusPreparing); err != nil {
		return err
	}
	now := time.Now()
	o.PreparingAt = &now
	return nil
}

// MarkReady transitions fulfillment to ready.
func (o *Order) MarkReady() error {
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusReady); err != nil {
		return err
	}
	now := time.Now()
	o.ReadyAt = &now
	return nil
}

// Complete transitions fulfillment to completed.
func (o *Order) Complete() error {
	if err := o.updateFulfillmentStatus(common.FulfillmentStatusCompleted); err != nil {
		return err
	}
	now := time.Now()
	o.CompletedAt = &now
	return nil
}

// UpdateFulfillmentStatus updates order fulfillment status with validation (kept for backward compat).
func (o *Order) UpdateFulfillmentStatus(newStatus common.FulfillmentStatus) error {
	return o.updateFulfillmentStatus(newStatus)
}

func (o *Order) updateFulfillmentStatus(newStatus common.FulfillmentStatus) error {
	if !isValidStatusTransition(o.FulfillmentStatus, newStatus) {
		return errors.New("invalid status transition")
	}
	o.FulfillmentStatus = newStatus
	o.UpdatedAt = time.Now()
	if newStatus == common.FulfillmentStatusCompleted {
		now := time.Now()
		o.CompletedAt = &now
	}
	return nil
}

func isValidStatusTransition(current, next common.FulfillmentStatus) bool {
	validTransitions := map[common.FulfillmentStatus][]common.FulfillmentStatus{
		common.FulfillmentStatusPaid:      {common.FulfillmentStatusPreparing},
		common.FulfillmentStatusPreparing: {common.FulfillmentStatusReady},
		common.FulfillmentStatusReady:     {common.FulfillmentStatusCompleted},
		common.FulfillmentStatusCompleted: {},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}
