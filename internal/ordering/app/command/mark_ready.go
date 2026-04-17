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

// MarkOrderReady marks an order ready for pickup or service.
type MarkOrderReady struct {
	OrderID common.OrderID
}

type MarkOrderReadyHandler decorator.CommandResultHandler[MarkOrderReady, *order.Order]

type markOrderReadyHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderReadyHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) MarkOrderReadyHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := markOrderReadyHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[MarkOrderReady, *order.Order](h, log, metrics)
}

func (h markOrderReadyHandler) Handle(ctx context.Context, cmd MarkOrderReady) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if err := o.MarkReady(); err != nil {
		return nil, err
	}

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	ev := event.OrderReady{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		ReadyAt:      time.Now(),
	}
	if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
		return nil, err
	}

	return o, nil
}
