package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	// Datastar event names
	EventDatastarPatchElements = "datastar-patch-elements"

	// Topic names for internal broadcasting
	TopicKitchen = "kitchen"
	TopicOrder   = "order:%s"
)

// SSEHandler handles Server-Sent Events
type SSEHandler struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]bool
}

// NewSSEHandler creates a new SSEHandler
func NewSSEHandler() *SSEHandler {
	return &SSEHandler{
		clients: make(map[string]map[chan []byte]bool),
	}
}

// OrderStatusStream handles GET /order/:orderNumber/stream
func (h *SSEHandler) OrderStatusStream(c echo.Context) error {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		return c.String(http.StatusBadRequest, "Order number required")
	}

	return h.handleStream(c, fmt.Sprintf(TopicOrder, orderNumber))
}

// KitchenStream handles GET /kitchen/stream
func (h *SSEHandler) KitchenStream(c echo.Context) error {
	return h.handleStream(c, TopicKitchen)
}

func (h *SSEHandler) handleStream(c echo.Context, topic string) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")

	clientChan := make(chan []byte)
	h.addClient(topic, clientChan)
	defer h.removeClient(topic, clientChan)

	// Send connection established message (optional, helpful for debugging)
	// c.Response().Write([]byte(": connected\n\n"))
	// c.Response().Flush()

	ticker := time.NewTicker(15 * time.Second) // Keep-alive
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case msg := <-clientChan:
			// Message is expected to be fully formatted SSE event
			if _, err := c.Response().Write(msg); err != nil {
				return err
			}
			c.Response().Flush()
		case <-ticker.C:
			if _, err := c.Response().Write([]byte(": keepalive\n\n")); err != nil {
				return err
			}
			c.Response().Flush()
		}
	}
}

func (h *SSEHandler) addClient(id string, client chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[id]; !ok {
		h.clients[id] = make(map[chan []byte]bool)
	}
	h.clients[id][client] = true
}

func (h *SSEHandler) removeClient(id string, client chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[id]; ok {
		delete(clients, client)
		close(client)
		if len(clients) == 0 {
			delete(h.clients, id)
		}
	}
}

// Broadcast sends message to all clients listening to id
func (h *SSEHandler) Broadcast(id string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[id]; ok {
		for client := range clients {
			select {
			case client <- message:
			default:
				// Drop message if client is blocked
			}
		}
	}
}

// FormatDatastarEvent formats the message as a Datastar SSE event
func FormatDatastarEvent(fragment string) []byte {
	return []byte(fmt.Sprintf("event: %s\ndata: elements %s\n\n", EventDatastarPatchElements, fragment))
}

// FormatDatastarPatch formats the message as a Datastar patch event with selector and mode
func FormatDatastarPatch(fragment, selector, mode string) []byte {
	return []byte(fmt.Sprintf("event: %s\ndata: selector %s\ndata: mode %s\ndata: elements %s\n\n",
		EventDatastarPatchElements, selector, mode, fragment))
}
