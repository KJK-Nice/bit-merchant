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

// MarkOrderPreparing moves a paid order into preparing state.
type MarkOrderPreparing struct {
	OrderID common.OrderID
}

type MarkOrderPreparingHandler decorator.CommandResultHandler[MarkOrderPreparing, *order.Order]

type markOrderPreparingHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderPreparingHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) MarkOrderPreparingHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := markOrderPreparingHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[MarkOrderPreparing, *order.Order](h, log, metrics)
}

func (h markOrderPreparingHandler) Handle(ctx context.Context, cmd MarkOrderPreparing) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if err := o.StartPreparing(); err != nil {
		return nil, err
	}

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	ev := event.OrderPreparing{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		PreparingAt:  time.Now(),
	}
	if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
		return nil, err
	}

	return o, nil
}
