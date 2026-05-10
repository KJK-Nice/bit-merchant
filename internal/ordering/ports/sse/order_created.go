package sse

import (
	"bytes"
	"context"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
)

// OrderCreatedHandler pushes new orders to the kitchen SSE stream.
type OrderCreatedHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewOrderCreatedHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *OrderCreatedHandler {
	return &OrderCreatedHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, ev event.OrderCreated) error {
	h.logger.Info("New Order Created", "orderID", ev.OrderID, "amount", ev.TotalAmount)

	order, err := h.repo.FindByID(ev.OrderID)
	if err != nil {
		h.logger.Error("Failed to find order for broadcasting", "error", err)
		return err
	}
	if order == nil {
		h.logger.Error("Order not found for broadcasting", "orderID", ev.OrderID)
		return nil
	}

	var buf bytes.Buffer
	if err := components.OrderCard(order).Render(ctx, &buf); err != nil {
		h.logger.Error("Failed to render OrderCard", "error", err)
		return err
	}

	msg := commonhttp.FormatDatastarPatch(buf.String(), "#orders-list", "prepend")
	h.sse.Broadcast(commonhttp.TopicKitchen, msg)

	if order.PaymentStatus != common.PaymentStatusPaid {
		var serverBuf bytes.Buffer
		if err := components.ServerOrderCard(order).Render(ctx, &serverBuf); err == nil {
			serverMsg := commonhttp.FormatDatastarPatch(serverBuf.String(), "#server-orders", "prepend")
			h.sse.Broadcast(commonhttp.TopicServer, serverMsg)
		}
	}

	return nil
}
