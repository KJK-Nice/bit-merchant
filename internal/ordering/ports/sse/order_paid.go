package sse

import (
	"bytes"
	"context"
	"fmt"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
)

// OrderPaidHandler broadcasts paid state to kitchen and customer streams.
type OrderPaidHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewOrderPaidHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *OrderPaidHandler {
	return &OrderPaidHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderPaidHandler) Handle(ctx context.Context, ev event.OrderPaid) error {
	h.logger.Info("Order Paid", "orderID", ev.OrderID)

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

	// Remove the now-paid card from the FOH/server view.
	removalSelector := fmt.Sprintf("#server-order-%s", order.ID)
	removalMsg := commonhttp.FormatDatastarPatch("", removalSelector, "remove")
	h.sse.Broadcast(commonhttp.TopicServer, removalMsg)

	broadcastCustomerStatus(ctx, h.logger, h.sse, h.repo, order)
	return nil
}
