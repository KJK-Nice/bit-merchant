# Architecture Rules (ARCH-01..08)

## ARCH-01: Standard Directory Layout (CRITICAL)

Every service MUST follow this directory structure:

```
<service>/
├── domain/<aggregate>/   # Pure business logic, entities, value objects, repository interfaces
├── app/                  # Application struct (app.go) with Commands + Queries
│   ├── command/          # Write use cases (command handlers)
│   └── query/            # Read use cases (query handlers + read model interfaces)
├── ports/                # Inbound adapters: HTTP handlers, gRPC servers, CLI
├── adapters/             # Outbound adapters: repository implementations, external clients
└── service/              # Composition root: wires all dependencies together
```

**Check procedure:**
1. Glob for these directories relative to the service root
2. Flag any missing standard directories
3. Flag any non-standard directories at the same level (e.g., `controllers/`, `models/`, `handlers/`)
4. Multiple aggregates can exist under `domain/` as sub-packages (e.g., `domain/hour/`, `domain/training/`)

**Reference (wild-workouts):**
```
internal/trainer/
├── domain/hour/
├── app/
│   ├── command/
│   └── query/
├── ports/
├── adapters/
└── service/
```

---

## ARCH-02: Dependency Direction (CRITICAL)

Dependencies MUST flow inward only: `ports/adapters → app → domain`

The domain layer MUST NOT import from:
- `app/`, `app/command/`, `app/query/`
- `ports/`
- `adapters/`
- Any external infrastructure package (database drivers, HTTP frameworks, etc.)

The app layer MUST NOT import from:
- `ports/`
- `adapters/`

**Check procedure:**
1. For every `.go` file in `domain/`, scan import statements
2. Flag any import that references `app/`, `ports/`, `adapters/`, or the service's own non-domain packages
3. For every `.go` file in `app/`, scan imports for `ports/` or `adapters/`
4. Domain MAY import standard library and pure utility packages

**Allowed domain imports:**
- Standard library (`context`, `time`, `errors`, `fmt`, `strings`, etc.)
- Pure value libraries (e.g., `github.com/google/uuid`)
- NOT: database drivers, HTTP routers, gRPC, logging libraries

---

## ARCH-03: Composition Root Isolation (CRITICAL)

All dependency wiring MUST happen exclusively in `service/`. The composition root is the **only** place that knows about concrete adapter types, infrastructure clients, and how dependencies connect.

**`main.go`** MUST only:
1. Initialize cross-cutting concerns (logging)
2. Call `service.NewApplication()`
3. Wire ports (pass `app.Application` to port constructors)
4. Start the server

`main.go` MUST NOT import `adapters/`, create infrastructure clients, or instantiate command/query handlers directly.

**Check procedure:**
1. Scan `main.go` imports — flag any reference to `adapters/`, database drivers, or external service clients
2. Scan all files outside `service/` — flag any call to adapter constructors (e.g., `adapters.New*`)
3. Verify `service/` returns `app.Application`

**Correct:**
```go
// main.go — only knows about service and ports
func main() {
    logs.Init()
    ctx := context.Background()

    app, cleanup := service.NewApplication(ctx)
    defer cleanup()

    server.RunHTTPServer(func(router chi.Router) http.Handler {
        return ports.HandlerFromMux(ports.NewHttpServer(app), router)
    })
}
```

**Wrong:**
```go
// main.go — VIOLATION: wiring infrastructure directly
func main() {
    client, _ := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))  // VIOLATION
    repo := adapters.NewFirestoreRepository(client)                   // VIOLATION
    handler := command.NewScheduleTrainingHandler(repo, logger, mc)   // VIOLATION
    // ...
}
```

---

## ARCH-04: Dual Constructor Pattern for Testability (WARNING)

The composition root MUST provide two constructors sharing a single private wiring function:
1. **`NewApplication(ctx) (app.Application, func())`** — production constructor, creates real infrastructure
2. **`NewComponentTestApplication(ctx) app.Application`** — test constructor, injects mocks/stubs

Both MUST delegate to a **private** `newApplication(...)` that accepts dependencies as interfaces, so the real vs test paths only differ in what they pass in.

