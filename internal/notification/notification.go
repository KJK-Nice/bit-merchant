package notification

import "context"

// Notification carries the content to be delivered across one or more channels.
type Notification struct {
	Title    string
	Body     string
	URL      string            // destination when the user taps the notification
	Metadata map[string]string // channel-specific hints (e.g. "role", "order_number", "restaurant_id")
}

// Notifier delivers a Notification via a single channel.
// Implement this interface to add Email, SMS, Line, etc.
type Notifier interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}
