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

// OrderReadyHandler broadcasts ready state to kitchen and customer streams.
type OrderReadyHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewOrderReadyHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *OrderReadyHandler {
	return &OrderReadyHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderReadyHandler) Handle(ctx context.Context, ev event.OrderReady) error {
	h.logger.Info("Order Ready", "orderID", ev.OrderID)

	order, err := h.repo.FindByID(ev.OrderID)
	if err != nil || order == nil {
		h.logger.Error("Order not found for broadcasting", "orderID", ev.OrderID)
		return err
	}

	var bufCard bytes.Buffer
	if err := components.OrderCard(order).Render(ctx, &bufCard); err == nil {
		msg := commonhttp.FormatDatastarEvent(bufCard.String())
		h.sse.Broadcast(commonhttp.TopicKitchen, msg)
	}

	broadcastCustomerStatus(ctx, h.logger, h.sse, h.repo, order)
	return nil
}
