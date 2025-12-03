package dsl

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// SSEClient provides a client for testing Server-Sent Events
type SSEClient struct {
	t      *testing.T
	app    *TestApplication
	events []SSEEvent
	ctx    context.Context
	cancel context.CancelFunc
	path   string
	topic  string
}

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Event    string
	Data     string
	ID       string
	Retry    string
	Selector string // Extracted from Datastar patch format
	Mode     string // Extracted from Datastar patch format
}

// NewSSEClient creates a new SSE client for testing
// The client will capture events broadcast to the SSE handler
func NewSSEClient(t *testing.T, app *TestApplication, path string) *SSEClient {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Determine topic from path
	topic := ""
	if path == "/kitchen/stream" {
		topic = "kitchen"
	} else if strings.HasPrefix(path, "/order/") && strings.HasSuffix(path, "/stream") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			orderNumber := parts[2]
			topic = fmt.Sprintf("order:%s", orderNumber)
		}
	}

	return &SSEClient{
		t:      t,
		app:    app,
		events: []SSEEvent{},
		ctx:    ctx,
		cancel: cancel,
		path:   path,
		topic:  topic,
	}
}

// ParseSSEMessage parses a raw SSE message into an SSEEvent
func ParseSSEMessage(message []byte) *SSEEvent {
	event := &SSEEvent{}
	lines := strings.Split(string(message), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			// Parse Datastar-specific format
			if strings.HasPrefix(data, "selector ") {
				event.Selector = strings.TrimSpace(strings.TrimPrefix(data, "selector "))
			} else if strings.HasPrefix(data, "mode ") {
				event.Mode = strings.TrimSpace(strings.TrimPrefix(data, "mode "))
			} else if strings.HasPrefix(data, "elements ") {
				event.Data = strings.TrimSpace(strings.TrimPrefix(data, "elements "))
			} else {
				event.Data = data
			}
		} else if strings.HasPrefix(line, "id:") {
			event.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "retry:") {
			event.Retry = strings.TrimSpace(strings.TrimPrefix(line, "retry:"))
		}
	}

	return event
}

// CaptureEvent captures an event that was broadcast
func (c *SSEClient) CaptureEvent(message []byte) {
	event := ParseSSEMessage(message)
	if event != nil {
		c.events = append(c.events, *event)
	}
}

// WaitForEvent waits for a specific event type to be captured
func (c *SSEClient) WaitForEvent(eventType string, timeout time.Duration) (*SSEEvent, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for i := range c.events {
			if c.events[i].Event == eventType {
				return &c.events[i], nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for event type: %s", eventType)
}

// WaitForDatastarPatch waits for a Datastar patch event
func (c *SSEClient) WaitForDatastarPatch(selector string, timeout time.Duration) (*SSEEvent, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for i := range c.events {
			if c.events[i].Event == "datastar-patch-elements" {
				if selector == "" || c.events[i].Selector == selector {
					return &c.events[i], nil
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for Datastar patch with selector: %s", selector)
}

// GetAllEvents returns all events received so far
func (c *SSEClient) GetAllEvents() []SSEEvent {
	return c.events
}

// GetEventsByType returns all events of a specific type
func (c *SSEClient) GetEventsByType(eventType string) []SSEEvent {
	var result []SSEEvent
	for _, event := range c.events {
		if event.Event == eventType {
			result = append(result, event)
		}
	}
	return result
}

// Close closes the SSE client
func (c *SSEClient) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// SSEStep represents connecting to an SSE stream
type SSEStep struct {
	Path string
}

func (s *SSEStep) Execute(t *testing.T, app *TestApplication) {
	client := NewSSEClient(t, app, s.Path)
	// Store client in test context for later assertions
	if app.context != nil {
		app.context.SetSSEClient(s.Path, client)
	}

	// Capture any already-broadcast events for this topic
	if app.testSSEHandler != nil && client.topic != "" {
		broadcasts := app.testSSEHandler.GetCapturedBroadcasts(client.topic)
		for _, msg := range broadcasts {
			client.CaptureEvent(msg)
		}
	}
}
