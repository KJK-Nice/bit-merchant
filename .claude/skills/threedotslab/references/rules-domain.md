# Domain Rules (DOM-01..09)

## DOM-01: Private Entity Fields (CRITICAL)

ALL entity struct fields MUST be unexported (lowercase). Entities are "types with behavior," not data bags.

**Check:** Scan all structs in `domain/` for exported fields. Any uppercase field name is a violation.

**Correct:**
```go
type Hour struct {
    hour         time.Time
    availability Availability
}
```

**Wrong:**
```go
type Hour struct {
    Hour         time.Time     // VIOLATION: exported field
    Availability Availability  // VIOLATION: exported field
}
```

**Exception:** DB model structs in `adapters/` MAY have exported fields for serialization tags.

---

## DOM-02: Factory Constructors (WARNING)

Entities MUST be created through factory constructors, never by direct struct literal.

Pattern: `func New{Type}(args...) (*Type, error)`

The constructor:
- Validates all invariants
- Returns an error if validation fails
- Returns a pointer to the new entity

**Reference:**
```go
func NewAvailableHour(hour time.Time) (*Hour, error) {
    if err := validateTime(hour); err != nil {
        return nil, err
    }
    return &Hour{hour: hour, availability: Available}, nil
}

func NewTraining(uuid, userUUID, userName string, trainingTime time.Time) (*Training, error) {
    if uuid == "" {
        return nil, errors.New("empty training uuid")
    }
    if userUUID == "" {
        return nil, errors.New("empty training user uuid")
    }
    // ... validate all fields
    return &Training{uuid: uuid, userUUID: userUUID, userName: userName, time: trainingTime}, nil
}
```

---

## DOM-03: MustNew Panic Constructors (INFO)

For use in tests and initialization code, provide `MustNew{Type}` that panics on error.

```go
func MustNewFactory(fc FactoryConfig) Factory {
    f, err := NewFactory(fc)
    if err != nil {
        panic(err)
    }
    return f
}
```

---

## DOM-04: UnmarshalFromDatabase (WARNING)

Entities MUST provide an `Unmarshal{Type}FromDatabase` function for reconstruction from persistence. This function:
- Bypasses normal validation (data was already valid when stored)
- Accepts all fields needed to reconstruct full state
- Is used ONLY by repository adapters

**Reference:**
```go
func UnmarshalHourFromDatabase(hour time.Time, availability Availability) *Hour {
    return &Hour{hour: hour, availability: availability}
}

func UnmarshalTrainingFromDatabase(
    uuid, userUUID, userName string,
    trainingTime time.Time,
    notes string,
    canceled bool,
    proposedNewTime time.Time,
    moveProposedBy UserType,
) (*Training, error) {
    return &Training{
        uuid: uuid, userUUID: userUUID, userName: userName,
        time: trainingTime, notes: notes, canceled: canceled,
        proposedNewTime: proposedNewTime, moveProposedBy: moveProposedBy,
    }, nil
}
```

---

## DOM-05: Value Objects as Structs (CRITICAL)

Value objects MUST be structs wrapping a private field, NOT raw strings, ints, or type aliases.

This ensures they cannot be constructed with arbitrary values — only through validated constructors or predefined constants.

**Correct:**
```go
type Availability struct {
    a string  // private — cannot be set directly
}

var (
    Available         = Availability{"available"}
    NotAvailable      = Availability{"not_available"}
    TrainingScheduled = Availability{"training_scheduled"}
)

type UserType struct {
    s string
}

var (
    Trainer  = UserType{"trainer"}
    Attendee = UserType{"attendee"}
)
```

**Wrong:**
```go
type Availability string  // VIOLATION: can be set to any string

const (
    Available    Availability = "available"
    NotAvailable Availability = "not_available"
)
```

---

## DOM-06: IsZero Method (WARNING)

Value objects and factory structs SHOULD implement `IsZero() bool` to check for zero-value state.

```go
func (a Availability) IsZero() bool {
    return a == Availability{}
}

func (f Factory) IsZero() bool {
    return f == Factory{}
}
```

---

## DOM-07: Behavior Methods Use Domain Language (CRITICAL)

Entity methods MUST use domain-specific language, NOT generic CRUD terms.

| Forbidden | Use Instead |
|-----------|------------|
| `SetStatus`, `Update` | `ScheduleTraining`, `CancelTraining`, `MakeAvailable` |
| `Create` | `Schedule`, `Register`, `Place`, `Submit` |
| `Delete` | `Cancel`, `Archive`, `Revoke` |
| `Get` | Use query noun phrases |

**Reference:**
```go
func (h *Hour) ScheduleTraining() error {
    if !h.IsAvailable() {
        return ErrHourNotAvailable
    }
    h.availability = TrainingScheduled
    return nil
}

func (h *Hour) CancelTraining() error { ... }
func (h *Hour) MakeAvailable() error { ... }
func (h *Hour) MakeNotAvailable() error { ... }

func (t *Training) ProposeReschedule(newTime time.Time, proposedBy UserType) error { ... }
func (t *Training) ApproveReschedule(approvedBy UserType) error { ... }
func (t *Training) RejectReschedule() error { ... }
```

---

## DOM-08: String Constructors Validate Input (WARNING)

When a value object can be constructed from a string, use `New{Type}FromString` with validation.

```go
func NewAvailabilityFromString(availabilityStr string) (Availability, error) {
    switch availabilityStr {
    case "available":
        return Available, nil
    case "not_available":
        return NotAvailable, nil
    case "training_scheduled":
        return TrainingScheduled, nil
    default:
        return Availability{}, fmt.Errorf("unknown availability: %s", availabilityStr)
    }
}
```

---

## DOM-09: Factory Struct for Complex Creation (INFO)

When entity creation requires configuration or external dependencies, use a Factory struct pattern.

```go
type FactoryConfig struct {
    MaxWeeksInTheFutureToSet int
    MinUtcHour               int
    MaxUtcHour               int
}

func (c FactoryConfig) Validate() error {
    var errs []error
    if c.MaxWeeksInTheFutureToSet <= 0 {
        errs = append(errs, errors.New("MaxWeeksInTheFutureToSet must be > 0"))
    }
    // ... more validations
    return multierr.Combine(errs...)
}

type Factory struct {
    fc FactoryConfig
}

func NewFactory(fc FactoryConfig) (Factory, error) {
    if err := fc.Validate(); err != nil {
        return Factory{}, err
    }
    return Factory{fc: fc}, nil
}

func MustNewFactory(fc FactoryConfig) Factory {
    f, err := NewFactory(fc)
    if err != nil {
        panic(err)
    }
    return f
}

func (f Factory) IsZero() bool {
    return f == Factory{}
}

func (f Factory) NewAvailableHour(hour time.Time) (*Hour, error) {
    // uses f.fc for validation bounds
}
```
