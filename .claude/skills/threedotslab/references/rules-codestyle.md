# Code Style Rules (STYLE-01..08)

## STYLE-01: Import Grouping (INFO)

Imports MUST be organized in groups separated by blank lines:
1. Standard library
2. External packages (third-party + internal modules)

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/sirupsen/logrus"
    "github.com/example/myproject/internal/trainer/domain/hour"
)
```

**Wrong:**
```go
import (
    "context"
    "github.com/sirupsen/logrus"  // VIOLATION: mixed with stdlib
    "fmt"
    "time"
)
```

---

## STYLE-02: Receiver Conventions (INFO)

- **Pointer receivers** (`*Type`) for methods that mutate state
- **Value receivers** (`Type`) for methods that only read state

```go
// Mutates — pointer receiver
func (h *Hour) ScheduleTraining() error {
    h.availability = TrainingScheduled
    return nil
}

// Read-only — value receiver
func (h Hour) IsAvailable() bool {
    return h.availability == Available
}

func (a Availability) IsZero() bool {
    return a == Availability{}
}
```

Receiver names should be short (1-2 chars), typically the first letter of the type.

---

## STYLE-03: t.Parallel() in Tests (WARNING)

Every test function and subtest SHOULD call `t.Parallel()` as its first statement.

```go
func TestScheduleTraining(t *testing.T) {
    t.Parallel()
    // ... test code
}

func TestRepository(t *testing.T) {
    t.Parallel()
    for i := range testCases {
        tc := testCases[i]
        t.Run(tc.Name, func(t *testing.T) {
            t.Parallel()
            // ... test code
        })
    }
}
```

---

## STYLE-04: require vs assert (INFO)

Use the testify library with:
- **`require`** for setup/preconditions that must succeed (fatal on failure)
- **`assert`** for actual test assertions (non-fatal, continues test)

```go
func TestSomething(t *testing.T) {
    // Setup — use require (fatal if fails)
    hour, err := hour.NewAvailableHour(testTime)
    require.NoError(t, err)

    // Act
    err = hour.ScheduleTraining()

    // Assert — use assert (non-fatal)
    assert.NoError(t, err)
    assert.Equal(t, hour.TrainingScheduled, hour.Availability())
}
```

---

## STYLE-05: Loop Variable Capture (WARNING)

When using loop variables in goroutines or subtests, ALWAYS capture them first.

```go
for i := range repositories {
    r := repositories[i]  // capture before subtest
    t.Run(r.Name, func(t *testing.T) {
        t.Parallel()
        testUpdateHour(t, r.Repository)
    })
}
```

**Note:** Go 1.22+ fixes loop variable capture for `range` loops, but the explicit capture pattern is still preferred for clarity and backward compatibility.

---

## STYLE-06: Table-Driven Tests (INFO)

Tests with multiple cases SHOULD use table-driven pattern with named test cases.

```go
func TestValidateTime(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        Name        string
        Hour        time.Time
        ExpectedErr error
    }{
        {
            Name:        "valid_hour",
            Hour:        time.Now().Truncate(time.Hour).Add(24 * time.Hour),
            ExpectedErr: nil,
        },
        {
            Name:        "past_hour",
            Hour:        time.Now().Add(-time.Hour),
            ExpectedErr: ErrPastHour,
        },
        {
            Name:        "not_full_hour",
            Hour:        time.Now().Add(30 * time.Minute),
            ExpectedErr: ErrNotFullHour,
        },
    }

    for i := range testCases {
        tc := testCases[i]
        t.Run(tc.Name, func(t *testing.T) {
            t.Parallel()
            err := validateTime(tc.Hour)
            assert.ErrorIs(t, err, tc.ExpectedErr)
        })
    }
}
```

---

## STYLE-07: Interfaces Where Consumed (WARNING)

Interfaces MUST be defined in the package that **uses** them, not the package that implements them. This follows Go's implicit interface philosophy.

**Correct:**
```go
// domain/hour/repository.go — consumer defines what it needs
package hour

type Repository interface {
    GetHour(ctx context.Context, hourTime time.Time) (*Hour, error)
    UpdateHour(ctx context.Context, hourTime time.Time,
        updateFn func(h *Hour) (*Hour, error)) error
}

// adapters/ — implicitly implements it
package adapters

type FirestoreHourRepository struct { ... }
func (r *FirestoreHourRepository) GetHour(...) (*hour.Hour, error) { ... }
func (r *FirestoreHourRepository) UpdateHour(...) error { ... }
```

**Wrong:**
```go
// adapters/interfaces.go  ← VIOLATION
package adapters

type HourRepository interface { ... }  // interface where implemented, not consumed
```

---

## STYLE-08: Context as First Parameter (WARNING)

All methods that perform I/O (database, HTTP, gRPC, file) MUST accept `context.Context` as their first parameter.

```go
// Repository methods
GetHour(ctx context.Context, hourTime time.Time) (*Hour, error)
UpdateHour(ctx context.Context, hourTime time.Time, updateFn func(h *Hour) (*Hour, error)) error

// Handler methods
Handle(ctx context.Context, cmd CancelTraining) error
Handle(ctx context.Context, q AvailableHours) ([]Date, error)

// Adapter methods
func (r *FirestoreHourRepository) GetHour(ctx context.Context, hourTime time.Time) (*hour.Hour, error)
```

**Wrong:**
```go
func (r *Repo) GetHour(hourTime time.Time) (*Hour, error)  // VIOLATION: no context
func (r *Repo) GetHour(hourTime time.Time, ctx context.Context) (*Hour, error)  // VIOLATION: ctx not first
```
