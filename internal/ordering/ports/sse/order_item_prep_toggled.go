package sse

import (
	"bytes"
	"context"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
)

// OrderItemPrepToggledHandler re-renders the order card so peer kitchen tabs
// see the same item-tick state.
type OrderItemPrepToggledHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewOrderItemPrepToggledHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *OrderItemPrepToggledHandler {
	return &OrderItemPrepToggledHandler{logger: logger, sse: sse, repo: repo}
}

func (h *OrderItemPrepToggledHandler) Handle(ctx context.Context, ev event.OrderItemPrepToggled) error {
	h.logger.Info("Order item prep toggled", "orderID", ev.OrderID, "itemID", ev.ItemID, "complete", ev.PrepComplete)

	o, err := h.repo.FindByID(ev.OrderID)
	if err != nil || o == nil {
		h.logger.Error("Order not found for broadcasting", "orderID", ev.OrderID)
		return err
	}

	var buf bytes.Buffer
	if err := components.OrderCard(o).Render(ctx, &buf); err == nil {
		msg := commonhttp.FormatDatastarEvent(buf.String())
		h.sse.Broadcast(commonhttp.TopicKitchen, msg)
	}
	return nil
}
