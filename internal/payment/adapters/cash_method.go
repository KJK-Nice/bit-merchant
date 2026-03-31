package adapters

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/payment/domain/payment"
)

type CashPaymentMethod struct{}

func NewCashPaymentMethod() *CashPaymentMethod {
	return &CashPaymentMethod{}
}

func (p *CashPaymentMethod) ProcessPayment(ctx context.Context, orderID common.OrderID, restaurantID common.RestaurantID, fiatAmount float64) (*payment.Payment, error) {
	paymentID := common.PaymentID(fmt.Sprintf("pay_%d", time.Now().UnixNano()))
	return payment.NewPayment(paymentID, orderID, restaurantID, common.PaymentMethodTypeCash, fiatAmount)
}

func (p *CashPaymentMethod) ValidatePayment(ctx context.Context, orderID common.OrderID) error {
	return nil
}

func (p *CashPaymentMethod) GetPaymentMethodType() common.PaymentMethodType {
	return common.PaymentMethodTypeCash
}