This ensures:
- Test mocks never leak into production wiring
- All wiring logic is shared — no drift between prod and test setups
- The private function signature documents the full set of external dependencies

**Check procedure:**
1. Look for exported `NewApplication` and `NewComponentTestApplication` in `service/`
2. Verify both call the same unexported function
3. The unexported function MUST accept dependencies as interfaces, not concrete types

**Correct:**
```go
// service/service.go
func NewApplication(ctx context.Context) (app.Application, func()) {
    trainerClient, closeTrainer, err := client.NewTrainerClient()
    if err != nil { panic(err) }

    trainerService := adapters.NewTrainerGrpc(trainerClient)

    return newApplication(ctx, trainerService),
        func() { _ = closeTrainer() }
}

func NewComponentTestApplication(ctx context.Context) app.Application {
    return newApplication(ctx, TrainerServiceMock{})
}

func newApplication(ctx context.Context, trainerService command.TrainerService) app.Application {
    // shared wiring logic — accepts interfaces, not concrete types
    repo := adapters.NewFirestoreRepository(client)
    return app.Application{ /* ... */ }
}
```

**Wrong:**
```go
// VIOLATION: separate wiring paths, no shared private function
func NewApplication(ctx context.Context) app.Application {
    repo := adapters.NewFirestoreRepository(client)
    return app.Application{
        Commands: app.Commands{
            ScheduleTraining: command.NewScheduleTrainingHandler(repo, logger, mc),
        },
    }
}

func NewTestApplication() app.Application {
    repo := NewMockRepo()  // VIOLATION: duplicated wiring, can drift
    return app.Application{
        Commands: app.Commands{
            ScheduleTraining: command.NewScheduleTrainingHandler(repo, logger, mc),
        },
    }
}
```

---

## ARCH-05: Cleanup Function for Resource Lifecycle (WARNING)

When the composition root creates resources that require cleanup (connections, clients, subscriptions), `NewApplication` MUST return a cleanup function alongside the application. The caller owns the lifecycle via `defer`.

This ensures:
- Resources are released even on panic
- `main.go` doesn't need to know *what* to clean up — just *that* it must
- Adding new infrastructure only changes `service/`, not `main.go`

**Check procedure:**
1. If `NewApplication` creates closeable resources (clients, connections), it MUST return `func()`
2. `main.go` MUST call `defer cleanup()` immediately after receiving it
3. The cleanup function MUST NOT be ignored (assigned to `_`)

**Correct:**
```go
// service/service.go
func NewApplication(ctx context.Context) (app.Application, func()) {
    trainerClient, closeTrainer, err := client.NewTrainerClient()
    if err != nil { panic(err) }
    usersClient, closeUsers, err := client.NewUsersClient()
    if err != nil { panic(err) }

    return newApplication(ctx, adapters.NewTrainerGrpc(trainerClient), adapters.NewUsersGrpc(usersClient)),
        func() {
            _ = closeTrainer()
            _ = closeUsers()
        }
}

// main.go
app, cleanup := service.NewApplication(ctx)
defer cleanup()
```

**Wrong:**
```go
// VIOLATION: caller must know internals to clean up
func NewApplication(ctx context.Context) (app.Application, *firestore.Client, *grpc.ClientConn) {
    // ...
}

// VIOLATION: cleanup responsibility leaks into main
app, fsClient, conn := service.NewApplication(ctx)
defer fsClient.Close()   // main.go shouldn't know about Firestore
defer conn.Close()        // main.go shouldn't know about gRPC
```

---

## ARCH-06: Server Startup via Callback (WARNING)

Server startup MUST be delegated to a shared `server.Run*Server()` function. `main.go` provides **only the application handler** via a callback. It MUST NOT configure server internals: middleware, routing, listening address, or transport-level concerns.

This ensures:
- Middleware stack (auth, logging, recovery, CORS, security headers) is consistent across all services
- Adding or changing middleware is a single change, not per-service
- `main.go` remains a thin orchestrator: init → wire app → provide handler → run

**Check procedure:**
1. `main.go` MUST call a shared `Run*Server()` function as the final blocking call
2. The callback passed to `Run*Server()` MUST only construct the handler from port constructors — no middleware setup, no router configuration, no listener creation
3. `main.go` MUST NOT import server infrastructure packages (e.g., `net/http.ListenAndServe`, `net.Listen`, middleware libraries)

