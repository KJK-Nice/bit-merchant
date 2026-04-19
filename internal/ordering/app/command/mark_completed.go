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

// MarkOrderCompleted settles an order and removes it from the active kitchen queue.
type MarkOrderCompleted struct {
	OrderID common.OrderID
}

type MarkOrderCompletedHandler decorator.CommandResultHandler[MarkOrderCompleted, *order.Order]

type markOrderCompletedHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderCompletedHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) MarkOrderCompletedHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := markOrderCompletedHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[MarkOrderCompleted, *order.Order](h, log, metrics)
}

func (h markOrderCompletedHandler) Handle(ctx context.Context, cmd MarkOrderCompleted) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if err := o.Complete(); err != nil {
		return nil, err
	}

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	ev := event.OrderCompleted{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		CompletedAt:  time.Now(),
	}
	if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
		return nil, err
	}

	return o, nil
}
