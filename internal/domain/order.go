package domain

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type OrderID = common.OrderID
type OrderNumber = common.OrderNumber
type OrderItemID = common.OrderItemID
type PaymentStatus = common.PaymentStatus
type FulfillmentStatus = common.FulfillmentStatus
type Order = order.Order
type OrderItem = order.OrderItem

var NewOrder = order.NewOrder
var NewOrderItem = order.NewOrderItem

const (
	PaymentStatusPending = common.PaymentStatusPending
	PaymentStatusPaid    = common.PaymentStatusPaid
	PaymentStatusFailed  = common.PaymentStatusFailed
	PaymentStatusExpired = common.PaymentStatusExpired
)

const (
	FulfillmentStatusPaid      = common.FulfillmentStatusPaid
	FulfillmentStatusPreparing = common.FulfillmentStatusPreparing
	FulfillmentStatusReady     = common.FulfillmentStatusReady
	FulfillmentStatusCompleted = common.FulfillmentStatusCompleted
)
