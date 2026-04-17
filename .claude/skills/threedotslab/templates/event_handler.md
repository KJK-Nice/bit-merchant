# Event Handler Scaffold Template

Generate a Watermill event handler port and its registration function. Event handlers are inbound adapters — they live in `ports/` and delegate to CQRS command/query handlers, identical to HTTP and gRPC handlers.

## Placeholders

- `{{Name}}` — PascalCase event name (e.g., `TrainingScheduled`)
- `{{name}}` — camelCase (e.g., `trainingScheduled`)
- `{{name_snake}}` — snake_case (e.g., `training_scheduled`)
- `{{topic}}` — Dot-notation topic name (e.g., `training.scheduled`)
- `{{module}}` — Go module path from go.mod
- `{{command}}` — Command to invoke, PascalCase (e.g., `ScheduleTraining`)

## File: `ports/event.go`

If this file already exists, append the handler method and registration line. If not, create it:

```go
package ports

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"

	"{{module}}/app"
	"{{module}}/app/command"
)

type EventHandlers struct {
	app app.Application
}

func RegisterEventHandlers(r *message.Router, sub message.Subscriber, application app.Application) {
	handlers := EventHandlers{app: application}

	r.AddNoPublisherHandler(
		"On{{Name}}",
		"{{topic}}",
		sub,
		handlers.On{{Name}},
	)
	// TODO: Register additional event handlers here
}

// {{Name}}Event is the event payload DTO — protocol-specific, not a domain object.
type {{Name}}Event struct {
	// TODO: Add event fields matching the publisher's payload
	// Example:
	// UUID string    `json:"uuid"`
	// Hour time.Time `json:"hour"`
}

func (h EventHandlers) On{{Name}}(msg *message.Message) error {
	var event {{Name}}Event
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	// TODO: Construct command and delegate to app layer
	// return h.app.Commands.{{command}}.Handle(msg.Context(), command.{{command}}{
	//     // Map event fields to command fields
	// })

	return nil
}
```

## Update `main.go`

Add `WithWatermillRouter` to the unified server and include it in `OnShutdown`:

```go
server.New(
    server.WithWatermillRouter("events", func(r *message.Router, sub message.Subscriber) {
        ports.RegisterEventHandlers(r, sub, application)
    }),
    server.WithHTTPHandler("api", func(router chi.Router) http.Handler {
        return ports.HandlerFromMux(ports.NewHttpServer(application), router)
    }),
    server.OnShutdown(
        server.Stop("events"),      // 1. stop consuming first
        server.Stop("api"),         // 2. then drain HTTP
        server.StopFunc(cleanup),   // 3. then close clients
    ),
).Run(ctx)
```

## Update `docker-compose.yml`

Add `AMQP_URI` to the service environment (no separate container needed — all transports run in one process):

```yaml
{{service}}:
  environment:
    AMQP_URI: amqp://guest:guest@rabbitmq:5672/
  depends_on:
    - rabbitmq
```