**Correct:**
```go
// main.go — provides handler, delegates everything else
func main() {
    logs.Init()
    ctx := context.Background()

    app, cleanup := service.NewApplication(ctx)
    defer cleanup()

    server.RunHTTPServer(func(router chi.Router) http.Handler {
        return ports.HandlerFromMux(ports.NewHttpServer(app), router)
    })
}
```

**Wrong:**
```go
// VIOLATION: main.go configures server internals
func main() {
    app, cleanup := service.NewApplication(ctx)
    defer cleanup()

    router := chi.NewRouter()
    router.Use(middleware.Logger)           // VIOLATION: middleware in main
    router.Use(middleware.Recoverer)        // VIOLATION: middleware in main
    router.Mount("/api", ports.NewHttpServer(app))

    http.ListenAndServe(":8080", router)   // VIOLATION: listening in main
}
```

---

## ARCH-07: Composition Root Must Not Own Server Lifecycle (CRITICAL)

The `service/` package wires dependencies and returns `app.Application`. It MUST NOT create transport servers, bind to network ports, handle OS signals, or manage graceful shutdown. Server lifecycle is a **separate concern** that belongs in a shared server package or the entry point.

`service/` MUST NOT:
- Create transport servers (`grpc.NewServer()`, `http.Server{}`, `message.NewRouter()`)
- Bind to network ports (`net.Listen()`)
- Handle OS signals (`signal.NotifyContext()`, `signal.Notify()`)
- Manage graceful shutdown (`GracefulStop()`, `router.Close()`)
- Import port packages (`ports/grpc`, `ports/amqp`, `ports/http`)

`service/` MUST only:
- Create infrastructure clients and adapters
- Wire command/query handlers with dependencies
- Return `app.Application` (and optionally a cleanup function)

**Check procedure:**
1. Scan all files in `service/` for imports of `net`, `os/signal`, `syscall`, transport packages, or `ports/`
2. Flag any function in `service/` that accepts or creates a server, listener, or router
3. A file named `server.go` in `service/` is a strong signal of violation

**Correct:**
```go
// service/service.go — only wires the application
func NewApplication(ctx context.Context, cfg *config.Config) (app.Application, func()) {
    repo := adapters.NewFirestoreRepository(client)
    syncer := tokensync.NewSyncer(fetchers, syncRepo, progressTracker)

    return newApplication(repo, syncer),
        func() { _ = client.Close() }
}

// Server lifecycle lives elsewhere (shared server package or entry point)
```

**Wrong:**
```go
// service/server.go — VIOLATION: server lifecycle in composition root
func RunServer(application app.Application, cfg *config.Config) error {
    ctx, stop := signal.NotifyContext(context.Background(), ...)  // VIOLATION: signal handling
    defer stop()

    grpcServer := grpc.NewServer()                                // VIOLATION: transport server
    pb.RegisterCommandsServer(grpcServer, ports.NewServer(app))   // VIOLATION: imports ports/

    lis, _ := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))    // VIOLATION: network binding
    go grpcServer.Serve(lis)                                      // VIOLATION: server lifecycle

    <-ctx.Done()
    grpcServer.GracefulStop()                                     // VIOLATION: shutdown management
    return nil
}
```

---

## ARCH-08: Unified Server with Named Components and OnShutdown (WARNING)

When a project has multiple transports (gRPC, HTTP, AMQP/Watermill), the shared server package SHOULD provide a **single `server.New(...).Run(ctx)`** with functional options per transport and an explicit `OnShutdown` that declares the shutdown sequence.

### Why explicit shutdown ordering matters

Different services have different dependency graphs between transports:
- A consumer that calls gRPC must stop consuming *before* gRPC clients close
- An HTTP API that publishes events must drain HTTP *before* the publisher closes
- Two independent ingress points (HTTP + gRPC) can shut down in parallel

Implicit ordering (LIFO based on registration) is fragile — reordering lines silently changes shutdown behavior. `OnShutdown` makes the sequence a readable, reviewable declaration.

### Core types

