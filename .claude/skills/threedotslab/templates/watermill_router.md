# Watermill Router Option + Publisher Client Scaffold Template

Generate the `WithWatermillRouter` server option in `internal/common/server/` and the publisher client factory in `internal/common/client/`. Requires the unified server scaffold (`/threedots scaffold unified_server`) to be in place first.

## Placeholders

- `{{module_common}}` — Go module path to `internal/common` (e.g., `github.com/example/myproject/internal/common`)

## File 1: `internal/common/server/watermill.go`

```go
package server

import (
	"context"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	wmMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func WithWatermillRouter(
	name string,
	configure func(*message.Router, message.Subscriber),
) Option {
	return func(s *Server) {
		wmLogger := watermill.NewStdLoggerWithOut(os.Stdout, true, false)

		amqpURI := os.Getenv("AMQP_URI")
		if amqpURI == "" {
			amqpURI = "amqp://guest:guest@rabbitmq:5672/"
		}
		amqpConfig := amqp.NewDurableQueueConfig(amqpURI)

		sub, err := amqp.NewSubscriber(amqpConfig, wmLogger)
		if err != nil {
			panic("cannot create watermill subscriber: " + err.Error())
		}

		r, err := message.NewRouter(message.RouterConfig{}, wmLogger)
		if err != nil {
			panic("cannot create watermill router: " + err.Error())
		}

		r.AddMiddleware(
			wmMiddleware.CorrelationID,
			wmMiddleware.Recoverer,
			wmMiddleware.Retry{MaxRetries: 3}.Middleware,
		)

		configure(r, sub)

		s.addComponent(name, component{
			name: name,
			start: func(ctx context.Context) error {
				return r.Run(ctx)
			},
			stop: func(ctx context.Context) error {
				return r.Close()
			},
		})
	}
}
```

## File 2: `internal/common/client/watermill.go`

```go
package client

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

func NewWatermillPublisher() (pub message.Publisher, close func() error, err error) {
	amqpURI := os.Getenv("AMQP_URI")
	if amqpURI == "" {
		return nil, func() error { return nil }, errors.New("empty env AMQP_URI")
	}

	logger := watermill.NewStdLoggerWithOut(os.Stdout, true, false)
	config := amqp.NewDurableQueueConfig(amqpURI)

	publisher, err := amqp.NewPublisher(config, logger)
	if err != nil {
		return nil, func() error { return nil }, errors.Wrap(err, "cannot create watermill publisher")
	}

	return publisher, publisher.Close, nil
}
```

## Post-Creation Instructions

After creating the Watermill option and publisher:

1. Add `github.com/ThreeDotsLabs/watermill` and `github.com/ThreeDotsLabs/watermill-amqp/v3` to `go.mod`
2. Add `AMQP_URI` to `.env`, `.test.env`, and `docker-compose.yml`
3. Add a RabbitMQ service to `docker-compose.yml`:
   ```yaml
   rabbitmq:
     image: rabbitmq:3-management
     ports:
       - "5672:5672"
       - "15672:15672"
   ```
4. Use `/threedots scaffold event_handler <Name>` to create event handlers in a service
5. Use `/threedots scaffold event_publisher <Name>` to create a publisher adapter
6. Add `server.WithWatermillRouter("events", ...)` and include `"events"` in `OnShutdown`
