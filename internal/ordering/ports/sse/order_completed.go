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

// OrderCompletedHandler broadcasts completed state to kitchen and customer streams.
type OrderCompletedHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewOrderCompletedHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *OrderCompletedHandler {
	return &OrderCompletedHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderCompletedHandler) Handle(ctx context.Context, ev event.OrderCompleted) error {
	h.logger.Info("Order Completed", "orderID", ev.OrderID)

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
