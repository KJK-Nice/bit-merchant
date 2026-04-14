package handlers

import (
	"bitmerchant/internal/infrastructure/logging"
	ifaceevents "bitmerchant/internal/interfaces/events"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/domain/order"
	"bytes"
	"context"
)

type OrderCreatedHandler struct {
	logger *logging.Logger
	sse    *handler.SSEHandler
	repo   order.Repository
}

func NewOrderCreatedHandler(logger *logging.Logger, sse *handler.SSEHandler, repo order.Repository) *OrderCreatedHandler {
	return &OrderCreatedHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, event ifaceevents.OrderCreated) error {
	h.logger.Info("New Order Created", "orderID", event.OrderID, "amount", event.TotalAmount)

	// Fetch full order to render card
	order, err := h.repo.FindByID(event.OrderID)
	if err != nil {
		h.logger.Error("Failed to find order for broadcasting", "error", err)
		return err
	}
	if order == nil {
		h.logger.Error("Order not found for broadcasting", "orderID", event.OrderID)
		return nil
	}

	// Render OrderCard
	var buf bytes.Buffer
	if err := components.OrderCard(order).Render(ctx, &buf); err != nil {
		h.logger.Error("Failed to render OrderCard", "error", err)
		return err
	}

	// Broadcast to Kitchen
	msg := handler.FormatDatastarPatch(buf.String(), "#orders-list", "prepend")
	h.sse.Broadcast(handler.TopicKitchen, msg)

	return nil
}
