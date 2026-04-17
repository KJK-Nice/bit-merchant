# Service Scaffold Template

Generate a complete service skeleton with all standard directories and stub files.

## Placeholders

- `{{Name}}` — PascalCase service/aggregate name (e.g., `Training`)
- `{{name}}` — camelCase (e.g., `training`)
- `{{name_snake}}` — snake_case (e.g., `training`)
- `{{name_lower}}` — all lowercase (e.g., `training`)
- `{{module}}` — Go module path from go.mod

## Files to Create

### 1. `domain/{{name_lower}}/{{name_snake}}.go`

```go
package {{name_lower}}

import (
	"errors"
	"time"
)

type {{Name}} struct {
	uuid      string
	createdAt time.Time
}

func New{{Name}}(uuid string) (*{{Name}}, error) {
	if uuid == "" {
		return nil, errors.New("empty {{name_lower}} uuid")
	}

	return &{{Name}}{
		uuid:      uuid,
		createdAt: time.Now(),
	}, nil
}

func Unmarshal{{Name}}FromDatabase(uuid string, createdAt time.Time) *{{Name}} {
	return &{{Name}}{
		uuid:      uuid,
		createdAt: createdAt,
	}
}

func (t {{Name}}) UUID() string {
	return t.uuid
}

func (t {{Name}}) CreatedAt() time.Time {
	return t.createdAt
}
```

### 2. `domain/{{name_lower}}/repository.go`

```go
package {{name_lower}}

import "context"

type Repository interface {
	Get{{Name}}(ctx context.Context, uuid string) (*{{Name}}, error)
	Update{{Name}}(ctx context.Context, uuid string,
		updateFn func(t *{{Name}}) (*{{Name}}, error)) error
}
```

### 3. `domain/{{name_lower}}/errors.go`

```go
package {{name_lower}}

import "errors"

var (
	ErrNotFound = errors.New("{{name_lower}} not found")
)
```

### 4. `app/app.go`

```go
package app

import (
	"{{module}}/app/command"
	"{{module}}/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	// Add command handlers here, e.g.:
	// Create{{Name}} command.Create{{Name}}Handler
}

type Queries struct {
	// Add query handlers here, e.g.:
	// {{Name}}ByUUID query.{{Name}}ByUUIDHandler
}
```

### 5. `app/command/.gitkeep`

Create empty directory placeholder.

### 6. `app/query/.gitkeep`

Create empty directory placeholder.

### 7. `ports/http.go`

```go
package ports

import (
	"{{module}}/app"
)

type HttpServer struct {
	app app.Application
}

func NewHttpServer(application app.Application) HttpServer {
	return HttpServer{app: application}
}
```

### 8. `main.go`

```go
package main

import (
	"context"
	"net/http"

	"{{module_common}}/logs"
	"{{module_common}}/server"
	"{{module}}/ports"
	"{{module}}/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	logs.Init()
	ctx := context.Background()

	app := service.NewApplication(ctx)

	server.New(
		server.WithHTTPHandler("api", func(router chi.Router) http.Handler {
			return ports.HandlerFromMux(ports.NewHttpServer(app), router)
		}),
		server.OnShutdown(
			server.Stop("api"),
		),
	).Run(ctx)
}
```

### 9. `adapters/memory_{{name_snake}}_repository.go`

```go
package adapters

import (
	"context"
	"sync"

	"{{module}}/domain/{{name_lower}}"
)

type Memory{{Name}}Repository struct {
	{{name_lower}}s map[string]{{name_lower}}.{{Name}}
	mu             sync.RWMutex
}

func NewMemory{{Name}}Repository() *Memory{{Name}}Repository {
	return &Memory{{Name}}Repository{
		{{name_lower}}s: make(map[string]{{name_lower}}.{{Name}}),
	}
}

func (r *Memory{{Name}}Repository) Get{{Name}}(ctx context.Context, uuid string) (*{{name_lower}}.{{Name}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.{{name_lower}}s[uuid]
	if !ok {
		return nil, {{name_lower}}.ErrNotFound
	}

	return &t, nil
}

func (r *Memory{{Name}}Repository) Update{{Name}}(
	ctx context.Context,
	uuid string,
	updateFn func(t *{{name_lower}}.{{Name}}) (*{{name_lower}}.{{Name}}, error),
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.{{name_lower}}s[uuid]
	if !ok {
		return {{name_lower}}.ErrNotFound
	}

	updated, err := updateFn(&current)
	if err != nil {
		return err
	}

	r.{{name_lower}}s[uuid] = *updated
	return nil
}
```

### 10. `service/application.go`

```go
package service

import (
	"context"

	"{{module}}/adapters"
	"{{module}}/app"
)

func NewApplication(ctx context.Context) app.Application {
	{{name_lower}}Repository := adapters.NewMemory{{Name}}Repository()
	_ = {{name_lower}}Repository // wire into handlers

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
```

## Post-Creation Instructions

After creating the service skeleton:

1. Ensure unified server exists: `/threedots scaffold unified_server`
2. Add your first command with `/threedots scaffold command <ActionName>`
3. Add your first query with `/threedots scaffold query <QueryName>`
4. Wire them in `service/application.go`
5. Add HTTP/gRPC handlers in `ports/`
6. When adding Watermill: `/threedots scaffold watermill_router` then `/threedots scaffold event_handler <Name>`
