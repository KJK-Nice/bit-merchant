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

// ToggleOrderItemPrep flips the prep_complete flag on a single line item.
type ToggleOrderItemPrep struct {
	OrderID  common.OrderID
	ItemID   common.OrderItemID
	Complete bool
}

type ToggleOrderItemPrepHandler decorator.CommandResultHandler[ToggleOrderItemPrep, *order.Order]

type toggleOrderItemPrepHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewToggleOrderItemPrepHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) ToggleOrderItemPrepHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := toggleOrderItemPrepHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[ToggleOrderItemPrep, *order.Order](h, log, metrics)
}

func (h toggleOrderItemPrepHandler) Handle(ctx context.Context, cmd ToggleOrderItemPrep) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}
	if !o.SetItemPrepComplete(cmd.ItemID, cmd.Complete) {
		return nil, errors.New("order item not found")
	}
	if err := h.repo.UpdateItemPrepComplete(o.ID, cmd.ItemID, cmd.Complete); err != nil {
		return nil, err
	}

	ev := event.OrderItemPrepToggled{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		ItemID:       cmd.ItemID,
		PrepComplete: cmd.Complete,
		ToggledAt:    time.Now(),
	}
	if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
		return nil, err
	}

	return o, nil
}
