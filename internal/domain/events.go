package domain

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type DomainEvent = common.DomainEvent
type EventBus = common.EventBus

const (
	EventOrderCreated   = common.EventOrderCreated
	EventOrderPaid      = common.EventOrderPaid
	EventOrderPreparing = common.EventOrderPreparing
	EventOrderReady     = common.EventOrderReady
	EventOrderCompleted = common.EventOrderCompleted
)

type OrderCreated = order.OrderCreated
type OrderPaid = order.OrderPaid
type OrderPreparing = order.OrderPreparing
type OrderReady = order.OrderReady
type OrderCompleted = order.OrderCompleted
