package domain

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/payment/domain/payment"
)

type PaymentID = common.PaymentID
type PaymentMethodType = common.PaymentMethodType
type Payment = payment.Payment

var NewPayment = payment.NewPayment

const (
	PaymentMethodTypeCash      = common.PaymentMethodTypeCash
	PaymentMethodTypeLightning = common.PaymentMethodTypeLightning
)

// PaymentMethod defines the interface for payment processing (facade).
// New code should use payment.PaymentMethod directly.
type PaymentMethod = payment.PaymentMethod
