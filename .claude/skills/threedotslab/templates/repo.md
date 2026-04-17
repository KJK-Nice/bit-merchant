# Repository Scaffold Template

Generate a repository interface in the domain package and a memory implementation in adapters.

## Placeholders

- `{{Name}}` — PascalCase entity name (e.g., `Training`)
- `{{name}}` — camelCase (e.g., `training`)
- `{{name_lower}}` — all lowercase package name (e.g., `training`)
- `{{name_snake}}` — snake_case (e.g., `training`)
- `{{module}}` — Go module path from go.mod

## File: `domain/{{name_lower}}/repository.go`

```go
package {{name_lower}}

import "context"

// Repository defines persistence operations for {{Name}}.
// Defined in domain — adapters implement it implicitly.
type Repository interface {
	// Get{{Name}} retrieves a {{Name}} by its UUID.
	Get{{Name}}(ctx context.Context, uuid string) (*{{Name}}, error)

	// Update{{Name}} loads a {{Name}}, applies the update function within a
	// transaction, and persists the result. The callback pattern ensures
	// domain logic is separated from transaction management.
	Update{{Name}}(ctx context.Context, uuid string,
		updateFn func(t *{{Name}}) (*{{Name}}, error)) error

	// TODO: Add other methods as needed. Examples:
	// Save{{Name}}(ctx context.Context, t *{{Name}}) error
	// Delete{{Name}}(ctx context.Context, uuid string) error
}
```

## File: `adapters/memory_{{name_snake}}_repository.go`

```go
package adapters

import (
	"context"
	"sync"

	"{{module}}/domain/{{name_lower}}"
)

// Memory{{Name}}Repository is an in-memory implementation of {{name_lower}}.Repository.
// Useful for tests and local development.
type Memory{{Name}}Repository struct {
	{{name}}s map[string]{{name_lower}}.{{Name}}
	mu       sync.RWMutex
}

func NewMemory{{Name}}Repository() *Memory{{Name}}Repository {
	return &Memory{{Name}}Repository{
		{{name}}s: make(map[string]{{name_lower}}.{{Name}}),
	}
}

func (r *Memory{{Name}}Repository) Get{{Name}}(ctx context.Context, uuid string) (*{{name_lower}}.{{Name}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.{{name}}s[uuid]
	if !ok {
		return nil, {{name_lower}}.ErrNotFound
	}

	// Return a copy to prevent mutation of stored value
	return &t, nil
}

func (r *Memory{{Name}}Repository) Update{{Name}}(
	ctx context.Context,
	uuid string,
	updateFn func(t *{{name_lower}}.{{Name}}) (*{{name_lower}}.{{Name}}, error),
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.{{name}}s[uuid]
	if !ok {
		return {{name_lower}}.ErrNotFound
	}

	updated, err := updateFn(&current)
	if err != nil {
		return err
	}

	r.{{name}}s[uuid] = *updated
	return nil
}

// Save{{Name}} stores a new {{Name}}. Used for initial creation.
func (r *Memory{{Name}}Repository) Save{{Name}}(ctx context.Context, t *{{name_lower}}.{{Name}}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.{{name}}s[t.UUID()] = *t
	return nil
}
```

## File: `adapters/memory_{{name_snake}}_repository_test.go`

```go
package adapters_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"{{module}}/adapters"
	"{{module}}/domain/{{name_lower}}"
)

func TestMemory{{Name}}Repository_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := adapters.NewMemory{{Name}}Repository()

	// Setup: create and save a {{name_lower}}
	entity, err := {{name_lower}}.New{{Name}}("test-uuid")
	require.NoError(t, err)

	err = repo.Save{{Name}}(ctx, entity)
	require.NoError(t, err)

	// Test: retrieve it
	got, err := repo.Get{{Name}}(ctx, "test-uuid")
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", got.UUID())
}

func TestMemory{{Name}}Repository_GetNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := adapters.NewMemory{{Name}}Repository()

	_, err := repo.Get{{Name}}(ctx, "nonexistent")
	assert.ErrorIs(t, err, {{name_lower}}.ErrNotFound)
}

func TestMemory{{Name}}Repository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := adapters.NewMemory{{Name}}Repository()

	// Setup
	entity, err := {{name_lower}}.New{{Name}}("test-uuid")
	require.NoError(t, err)
	err = repo.Save{{Name}}(ctx, entity)
	require.NoError(t, err)

	// Test: update via callback
	err = repo.Update{{Name}}(ctx, "test-uuid", func(t *{{name_lower}}.{{Name}}) (*{{name_lower}}.{{Name}}, error) {
		// TODO: Apply domain action
		return t, nil
	})
	assert.NoError(t, err)
}
```

## Extending to Production Adapters

When adding a real database adapter (e.g., PostgreSQL):

### 1. Create DB model struct

```go
// adapters/postgres_{{name_snake}}_repository.go

type postgres{{Name}} struct {
	UUID      string    `db:"uuid"`
	CreatedAt time.Time `db:"created_at"`
	// ... map all persisted fields
}
```

### 2. Implement conversion methods

```go
func (r *Postgres{{Name}}Repository) to{{Name}}(m postgres{{Name}}) *{{name_lower}}.{{Name}} {
	return {{name_lower}}.Unmarshal{{Name}}FromDatabase(m.UUID, m.CreatedAt)
}
```

### 3. Run shared tests against all implementations

```go
type TestRepository struct {
	Name       string
	Repository {{name_lower}}.Repository
}

func createRepositories(t *testing.T) []TestRepository {
	return []TestRepository{
		{Name: "memory", Repository: adapters.NewMemory{{Name}}Repository()},
		{Name: "postgres", Repository: newPostgresRepository(t)},
	}
}
```
