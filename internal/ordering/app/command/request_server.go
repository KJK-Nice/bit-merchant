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

// RequestServer records a customer "call server" request. Repeated taps within
// order.ServiceRequestThrottle are no-ops (no event published).
type RequestServer struct {
	OrderID common.OrderID
}

type RequestServerHandler decorator.CommandResultHandler[RequestServer, *order.Order]

type requestServerHandler struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewRequestServerHandler(repo order.Repository, eventBus common.EventBus, log *slog.Logger, metrics decorator.MetricsClient) RequestServerHandler {
	if repo == nil {
		panic("nil order.Repository")
	}
	h := requestServerHandler{repo: repo, eventBus: eventBus}
	return decorator.ApplyCommandResultDecorators[RequestServer, *order.Order](h, log, metrics)
}

func (h requestServerHandler) Handle(ctx context.Context, cmd RequestServer) (*order.Order, error) {
	o, err := h.repo.FindByID(cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if !o.RequestServer(time.Now()) {
		// Throttled: already requested within the window. Idempotent success.
		return o, nil
	}

	if err := h.repo.Update(o); err != nil {
		return nil, err
	}

	if h.eventBus != nil {
		ev := event.ServerCalled{
			OrderID:      o.ID,
			RestaurantID: o.RestaurantID,
			OrderNumber:  o.OrderNumber,
			TableLabel:   o.TableLabel,
			CustomerName: o.CustomerName,
			CalledAt:     *o.ServerCalledAt,
		}
		if err := h.eventBus.Publish(ctx, ev.EventName(), ev); err != nil {
			return nil, err
		}
	}

	return o, nil
}
