# Watermill Rules (WM-01..10)

## WM-01: Watermill as a Named Component in Unified Server (CRITICAL)

Watermill router MUST be registered as a named component via `server.WithWatermillRouter(name, configure)` — same pattern as `WithHTTPHandler` and `WithGRPCServer`. The `With*` option owns AMQP connection, middleware, and router lifecycle. The caller provides **only handler registration** via callback.

This ensures:
- Middleware stack (retry, correlation, recovery) is consistent across all services
- Broker config is centralized — swapping AMQP for Kafka changes one file
- Shutdown ordering is explicit via `server.OnShutdown(server.Stop(name))`

**Check procedure:**
1. Scan `main.go` for direct Watermill router creation (`message.NewRouter`, `amqp.NewSubscriber`)
2. Flag any middleware setup outside `server/watermill.go`
3. Verify Watermill component appears in `OnShutdown` with correct ordering

**Correct:**
```go
// internal/common/server/watermill.go
func WithWatermillRouter(
    name string,
    configure func(*message.Router, message.Subscriber),
) Option {
    return func(s *Server) {
        wmLogger := watermill.NewStdLoggerWithOut(os.Stdout, true, false)
        amqpURI := os.Getenv("AMQP_URI")
        amqpConfig := amqp.NewDurableQueueConfig(amqpURI)

        sub, err := amqp.NewSubscriber(amqpConfig, wmLogger)
        if err != nil { panic(err) }

        r, err := message.NewRouter(message.RouterConfig{}, wmLogger)
        if err != nil { panic(err) }

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

// main.go — registered as named component
server.New(
    server.WithWatermillRouter("events", func(r *message.Router, sub message.Subscriber) {
        ports.RegisterEventHandlers(r, sub, application)
    }),
    server.WithHTTPHandler("api", createHandler),
    server.OnShutdown(
        server.Stop("events"),      // 1. stop consuming
        server.Stop("api"),         // 2. drain HTTP
        server.StopFunc(cleanup),   // 3. close clients
    ),
).Run(ctx)
```

**Wrong:**
```go
// main.go — VIOLATION: infrastructure in main
func main() {
    sub, _ := amqp.NewSubscriber(amqpConfig, logger)           // VIOLATION
    r, _ := message.NewRouter(message.RouterConfig{}, logger)   // VIOLATION
    r.AddMiddleware(wmMiddleware.Recoverer)                      // VIOLATION
    r.Run(context.Background())
}

// main.go — VIOLATION: standalone RunWatermillRouter without unified server
server.RunWatermillRouter(func(r *message.Router, sub message.Subscriber) { ... })
// Cannot coordinate shutdown with other transports
```

---

## WM-02: Publisher Factory Returns (Publisher, Close, Error) Triple (CRITICAL)

Publisher creation MUST follow the same `(client, closeFunc, error)` triple-return pattern as `client.NewTrainerClient()` and `client.NewUsersClient()`. Config comes from environment variables.

**Check procedure:**
1. Verify publisher factory in `internal/common/client/watermill.go`
2. Must return `(message.Publisher, func() error, error)`
3. Must read `AMQP_URI` from env
4. Error case must return a no-op close function, never nil

**Correct:**
```go
// internal/common/client/watermill.go
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

**Wrong:**
```go
// VIOLATION: returns raw connection, no close function
func NewPublisher() *amqp.Publisher {
    pub, _ := amqp.NewPublisher(config, logger)
    return pub
}

// VIOLATION: nil close function on error path
func NewPublisher() (message.Publisher, func() error, error) {
    // ...
    return nil, nil, err  // nil close panics on defer
}
```

---

## WM-03: Event Handlers Live in Ports (CRITICAL)

Watermill event handlers are **inbound adapters** — they are ports, just like HTTP and gRPC handlers. They MUST:
- Live in `ports/`
- Hold `app.Application`
- Delegate to command/query handlers
- Contain NO business logic

**Check procedure:**
1. Scan for `message.HandlerFunc` or `func(*message.Message) error` signatures
2. These MUST be in `ports/` package
3. Must import `app/`, `app/command/`, or `app/query/` — not `domain/` directly
4. Must follow the same delegation pattern as HTTP/gRPC handlers

**Correct:**
```go
// ports/event.go
type EventHandlers struct {
    app app.Application
}

func RegisterEventHandlers(r *message.Router, sub message.Subscriber, application app.Application) {
    handlers := EventHandlers{app: application}

    r.AddNoPublisherHandler(
        "OnTrainingScheduled",
        "training.scheduled",
        sub,
        handlers.OnTrainingScheduled,
    )
}

