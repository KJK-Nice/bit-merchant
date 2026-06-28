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

// RequestBill records a customer "request bill" action. Repeated taps within
// order.ServiceRequestThrottle are no-ops (no event published).
type RequestBill struct {
	OrderID common.OrderID
}

type RequestBillHandler decorator.CommandResultHandler[RequestBill, *order.Order]

type requestBillHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewRequestBillHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) RequestBillHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := requestBillHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[RequestBill, *order.Order](h, log, metrics)
}

func (h requestBillHandler) Handle(ctx context.Context, cmd RequestBill) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if !o.RequestBill(time.Now()) {
		// Throttled: already requested within the window. Idempotent success.
		return o, nil
	}

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	if h.eventBus != nil {
		ev := event.BillRequested{
			OrderID:      o.ID,
			RestaurantID: o.RestaurantID,
			OrderNumber:  o.OrderNumber,
			TableLabel:   o.TableLabel,
			CustomerName: o.CustomerName,
			RequestedAt:  *o.BillRequestedAt,
		}
		if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
			return nil, err
		}
	}

	return o, nil
}