```go
// server/server.go
type Server struct {
    components map[string]component
    startOrder []string
    shutdownSteps []ShutdownStep
}

type component struct {
    name  string
    start func(ctx context.Context) error
    stop  func(ctx context.Context) error
}

type Option func(*Server)

type ShutdownStep struct {
    componentNames []string
    fn             func(ctx context.Context) error
}
```

### API

```go
// Stop creates a step that stops named components.
// Multiple names = parallel shutdown within the step.
func Stop(names ...string) ShutdownStep

// StopFunc creates a step that runs an arbitrary cleanup function.
func StopFunc(fn func()) ShutdownStep

// StopFuncWithErr creates a step with error return.
func StopFuncWithErr(fn func(ctx context.Context) error) ShutdownStep

// OnShutdown declares the shutdown sequence.
// Steps execute top-to-bottom. Each step completes before the next starts.
// Components not mentioned stop last (with a warning log).
func OnShutdown(steps ...ShutdownStep) Option
```

### Shutdown execution

1. Steps execute sequentially in declaration order
2. Within a `Stop("a", "b")` call, components stop in parallel
3. Each step's `wg.Wait()` completes before the next step begins
4. Components not mentioned in any `Stop()` get a catch-all parallel stop after all explicit steps (with a warning log — every component should be in OnShutdown)
5. A global timeout (default 30s) bounds the entire sequence

### Key design principles

- Each `With*` option takes a `name string` as first argument — used in `Stop(name)` to reference it
- `OnShutdown` reads top-to-bottom as a shutdown script
- The factory owns `signal.NotifyContext` — callers never handle signals
- `defer cleanup()` from `NewApplication` naturally runs after `Run()` returns — it is the implicit last phase
- Duplicate component names panic at startup — caught immediately

**Check procedure:**
1. If a project uses 2+ transports, verify `server.New()` is used (not multiple `Run*Server` calls)
2. Verify `OnShutdown` is present and lists all components
3. Verify shutdown order makes sense: consumers before servers, servers before clients
4. No `signal.NotifyContext`, `net.Listen`, or `GracefulStop` calls outside `common/server/`

**Correct:**
```go
// Trainer: HTTP + gRPC + Watermill consumer
func main() {
    logs.Init()
    ctx := context.Background()

    app, cleanup := service.NewApplication(ctx)
    defer cleanup()

    server.New(
        server.WithWatermillRouter("events", func(r *message.Router, sub message.Subscriber) {
            ports.RegisterEventHandlers(r, sub, app)
        }),
        server.WithHTTPHandler("api", func(router chi.Router) http.Handler {
            return ports.HandlerFromMux(ports.NewHttpServer(app), router)
        }),
        server.WithGRPCServer("grpc", func(s *grpc.Server) {
            trainer.RegisterTrainerServiceServer(s, ports.NewGrpcServer(app))
        }),
        server.OnShutdown(
            server.Stop("events"),          // 1. stop consuming
            server.Stop("api", "grpc"),     // 2. drain both servers in parallel
            server.StopFunc(cleanup),       // 3. close clients & publisher
        ),
    ).Run(ctx)
}

// Trainings: HTTP-only, publishes events (publisher in cleanup)
func main() {
    logs.Init()
    ctx := context.Background()

    app, cleanup := service.NewApplication(ctx)
    defer cleanup()

    server.New(
        server.WithHTTPHandler("api", func(router chi.Router) http.Handler {
            return ports.HandlerFromMux(ports.NewHttpServer(app), router)
        }),
        server.OnShutdown(
            server.Stop("api"),       // 1. drain HTTP (in-flight may publish events)
            server.StopFunc(cleanup), // 2. close publisher + gRPC clients
        ),
    ).Run(ctx)
}
```

**Wrong:**
```go
// VIOLATION: implicit LIFO ordering — fragile
server.New(
    server.WithHTTPHandler("api", createHandler),
    server.WithWatermillRouter("events", configureRouter),
    // no OnShutdown — relies on registration order
).Run(ctx)

// VIOLATION: manual lifecycle per transport
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

    grpcServer := grpc.NewServer()
    go grpcServer.Serve(lis)

    router, _ := message.NewRouter(...)
    go router.Run(ctx)

    <-ctx.Done()
    grpcServer.GracefulStop()
    router.Close()
}
```
