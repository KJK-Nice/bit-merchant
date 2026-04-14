package events

import (
	"time"

	"bitmerchant/internal/common"
)

type OrderCreated struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	TotalAmount  int64
	CreatedAt    time.Time
}

func (e OrderCreated) EventName() string     { return common.EventOrderCreated }
func (e OrderCreated) OccurredAt() time.Time { return e.CreatedAt }

type OrderPaid struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	TotalAmount  int64
	PaidAt       time.Time
}

func (e OrderPaid) EventName() string     { return common.EventOrderPaid }
func (e OrderPaid) OccurredAt() time.Time { return e.PaidAt }

type OrderPreparing struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	PreparingAt  time.Time
}

func (e OrderPreparing) EventName() string     { return common.EventOrderPreparing }
func (e OrderPreparing) OccurredAt() time.Time { return e.PreparingAt }

type OrderReady struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	ReadyAt      time.Time
}

func (e OrderReady) EventName() string     { return common.EventOrderReady }
func (e OrderReady) OccurredAt() time.Time { return e.ReadyAt }

type OrderCompleted struct {
	OrderID      common.OrderID
	RestaurantID common.RestaurantID
	OrderNumber  common.OrderNumber
	CompletedAt  time.Time
}

func (e OrderCompleted) EventName() string     { return common.EventOrderCompleted }
func (e OrderCompleted) OccurredAt() time.Time { return e.CompletedAt }
