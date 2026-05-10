package event

import (
	"time"

	"bitmerchant/internal/common"
)

// OrderCreated is published when a new order is persisted.
type OrderCreated struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	TotalAmount  int64
	CreatedAt    time.Time
}

func (e OrderCreated) EventName() string     { return common.EventOrderCreated }
func (e OrderCreated) OccurredAt() time.Time { return e.CreatedAt }

// OrderPaid is published when payment is marked received.
type OrderPaid struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	TotalAmount  int64
	PaidAt       time.Time
}

func (e OrderPaid) EventName() string     { return common.EventOrderPaid }
func (e OrderPaid) OccurredAt() time.Time { return e.PaidAt }

// OrderPreparing is published when the kitchen starts preparing an order.
type OrderPreparing struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	PreparingAt  time.Time
}

func (e OrderPreparing) EventName() string     { return common.EventOrderPreparing }
func (e OrderPreparing) OccurredAt() time.Time { return e.PreparingAt }

// OrderReady is published when an order is ready for pickup / service.
type OrderReady struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	ReadyAt      time.Time
}

func (e OrderReady) EventName() string     { return common.EventOrderReady }
func (e OrderReady) OccurredAt() time.Time { return e.ReadyAt }

// OrderCompleted is reserved for a future completed lifecycle event.
type OrderCompleted struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	CompletedAt  time.Time
}

func (e OrderCompleted) EventName() string     { return common.EventOrderCompleted }
func (e OrderCompleted) OccurredAt() time.Time { return e.CompletedAt }

// OrderItemPrepToggled is published when a kitchen toggles a line item's prep_complete flag.
type OrderItemPrepToggled struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	ItemID       common.OrderItemID
	PrepComplete bool
	ToggledAt    time.Time
}

func (e OrderItemPrepToggled) EventName() string     { return common.EventOrderItemPrepToggled }
func (e OrderItemPrepToggled) OccurredAt() time.Time { return e.ToggledAt }
