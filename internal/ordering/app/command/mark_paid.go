package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
)

// MarkOrderPaid records payment for an order and publishes OrderPaid.
type MarkOrderPaid struct {
	OrderID common.OrderID
}

type MarkOrderPaidHandler decorator.CommandResultHandler[MarkOrderPaid, *order.Order]

type markOrderPaidHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderPaidHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) MarkOrderPaidHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := markOrderPaidHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[MarkOrderPaid, *order.Order](h, log, metrics)
}

func (h markOrderPaidHandler) Handle(ctx context.Context, cmd MarkOrderPaid) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	o.MarkPaid()

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	ev := event.OrderPaid{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		TotalAmount:  o.TotalAmount,
		PaidAt:       time.Now(),
	}
	if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
		return nil, err
	}

	return o, nil
}
