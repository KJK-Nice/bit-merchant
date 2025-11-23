package handlers

import (
	"bytes"
	"context"
	"fmt"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/logging"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/interfaces/templates/components"
)

type OrderReadyHandler struct {
	logger *logging.Logger
	sse    *handler.SSEHandler
	repo   domain.OrderRepository
}

func NewOrderReadyHandler(logger *logging.Logger, sse *handler.SSEHandler, repo domain.OrderRepository) *OrderReadyHandler {
	return &OrderReadyHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderReadyHandler) Handle(ctx context.Context, event domain.OrderReady) error {
	h.logger.Info("Order Ready", "orderID", event.OrderID)

	order, err := h.repo.FindByID(event.OrderID)
	if err != nil || order == nil {
		h.logger.Error("Order not found for broadcasting", "orderID", event.OrderID)
		return err
	}

	// 1. Broadcast to Kitchen (OrderCard)
	var bufCard bytes.Buffer
	if err := components.OrderCard(order).Render(ctx, &bufCard); err == nil {
		msg := handler.FormatDatastarEvent(bufCard.String())
		h.sse.Broadcast(handler.TopicKitchen, msg)
	}

	// 2. Broadcast to Customer (OrderStatus)
	var bufStatus bytes.Buffer
	if err := templates.OrderStatus(order).Render(ctx, &bufStatus); err == nil {
		msg := handler.FormatDatastarEvent(bufStatus.String())
		h.sse.Broadcast(fmt.Sprintf(handler.TopicOrder, order.OrderNumber), msg)
	}

	return nil
}

