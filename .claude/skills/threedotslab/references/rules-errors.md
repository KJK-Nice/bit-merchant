# Error Rules (ERR-01..05)

## Three-Tier Error Architecture

The error system has three tiers:

1. **Domain errors** — sentinel variables and typed structs in `domain/`
2. **Application errors** — `SlugError` with machine-readable slugs in `app/`
3. **Port errors** — protocol-specific error mapping in `ports/`

---

## ERR-01: Sentinel Error Variables (WARNING)

Simple domain errors without context SHOULD use sentinel `var` declarations.

```go
// domain/hour/errors.go
var (
    ErrNotFullHour         = errors.New("hour should be a full hour")
    ErrPastHour            = errors.New("cannot create hour in the past")
    ErrTrainingScheduled   = errors.New("unable to modify hour, because scheduled training")
    ErrHourNotAvailable    = errors.New("hour is not available")
    ErrNoTrainingScheduled = errors.New("no training scheduled")
)
```

**Naming:** `Err{DescriptiveName}` — always starts with `Err`.

**Usage in domain methods:**
```go
func (h *Hour) ScheduleTraining() error {
    if !h.IsAvailable() {
        return ErrHourNotAvailable
    }
    h.availability = TrainingScheduled
    return nil
}
```

---

## ERR-02: Typed Error Structs (WARNING)

Errors that carry context (values for logging/display) SHOULD be typed structs implementing the `error` interface.

```go
type TooDistantDateError struct {
    MaxWeeksInTheFutureToSet int
    ProvidedDate             time.Time
}

func (e TooDistantDateError) Error() string {
    return fmt.Sprintf(
        "schedule can be only set for next %d weeks, provided date: %s",
        e.MaxWeeksInTheFutureToSet, e.ProvidedDate,
    )
}

type TooEarlyHourError struct {
    MinUtcHour int
    ProvidedTime time.Time
}

type ForbiddenToSeeTrainingError struct {
    RequestingUserUUID string
    TrainingOwnerUUID  string
}

type NotFoundError struct {
    TrainingUUID string
}
```

**Naming:** `{Condition}Error` — describes the error condition.

---

## ERR-03: SlugError for Application Layer (WARNING)

Application-layer errors (command/query handlers) SHOULD use `SlugError` from the common errors package. SlugErrors carry:
- Human-readable error message
- Machine-readable slug (used by API clients)
- Error type (authorization, incorrect-input, unknown)

```go
// common/errors/errors.go
type ErrorType struct {
    t string
}

var (
    ErrorTypeUnknown        = ErrorType{"unknown"}
    ErrorTypeAuthorization  = ErrorType{"authorization"}
    ErrorTypeIncorrectInput = ErrorType{"incorrect-input"}
)

type SlugError struct {
    error     string
    slug      string
    errorType ErrorType
}

func NewSlugError(error string, slug string) SlugError
func NewAuthorizationError(error string, slug string) SlugError
func NewIncorrectInputError(error string, slug string) SlugError
```

**Usage in handlers:**
```go
func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
    if err := h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *hour.Hour) (*hour.Hour, error) {
        if err := h.CancelTraining(); err != nil {
            return nil, err
        }
        return h, nil
    }); err != nil {
        return errors.NewSlugError(err.Error(), "unable-to-update-availability")
    }
    return nil
}
```

---

## ERR-04: Error Wrapping with Context (INFO)

When re-raising errors, wrap them with context using `fmt.Errorf("context: %w", err)` or a wrapping library.

```go
// In adapters
if err := doc.DataTo(&model); err != nil {
    return nil, fmt.Errorf("unmarshaling hour from firestore: %w", err)
}
```

---

## ERR-05: No Bare fmt.Errorf in Domain (CRITICAL)

The domain package MUST NOT use `fmt.Errorf` for error creation. Domain errors must be either:
- Sentinel variables (`var ErrX = errors.New(...)`)
- Typed error structs
- Standard `errors.New(...)` for simple cases

**Check:** Grep `domain/` for `fmt.Errorf`. Any match in non-test files is a violation.

**Rationale:** `fmt.Errorf` creates untyped errors that cannot be checked with `errors.Is` or `errors.As`. Domain errors should be programmatically handleable.

**Exception:** `fmt.Errorf` with `%w` for wrapping IS acceptable in domain validation helpers that combine multiple checks, but prefer typed errors or sentinel variables.
