# Command Handler Scaffold Template

Generate a single command handler file following the 4-component pattern.

## Placeholders

- `{{Name}}` — PascalCase command name (e.g., `ScheduleTraining`)
- `{{name}}` — camelCase (e.g., `scheduleTraining`)
- `{{module}}` — Go module path from go.mod
- `{{entity}}` — Domain entity name, lowercase (e.g., `hour`)
- `{{Entity}}` — Domain entity name, PascalCase (e.g., `Hour`)

## File: `app/command/{{name_snake}}.go`

```go
package command

import (
	"context"

	"github.com/sirupsen/logrus"

	"{{module}}/domain/{{entity}}"
	"{{module_common}}/decorator"
)

// 1. Command struct — imperative verb + noun, plain data
type {{Name}} struct {
	// TODO: Add command fields
	// Example:
	// UUID string
	// Hour time.Time
}

// 2. Exported handler type alias
type {{Name}}Handler decorator.CommandHandler[{{Name}}]

// 3. Unexported concrete handler struct
type {{name}}Handler struct {
	{{entity}}Repo {{entity}}.Repository
}

// 4. Constructor with nil-checks + decorator wrapping
func New{{Name}}Handler(
	{{entity}}Repo {{entity}}.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) {{Name}}Handler {
	if {{entity}}Repo == nil {
		panic("nil {{entity}}Repo")
	}
	if logger == nil {
		panic("nil logger")
	}
	if metricsClient == nil {
		panic("nil metricsClient")
	}

	return decorator.ApplyCommandDecorators[{{Name}}](
		{{name}}Handler{{"{"}}{{entity}}Repo: {{entity}}Repo},
		logger,
		metricsClient,
	)
}

// Handle — orchestrates domain logic, does NOT contain business rules
func (h {{name}}Handler) Handle(ctx context.Context, cmd {{Name}}) error {
	// TODO: Implement command handling
	//
	// Typical patterns:
	//
	// Pattern A — Update via callback:
	//   return h.{{entity}}Repo.Update{{Entity}}(ctx, cmd.UUID, func(e *{{entity}}.{{Entity}}) (*{{entity}}.{{Entity}}, error) {
	//       if err := e.SomeDomainAction(); err != nil {
	//           return nil, err
	//       }
	//       return e, nil
	//   })
	//
	// Pattern B — Create new entity:
	//   entity, err := {{entity}}.New{{Entity}}(cmd.UUID, ...)
	//   if err != nil {
	//       return err
	//   }
	//   return h.{{entity}}Repo.Save(ctx, entity)

	return nil
}
```

## Update `app/app.go`

After creating the handler, add it to the `Commands` struct:

```go
type Commands struct {
	// ... existing handlers ...
	{{Name}} command.{{Name}}Handler
}
```

## Update `service/application.go`

Wire the handler in the composition root:

```go
Commands: app.Commands{
	// ... existing handlers ...
	{{Name}}: command.New{{Name}}Handler(
		{{entity}}Repository,
		logger,
		metricsClient,
	),
},
```
