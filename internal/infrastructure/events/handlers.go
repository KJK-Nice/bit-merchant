package events

import (
	"context"
	"encoding/json"

	"bitmerchant/internal/domain"
)

// OrderPaidHandler handles OrderPaid events
type OrderPaidHandler struct {
	createOrderUseCase interface {
		Execute(req interface{}) (interface{}, error)
	}
}

// NewOrderPaidHandler creates a new OrderPaidHandler
func NewOrderPaidHandler(createOrderUseCase interface {
	Execute(req interface{}) (interface{}, error)
}) *OrderPaidHandler {
	return &OrderPaidHandler{
		createOrderUseCase: createOrderUseCase,
	}
}

// Handle processes OrderPaid event
func (h *OrderPaidHandler) Handle(ctx context.Context, event domain.OrderPaid) error {
	// This will be called by Watermill when OrderPaid event is published
	// For now, it's a placeholder - actual implementation will be in handlers.go
	return nil
}

// OrderStatusChangedHandler handles OrderStatusChanged events
type OrderStatusChangedHandler struct {
	sseHub interface {
		Broadcast(orderNumber domain.OrderNumber, message []byte)
	}
}

// NewOrderStatusChangedHandler creates a new OrderStatusChangedHandler
func NewOrderStatusChangedHandler(sseHub interface {
	Broadcast(orderNumber domain.OrderNumber, message []byte)
}) *OrderStatusChangedHandler {
	return &OrderStatusChangedHandler{
		sseHub: sseHub,
	}
}

// Handle processes OrderStatusChanged event
func (h *OrderStatusChangedHandler) Handle(ctx context.Context, event domain.OrderStatusChanged) error {
	// Broadcast status change via SSE
	message, err := json.Marshal(map[string]interface{}{
		"type":           "order-status-changed",
		"orderNumber":    string(event.OrderNumber),
		"previousStatus": string(event.PreviousStatus),
		"newStatus":      string(event.NewStatus),
		"updatedAt":      event.UpdatedAt,
	})
	if err != nil {
		return err
	}

	h.sseHub.Broadcast(event.OrderNumber, message)
	return nil
}
