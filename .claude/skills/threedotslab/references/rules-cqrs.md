# CQRS Rules (CQRS-01..10)

## CQRS-01: Command Struct Pattern (CRITICAL)

Commands MUST be:
- Named with imperative verb + noun (domain language, NOT CRUD)
- Plain data structs (no methods, no interfaces)
- Their handler returns `error` only — no data

**Correct:**
```go
type ScheduleTraining struct {
    Hour time.Time
}

type CancelTraining struct {
    Hour time.Time
}

type MakeHoursAvailable struct {
    Hours []time.Time
}
```

**Wrong:**
```go
type CreateTraining struct { ... }  // VIOLATION: CRUD naming
type UpdateHour struct { ... }      // VIOLATION: CRUD naming
```

---

## CQRS-02: Query Struct Pattern (CRITICAL)

Queries MUST be:
- Named with noun phrases (NOT "Get" + noun)
- Plain data structs
- Their handler returns `(ResultType, error)`

**Correct:**
```go
type AvailableHours struct {
    From time.Time
    To   time.Time
}

type HourAvailability struct {
    Hour time.Time
}
```

**Wrong:**
```go
type GetAvailableHours struct { ... }  // VIOLATION: "Get" prefix
type FetchTrainings struct { ... }     // VIOLATION: "Fetch" prefix
```

---

## CQRS-03: Exported Handler Type Alias (WARNING)

Each handler file MUST define an exported type alias using the generic decorator interface.

```go
// For commands:
type CancelTrainingHandler decorator.CommandHandler[CancelTraining]

// For queries:
type AvailableHoursHandler decorator.QueryHandler[AvailableHours, []Date]
```

This allows callers to depend on the decorated interface, not the concrete struct.

---

## CQRS-04: Unexported Handler Struct (WARNING)

The concrete handler struct MUST be unexported (lowercase). It holds dependencies injected via constructor.

```go
type cancelTrainingHandler struct {
    hourRepo hour.Repository
}

type availableHoursHandler struct {
    readModel AvailableHoursReadModel
}
```

---

## CQRS-05: Constructor Wraps with Decorators (WARNING)

Handler constructors MUST wrap the concrete handler with `ApplyCommandDecorators` or `ApplyQueryDecorators`.

```go
func NewCancelTrainingHandler(
    hourRepo hour.Repository,
    logger *logrus.Entry,
    metricsClient decorator.MetricsClient,
) CancelTrainingHandler {
    return decorator.ApplyCommandDecorators[CancelTraining](
        cancelTrainingHandler{hourRepo: hourRepo},
        logger,
        metricsClient,
    )
}

func NewAvailableHoursHandler(
    readModel AvailableHoursReadModel,
    logger *logrus.Entry,
    metricsClient decorator.MetricsClient,
) AvailableHoursHandler {
    return decorator.ApplyQueryDecorators[AvailableHours, []Date](
        availableHoursHandler{readModel: readModel},
        logger,
        metricsClient,
    )
}
```

---

## CQRS-06: Constructor Nil-Checks with Panic (WARNING)

Handler constructors SHOULD nil-check all injected dependencies and panic if any are nil. This is a fail-fast pattern — misconfiguration is caught at startup, not at runtime.

```go
func NewCancelTrainingHandler(
    hourRepo hour.Repository,
    logger *logrus.Entry,
    metricsClient decorator.MetricsClient,
) CancelTrainingHandler {
    if hourRepo == nil {
        panic("nil hourRepo")
    }
    if logger == nil {
        panic("nil logger")
    }
    if metricsClient == nil {
        panic("nil metricsClient")
    }
    return decorator.ApplyCommandDecorators[CancelTraining](
        cancelTrainingHandler{hourRepo: hourRepo},
        logger,
        metricsClient,
    )
}
```

---

## CQRS-07: Application Struct (CRITICAL)

The `app/app.go` file MUST define an `Application` struct that bundles `Commands` and `Queries` sub-structs.

```go
type Application struct {
    Commands Commands
    Queries  Queries
}

type Commands struct {
    CancelTraining       command.CancelTrainingHandler
    ScheduleTraining     command.ScheduleTrainingHandler
    MakeHoursAvailable   command.MakeHoursAvailableHandler
    MakeHoursUnavailable command.MakeHoursUnavailableHandler
}

type Queries struct {
    HourAvailability      query.HourAvailabilityHandler
    TrainerAvailableHours query.AvailableHoursHandler
}
```

**Check:** Look for `app.go` in the `app/` package. Verify it has `Application`, `Commands`, and `Queries` types.

---

## CQRS-08: Read Model Interface for Queries (WARNING)

Query handlers SHOULD depend on a dedicated read model interface, not the write repository.

```go
// In app/query/ — defines what it needs
type AvailableHoursReadModel interface {
    AvailableHours(ctx context.Context, from, to time.Time) ([]Date, error)
}
```

This keeps reads and writes separate. The same adapter may implement both the write `Repository` and a read model interface, but the query handler only knows about the read model.

---

## CQRS-09: Command/Query Separation (CRITICAL)

- **Commands** MUST modify state and return only `error`
- **Queries** MUST read state and return `(ResultType, error)` — they MUST NOT modify state

A handler that both reads and writes violates CQRS.

**Check:** Command handlers returning anything besides `error` is a violation. Query handlers calling mutation methods on repositories is a violation.

---

## CQRS-10: No Business Logic in Handlers (WARNING)

Handlers are orchestrators. Business rules live in domain entities.

**Correct** — handler delegates to domain:
```go
func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
    return h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *hour.Hour) (*hour.Hour, error) {
        if err := h.CancelTraining(); err != nil {  // domain method
            return nil, err
        }
        return h, nil
    })
}
```

**Wrong** — business logic in handler:
```go
func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
    hour, _ := h.hourRepo.GetHour(ctx, cmd.Hour)
    if hour.Availability != "training_scheduled" {  // VIOLATION: logic belongs in domain
        return errors.New("no training to cancel")
    }
    hour.Availability = "available"  // VIOLATION: direct field mutation
    return h.hourRepo.Save(ctx, hour)
}
```

---

## Complete Handler File Template

Every command/query handler file follows this 4-component pattern:

```go
package command

// 1. Command struct
type CancelTraining struct {
    Hour time.Time
}

// 2. Exported handler type (alias to decorator interface)
type CancelTrainingHandler decorator.CommandHandler[CancelTraining]

// 3. Unexported concrete handler
type cancelTrainingHandler struct {
    hourRepo hour.Repository
}

// 4. Constructor with nil-checks + decorator wrapping
func NewCancelTrainingHandler(
    hourRepo hour.Repository,
    logger *logrus.Entry,
    metricsClient decorator.MetricsClient,
) CancelTrainingHandler {
    if hourRepo == nil {
        panic("nil hourRepo")
    }
    return decorator.ApplyCommandDecorators[CancelTraining](
        cancelTrainingHandler{hourRepo: hourRepo},
        logger,
        metricsClient,
    )
}

// Handle method on unexported struct
func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
    // orchestration only — delegate to domain
}
```
