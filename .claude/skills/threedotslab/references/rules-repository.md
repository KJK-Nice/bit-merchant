# Repository Rules (REPO-01..07)

## REPO-01: Interface Defined in Domain (CRITICAL)

Repository interfaces MUST be defined in the domain package, next to the entity they persist. This follows the Dependency Inversion Principle — the domain defines what it needs, adapters implement it.

**Correct:**
```go
// domain/hour/repository.go
package hour

type Repository interface {
    GetHour(ctx context.Context, hourTime time.Time) (*Hour, error)
    UpdateHour(ctx context.Context, hourTime time.Time,
        updateFn func(h *Hour) (*Hour, error)) error
}
```

**Wrong:**
```go
// adapters/repository.go  ← VIOLATION: interface in adapter layer
package adapters

type HourRepository interface { ... }
```

**Check:** Grep `domain/` for `type.*Repository interface`. Grep `adapters/` for the same — if found in adapters, it's a violation.

---

## REPO-02: Update Callback Pattern (WARNING)

Repository update methods SHOULD use a callback/closure pattern. The repository handles transaction lifecycle; the callback handles domain logic.

```go
// Interface
UpdateHour(ctx context.Context, hourTime time.Time,
    updateFn func(h *Hour) (*Hour, error)) error

// Usage in handler
err := h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *hour.Hour) (*hour.Hour, error) {
    if err := h.CancelTraining(); err != nil {
        return nil, err
    }
    return h, nil
})
```

Benefits:
- Transaction scope is clear
- Domain logic is isolated from persistence details
- Enables optimistic locking, retries, etc. transparently

---

## REPO-03: Separate DB Model Structs (WARNING)

Adapter implementations MUST use separate structs for database representation. Domain entities should NOT have serialization tags.

**Correct:**
```go
// adapters/ — DB model
type mysqlHour struct {
    ID           int       `db:"id"`
    Hour         time.Time `db:"hour"`
    Availability string    `db:"availability"`
}

// or for Firestore (needs exported fields for tags)
type TrainingModel struct {
    UUID            string    `firestore:"Uuid"`
    UserUUID        string    `firestore:"UserUuid"`
    Time            time.Time `firestore:"Time"`
}

// Conversion in adapter
func (r *MySQLHourRepository) toHour(m mysqlHour) (*hour.Hour, error) {
    availability, err := hour.NewAvailabilityFromString(m.Availability)
    if err != nil {
        return nil, err
    }
    return hour.UnmarshalHourFromDatabase(m.Hour, availability), nil
}
```

**Wrong:**
```go
// domain/hour/hour.go
type Hour struct {
    Hour         time.Time `json:"hour" db:"hour"`  // VIOLATION: DB tags on domain entity
    Availability string    `json:"availability"`     // VIOLATION: serialization concern in domain
}
```

---

## REPO-04: Adapter Constructor Naming (INFO)

Repository adapter constructors follow: `New{Technology}{Entity}Repository`

```go
func NewFirestoreHourRepository(client *firestore.Client, factory hour.Factory) *FirestoreHourRepository
func NewMySQLHourRepository(db *sqlx.DB) *MySQLHourRepository
func NewMemoryHourRepository(factory hour.Factory) *MemoryHourRepository
```

---

## REPO-05: Technology Suffix Naming (INFO)

Adapter types use technology as a suffix/prefix to distinguish implementations.

```go
type FirestoreHourRepository struct { ... }
type MySQLHourRepository struct { ... }
type MemoryHourRepository struct { ... }

// For external service clients
type TrainerGrpc struct { ... }
type UsersGrpc struct { ... }
```

---

## REPO-06: Shared Test Suite (WARNING)

Repository tests SHOULD run the same test logic against ALL implementations (memory, MySQL, Firestore, etc.). This ensures behavioral consistency.

**Pattern:**
```go
func createRepositories(t *testing.T) []Repository {
    return []Repository{
        {Name: "Firebase", Repository: newFirebaseRepository(t)},
        {Name: "MySQL", Repository: newMySQLRepository(t)},
        {Name: "memory", Repository: adapters.NewMemoryHourRepository(testFactory)},
    }
}

func TestRepository(t *testing.T) {
    repositories := createRepositories(t)
    for i := range repositories {
        r := repositories[i]  // capture loop variable
        t.Run(r.Name, func(t *testing.T) {
            t.Parallel()
            testUpdateHour(t, r.Repository)
            testUpdateHour_parallel(t, r.Repository)
        })
    }
}
```

**Check:** Look for test files in `adapters/` that test repository implementations. Verify they use a shared test function or table-driven approach.

---

## REPO-07: UnmarshalFromDatabase Usage (WARNING)

Adapter implementations MUST use the entity's `UnmarshalFromDatabase` function to reconstruct domain objects from persistence, not the regular constructor.

**Correct:**
```go
func (r *FirestoreHourRepository) toHour(doc *firestore.DocumentSnapshot) (*hour.Hour, error) {
    var m HourModel
    if err := doc.DataTo(&m); err != nil {
        return nil, err
    }
    availability, err := hour.NewAvailabilityFromString(m.Availability)
    if err != nil {
        return nil, err
    }
    return hour.UnmarshalHourFromDatabase(m.Hour, availability), nil
}
```

**Wrong:**
```go
func (r *FirestoreHourRepository) toHour(doc *firestore.DocumentSnapshot) (*hour.Hour, error) {
    // VIOLATION: using business constructor for DB reconstruction
    return hour.NewAvailableHour(m.Hour)  // This re-validates and may reject valid stored data
}
```
