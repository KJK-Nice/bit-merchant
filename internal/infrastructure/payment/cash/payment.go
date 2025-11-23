package cash

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/domain"
)

type CashPaymentMethod struct{}

func NewCashPaymentMethod() *CashPaymentMethod {
	return &CashPaymentMethod{}
}

func (p *CashPaymentMethod) ProcessPayment(ctx context.Context, order *domain.Order) (*domain.Payment, error) {
	// For cash, we just create a pending payment record.
	// Actual payment is confirmed manually by staff.
	paymentID := domain.PaymentID(fmt.Sprintf("pay_%d", time.Now().UnixNano()))

	return domain.NewPayment(
		paymentID,
		order.ID,
		order.RestaurantID,
		domain.PaymentMethodTypeCash,
		order.FiatAmount,
	)
}

func (p *CashPaymentMethod) ValidatePayment(ctx context.Context, order *domain.Order) error {
	// Cash is always valid to initiate
	return nil
}

func (p *CashPaymentMethod) GetPaymentMethodType() domain.PaymentMethodType {
	return domain.PaymentMethodTypeCash
}

