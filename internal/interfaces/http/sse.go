package http

import (
	"fmt"
	"net/http"
	"sync"

	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// SSEConnection represents a Server-Sent Events connection
type SSEConnection struct {
	orderNumber domain.OrderNumber
	ch          chan []byte
}

// SSEHub manages SSE connections for order status updates
type SSEHub struct {
	mu          sync.RWMutex
	connections map[domain.OrderNumber][]*SSEConnection
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		connections: make(map[domain.OrderNumber][]*SSEConnection),
	}
}

// Subscribe adds a new SSE connection
func (h *SSEHub) Subscribe(orderNumber domain.OrderNumber) *SSEConnection {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn := &SSEConnection{
		orderNumber: orderNumber,
		ch:          make(chan []byte, 10),
	}

	h.connections[orderNumber] = append(h.connections[orderNumber], conn)
	return conn
}

// Unsubscribe removes an SSE connection
func (h *SSEHub) Unsubscribe(orderNumber domain.OrderNumber, conn *SSEConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[orderNumber]
	for i, c := range conns {
		if c == conn {
			h.connections[orderNumber] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	close(conn.ch)
}

// Broadcast sends message to all connections for an order
func (h *SSEHub) Broadcast(orderNumber domain.OrderNumber, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.connections[orderNumber]
	for _, conn := range conns {
		select {
		case conn.ch <- message:
		default:
			// Channel full, skip
		}
	}
}

// OrderSSEHandler handles SSE for order status updates
type OrderSSEHandler struct {
	hub *SSEHub
}

// NewOrderSSEHandler creates a new OrderSSEHandler
func NewOrderSSEHandler(hub *SSEHub) *OrderSSEHandler {
	return &OrderSSEHandler{
		hub: hub,
	}
}

// StreamOrderStatus handles GET /order/:orderNumber/stream
func (h *OrderSSEHandler) StreamOrderStatus(c echo.Context) error {
	orderNumber := domain.OrderNumber(c.Param("orderNumber"))
	if orderNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "order number is required"})
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to updates
	conn := h.hub.Subscribe(orderNumber)
	defer h.hub.Unsubscribe(orderNumber, conn)

	// Send initial connection message
	fmt.Fprintf(c.Response(), "data: {\"type\":\"connected\"}\n\n")
	c.Response().Flush()

	// Stream updates
	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case message := <-conn.ch:
			fmt.Fprintf(c.Response(), "data: %s\n\n", message)
			c.Response().Flush()
		}
	}
}
