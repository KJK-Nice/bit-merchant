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

type OrderPreparingHandler struct {
	logger *logging.Logger
	sse    *handler.SSEHandler
	repo   domain.OrderRepository
}

func NewOrderPreparingHandler(logger *logging.Logger, sse *handler.SSEHandler, repo domain.OrderRepository) *OrderPreparingHandler {
	return &OrderPreparingHandler{
		logger: logger,
		sse:    sse,
		repo:   repo,
	}
}

func (h *OrderPreparingHandler) Handle(ctx context.Context, event domain.OrderPreparing) error {
	h.logger.Info("Order Preparing", "orderID", event.OrderID)

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

