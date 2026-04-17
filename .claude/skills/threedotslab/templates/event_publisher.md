# Event Publisher Adapter Scaffold Template

Generate a Watermill publisher adapter that implements a domain/app-layer interface. The adapter lives in `adapters/` and translates domain operations into published messages. The interface lives in `app/command/services.go`.

## Placeholders

- `{{Name}}` — PascalCase aggregate name (e.g., `Training`)
- `{{name}}` — camelCase (e.g., `training`)
- `{{name_snake}}` — snake_case (e.g., `training`)
- `{{name_lower}}` — all lowercase (e.g., `training`)
- `{{module}}` — Go module path from go.mod
- `{{event}}` — PascalCase first event name (e.g., `TrainingScheduled`)
- `{{topic}}` — Dot-notation topic (e.g., `training.scheduled`)

## File 1: `app/command/services.go`

If this file already exists, add the interface. Otherwise create it:

```go
package command

import "context"

// {{Name}}EventPublisher defines events that can be emitted for {{name_lower}} operations.
// Implemented by adapters (e.g., Watermill AMQP adapter).
type {{Name}}EventPublisher interface {
	{{event}}(ctx context.Context) error
	// TODO: Add more event methods as needed
	// Example:
	// {{Name}}Cancelled(ctx context.Context, uuid string) error
}
```

## File 2: `adapters/{{name_snake}}_event_publisher.go`

```go
package adapters

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

type Watermill{{Name}}EventPublisher struct {
	pub message.Publisher
}

func NewWatermill{{Name}}EventPublisher(pub message.Publisher) Watermill{{Name}}EventPublisher {
	return Watermill{{Name}}EventPublisher{pub: pub}
}

// {{event}}Event is the wire format for the {{topic}} topic.
type {{event}}Event struct {
	// TODO: Add event payload fields
	// Example:
	// UUID string    `json:"uuid"`
	// Hour time.Time `json:"hour"`
}

func (p Watermill{{Name}}EventPublisher) {{event}}(ctx context.Context) error {
	event := {{event}}Event{
		// TODO: Map domain data to event fields
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(watermill.NewUUID(), msg)

	return p.pub.Publish("{{topic}}", msg)
}
```

## Update `service/application.go`

Wire the publisher adapter in the composition root:

```go
func NewApplication(ctx context.Context) (app.Application, func()) {
    // ... existing clients ...

    publisher, closePub, err := client.NewWatermillPublisher()
    if err != nil { panic(err) }

    eventPublisher := adapters.NewWatermill{{Name}}EventPublisher(publisher)

    return newApplication(ctx, eventPublisher),
        func() {
            // ... existing cleanup ...
            _ = closePub()
        }
}
```

Update the private `newApplication` to accept the publisher interface:

```go
func newApplication(
    ctx context.Context,
    eventPublisher command.{{Name}}EventPublisher,
    // ... existing deps ...
) app.Application {
    // ... pass eventPublisher to command handlers that need it
}
```

## Update command handler

Inject the publisher into the command handler that triggers the event:

```go
type {{name}}Handler struct {
    {{name_lower}}Repo    {{name_lower}}.Repository
    eventPublisher command.{{Name}}EventPublisher
}

func (h {{name}}Handler) Handle(ctx context.Context, cmd {{command}}) error {
    // ... domain logic ...
    return h.eventPublisher.{{event}}(ctx)
}
```
