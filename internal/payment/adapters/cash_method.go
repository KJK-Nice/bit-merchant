package adapters

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/payment/domain/payment"
)

type CashPaymentMethod struct{}

func NewCashPaymentMethod() *CashPaymentMethod {
	return &CashPaymentMethod{}
}

func (p *CashPaymentMethod) ProcessPayment(ctx context.Context, orderID common.OrderID, restaurantID common.RestaurantID, amount money.Money) (*payment.Payment, error) {
	paymentID := common.PaymentID(fmt.Sprintf("pay_%d", time.Now().UnixNano()))
	return payment.NewPaymentWithCurrency(paymentID, orderID, restaurantID, common.PaymentMethodTypeCash, amount.Major(), amount.Currency)
}

func (p *CashPaymentMethod) ValidatePayment(ctx context.Context, orderID common.OrderID) error {
	return nil
}

func (p *CashPaymentMethod) GetPaymentMethodType() common.PaymentMethodType {
	return common.PaymentMethodTypeCash
}
