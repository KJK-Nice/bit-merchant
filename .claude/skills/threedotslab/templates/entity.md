# Domain Entity Scaffold Template

Generate a domain entity with factory constructor, value objects, and errors.

## Placeholders

- `{{Name}}` — PascalCase entity name (e.g., `Training`, `Hour`, `Order`)
- `{{name}}` — camelCase (e.g., `training`)
- `{{name_lower}}` — all lowercase package name (e.g., `training`)
- `{{name_snake}}` — snake_case (e.g., `training`)

## File: `domain/{{name_lower}}/{{name_snake}}.go`

```go
package {{name_lower}}

import (
	"errors"
	"time"
)

// {{Name}} is the aggregate root for the {{name_lower}} domain.
type {{Name}} struct {
	uuid      string
	createdAt time.Time
	// TODO: Add domain fields (all private)
	// status    Status  // value object, not raw string
}

// New{{Name}} creates a new {{Name}} with validated invariants.
func New{{Name}}(uuid string) (*{{Name}}, error) {
	if uuid == "" {
		return nil, errors.New("empty {{name_lower}} uuid")
	}

	return &{{Name}}{
		uuid:      uuid,
		createdAt: time.Now(),
	}, nil
}

// Unmarshal{{Name}}FromDatabase reconstructs a {{Name}} from persistence.
// Bypasses validation — data was valid when stored.
func Unmarshal{{Name}}FromDatabase(
	uuid string,
	createdAt time.Time,
	// TODO: Add all persisted fields
) *{{Name}} {
	return &{{Name}}{
		uuid:      uuid,
		createdAt: createdAt,
	}
}

// Accessor methods — expose state without allowing mutation.

func (t {{Name}}) UUID() string {
	return t.uuid
}

func (t {{Name}}) CreatedAt() time.Time {
	return t.createdAt
}

// TODO: Add behavior methods using domain language.
// Examples:
//
// func (t *{{Name}}) Approve() error {
//     if t.status != Pending {
//         return ErrNotPending
//     }
//     t.status = Approved
//     return nil
// }
//
// func (t *{{Name}}) Cancel() error { ... }
// func (t *{{Name}}) Submit(details string) error { ... }
```

## File: `domain/{{name_lower}}/errors.go`

```go
package {{name_lower}}

import "errors"

// Sentinel errors — simple, no context needed.
var (
	ErrNotFound = errors.New("{{name_lower}} not found")
	// TODO: Add domain-specific errors
	// ErrAlreadyCanceled = errors.New("{{name_lower}} already canceled")
	// ErrNotPending      = errors.New("{{name_lower}} is not in pending state")
)

// Typed errors — carry context for logging/display.
// Example:
//
// type ForbiddenError struct {
//     RequestingUserUUID string
//     OwnerUUID          string
// }
//
// func (e ForbiddenError) Error() string {
//     return fmt.Sprintf("user %s cannot access {{name_lower}} owned by %s",
//         e.RequestingUserUUID, e.OwnerUUID)
// }
```

## File: `domain/{{name_lower}}/status.go` (Optional Value Object)

```go
package {{name_lower}}

import "fmt"

// Status is a value object — cannot be constructed with arbitrary values.
type Status struct {
	s string
}

var (
	Pending  = Status{"pending"}
	Approved = Status{"approved"}
	Canceled = Status{"canceled"}
)

func NewStatusFromString(s string) (Status, error) {
	switch s {
	case "pending":
		return Pending, nil
	case "approved":
		return Approved, nil
	case "canceled":
		return Canceled, nil
	default:
		return Status{}, fmt.Errorf("unknown {{name_lower}} status: %s", s)
	}
}

func (s Status) String() string {
	return s.s
}

func (s Status) IsZero() bool {
	return s == Status{}
}
```

## Post-Creation Checklist

- [ ] All struct fields are private (unexported)
- [ ] Factory constructor validates all invariants
- [ ] UnmarshalFromDatabase accepts all persisted fields
- [ ] Value objects are struct wrappers, not type aliases
- [ ] Behavior methods use domain language, not CRUD
- [ ] Errors are sentinel vars or typed structs