func (h EventHandlers) OnTrainingScheduled(msg *message.Message) error {
    var event TrainingScheduledEvent
    if err := json.Unmarshal(msg.Payload, &event); err != nil {
        return err
    }
    return h.app.Commands.ScheduleTraining.Handle(
        msg.Context(),
        command.ScheduleTraining{Hour: event.Hour},
    )
}
```

**Wrong:**
```go
// adapters/event_handler.go — VIOLATION: handler in adapters/
func HandleTrainingScheduled(msg *message.Message) error {
    repo.Save(ctx, training)  // VIOLATION: direct repo access
}

// app/command/schedule_training.go — VIOLATION: message parsing in app layer
func (h handler) Handle(ctx context.Context, msg *message.Message) error { ... }
```

---

## WM-04: Event Publisher Adapter Implements Domain Interface (WARNING)

Publishing events MUST go through an adapter that implements an interface defined in the app or domain layer. The app layer defines *what* events to publish; the adapter knows *how*.

This keeps Watermill as a swappable infrastructure detail.

**Check procedure:**
1. Look for `message.Publisher` usage — it MUST NOT appear in `app/` or `domain/`
2. An interface like `EventPublisher` should be in `app/command/services.go` or similar
3. The concrete adapter in `adapters/` implements it using Watermill

**Correct:**
```go
// app/command/services.go
type TrainingEventPublisher interface {
    TrainingScheduled(ctx context.Context, t training.Training) error
    TrainingCancelled(ctx context.Context, trainingUUID string) error
}

// adapters/training_event_publisher.go
type WatermillTrainingEventPublisher struct {
    pub message.Publisher
}

func NewWatermillTrainingEventPublisher(pub message.Publisher) WatermillTrainingEventPublisher {
    return WatermillTrainingEventPublisher{pub: pub}
}

func (p WatermillTrainingEventPublisher) TrainingScheduled(ctx context.Context, t training.Training) error {
    payload, err := json.Marshal(TrainingScheduledEvent{UUID: t.UUID(), Hour: t.Time()})
    if err != nil { return err }
    msg := message.NewMessage(watermill.NewUUID(), payload)
    middleware.SetCorrelationID(middleware.MessageCorrelationID(msg), msg)
    return p.pub.Publish("training.scheduled", msg)
}
```

**Wrong:**
```go
// app/command/schedule_training.go — VIOLATION: Watermill in app layer
import "github.com/ThreeDotsLabs/watermill/message"

func (h handler) Handle(ctx context.Context, cmd ScheduleTraining) error {
    msg := message.NewMessage(watermill.NewUUID(), payload)  // VIOLATION
    h.publisher.Publish("topic", msg)                         // VIOLATION: infra detail
}
```

---

## WM-05: Topic Naming Uses Domain Language (WARNING)

Topic/queue names MUST use domain language with dot notation: `{aggregate}.{past-tense-event}`. No CRUD names, no technical prefixes.

**Correct:**
```
training.scheduled
training.cancelled
training.reschedule_requested
hour.made_available
```

**Wrong:**
```
create-training          // VIOLATION: CRUD name
events.training.created  // VIOLATION: redundant "events" prefix, CRUD
TRAINING_QUEUE           // VIOLATION: technical name, not domain event
```

---

## WM-06: Event Structs Live in the Publishing Port or Adapter (INFO)

Event DTOs (the JSON payloads) are protocol-specific — they belong in `ports/` or `adapters/`, NOT in `domain/`. Domain entities are the canonical model; events are a serialization concern.

**Check procedure:**
1. Look for event structs (e.g., `TrainingScheduledEvent`)
2. They MUST be in `ports/` (if consumed by event handlers) or `adapters/` (if produced by publisher adapters)
3. They MUST NOT be in `domain/`

**Correct:**
```go
// ports/event.go or adapters/training_event_publisher.go
type TrainingScheduledEvent struct {
    UUID string    `json:"uuid"`
    Hour time.Time `json:"hour"`
}
```

---

## WM-07: Watermill Middleware in With* Option Only (WARNING)

Watermill middleware (retry, correlation ID, recoverer, throttle, etc.) MUST be configured exclusively inside the `WithWatermillRouter` option in `internal/common/server/watermill.go` — same principle as ARCH-06 for HTTP/gRPC middleware.

**Check procedure:**
1. Scan for `r.AddMiddleware` or `router.AddMiddleware` calls
2. All MUST be in `internal/common/server/watermill.go` (inside `WithWatermillRouter`)
3. Flag any middleware setup in `main.go`, `ports/`, or `service/`

---

## WM-08: Publisher Cleanup via OnShutdown or Composition Root (WARNING)

When a service publishes events, the publisher's close function MUST be closed as part of the shutdown sequence. Two valid patterns:

**Pattern A — cleanup in OnShutdown (preferred when using unified server):**
```go
server.New(
    server.WithHTTPHandler("api", createHandler),
    server.OnShutdown(
        server.Stop("api"),         // 1. drain HTTP (in-flight may publish)
        server.StopFunc(cleanup),   // 2. close publisher + clients
    ),
).Run(ctx)
```

**Pattern B — cleanup via defer (simpler services):**
```go
app, cleanup := service.NewApplication(ctx)
defer cleanup()  // runs after Run() returns

