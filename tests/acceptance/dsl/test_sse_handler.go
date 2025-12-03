package dsl

import (
	handler "bitmerchant/internal/interfaces/http"
)

// TestSSEHandler wraps SSEHandler to capture broadcasts for testing
type TestSSEHandler struct {
	*handler.SSEHandler
	capturedBroadcasts map[string][][]byte // topic -> messages
}

// NewTestSSEHandler creates a test SSE handler that captures broadcasts
func NewTestSSEHandler() *TestSSEHandler {
	return &TestSSEHandler{
		SSEHandler:         handler.NewSSEHandler(),
		capturedBroadcasts: make(map[string][][]byte),
	}
}

// Broadcast captures the broadcast and forwards it to the real handler
func (h *TestSSEHandler) Broadcast(topic string, message []byte) {
	// Capture the broadcast
	if h.capturedBroadcasts[topic] == nil {
		h.capturedBroadcasts[topic] = [][]byte{}
	}
	h.capturedBroadcasts[topic] = append(h.capturedBroadcasts[topic], message)

	// Forward to real handler
	h.SSEHandler.Broadcast(topic, message)
}

// GetCapturedBroadcasts returns all captured broadcasts for a topic
func (h *TestSSEHandler) GetCapturedBroadcasts(topic string) [][]byte {
	return h.capturedBroadcasts[topic]
}

// ClearCapturedBroadcasts clears captured broadcasts
func (h *TestSSEHandler) ClearCapturedBroadcasts() {
	h.capturedBroadcasts = make(map[string][][]byte)
}