server.New(
    server.WithHTTPHandler("api", createHandler),
    server.OnShutdown(
        server.Stop("api"),
    ),
).Run(ctx)
// cleanup() runs here via defer — publisher closes after server drained
```

**Check procedure:**
1. If `service/application.go` creates a publisher, verify close is either in `OnShutdown` or in the cleanup function
2. Publisher close MUST happen *after* all transports that might publish are stopped
3. Closing publisher before draining HTTP/gRPC = lost messages

**Wrong:**
```go
// main.go — VIOLATION: publisher lifecycle in main, not ordered
func main() {
    pub, closePub, _ := client.NewWatermillPublisher()
    defer closePub()                                    // VIOLATION: may close before HTTP drains
    app := service.NewApplication(ctx, pub)             // VIOLATION: infra detail leaked
}
```

---

## WM-09: Named Components Replace SERVER_TO_RUN Switch (INFO)

With the unified server pattern (ARCH-08), the `SERVER_TO_RUN` environment variable switch is replaced by composing `With*` options. A service that needs HTTP + Watermill simply registers both.

**Correct — unified server:**
```go
// All transports in one process, explicit shutdown order
server.New(
    server.WithWatermillRouter("events", func(r *message.Router, sub message.Subscriber) {
        ports.RegisterEventHandlers(r, sub, app)
    }),
    server.WithHTTPHandler("api", func(router chi.Router) http.Handler {
        return ports.HandlerFromMux(ports.NewHttpServer(app), router)
    }),
    server.OnShutdown(
        server.Stop("events"),
        server.Stop("api"),
        server.StopFunc(cleanup),
    ),
).Run(ctx)
```

**Also acceptable — SERVER_TO_RUN for single-transport deployments:**
```go
// When deploying each transport as a separate container
switch serverType {
case "http":
    server.New(
        server.WithHTTPHandler("api", createHandler),
        server.OnShutdown(server.Stop("api")),
    ).Run(ctx)
case "watermill":
    server.New(
        server.WithWatermillRouter("events", configureRouter),
        server.OnShutdown(server.Stop("events")),
    ).Run(ctx)
}
```

---

## WM-10: No Synchronous Side Effects Replaced by Fire-and-Forget (CRITICAL)

When replacing synchronous gRPC calls with async events, you MUST ensure the operation tolerates eventual consistency. If the caller needs confirmation that the action succeeded, keep it synchronous (gRPC) or use a saga/process manager — do NOT simply drop the response.

**Check procedure:**
1. For each gRPC adapter being replaced by events, check if the calling command inspects the return value or error
2. If the command makes decisions based on the result, it MUST remain synchronous or use a compensation pattern
3. Fire-and-forget is only valid for notifications, projections, and truly independent side effects

**Correct use of async:**
```go
// Notification — caller doesn't need the result
func (h handler) Handle(ctx context.Context, cmd ScheduleTraining) error {
    // ... create training ...
    // Fire event — consumer will send email, update dashboard, etc.
    return h.eventPublisher.TrainingScheduled(ctx, training)
}
```

**Wrong use of async:**
```go
// VIOLATION: caller needs confirmation that hours were reserved
func (h handler) Handle(ctx context.Context, cmd ScheduleTraining) error {
    training, _ := training.NewTraining(...)
    h.eventPublisher.TrainingScheduled(ctx, training)  // VIOLATION: no guarantee hours are available
    return h.repo.Save(ctx, training)                   // saved training without confirmed availability
}
// Previously this was a synchronous gRPC call that could fail and roll back
```
