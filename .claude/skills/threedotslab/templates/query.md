# Query Handler Scaffold Template

Generate a query handler file with a read model interface.

## Placeholders

- `{{Name}}` — PascalCase query name (e.g., `AvailableHours`)
- `{{name}}` — camelCase (e.g., `availableHours`)
- `{{name_snake}}` — snake_case (e.g., `available_hours`)
- `{{module}}` — Go module path from go.mod
- `{{Result}}` — Result type (e.g., `[]Date`, `*HourDetails`)

## File: `app/query/{{name_snake}}.go`

```go
package query

import (
	"context"

	"github.com/sirupsen/logrus"

	"{{module_common}}/decorator"
)

// Read model — defines what data the query needs
// Implemented by adapters (repository or dedicated read store)
type {{Name}}ReadModel interface {
	{{Name}}(ctx context.Context /* TODO: add query params */) ({{Result}}, error)
}

// 1. Query struct — noun phrase, plain data
type {{Name}} struct {
	// TODO: Add query parameters
	// Example:
	// From time.Time
	// To   time.Time
}

// Result types — optimized for reading, may differ from domain entities
// type Date struct {
// 	Date  time.Time
// 	Hours []Hour
// }

// 2. Exported handler type alias
type {{Name}}Handler decorator.QueryHandler[{{Name}}, {{Result}}]

// 3. Unexported concrete handler struct
type {{name}}Handler struct {
	readModel {{Name}}ReadModel
}

// 4. Constructor with nil-checks + decorator wrapping
func New{{Name}}Handler(
	readModel {{Name}}ReadModel,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) {{Name}}Handler {
	if readModel == nil {
		panic("nil readModel")
	}
	if logger == nil {
		panic("nil logger")
	}
	if metricsClient == nil {
		panic("nil metricsClient")
	}

	return decorator.ApplyQueryDecorators[{{Name}}, {{Result}}](
		{{name}}Handler{readModel: readModel},
		logger,
		metricsClient,
	)
}

// Handle — delegates to read model, may add input validation
func (h {{name}}Handler) Handle(ctx context.Context, q {{Name}}) ({{Result}}, error) {
	// TODO: Add input validation if needed
	// Example:
	// if q.From.After(q.To) {
	//     return nil, errors.NewIncorrectInputError("date-from-after-date-to", "date from is after date to")
	// }

	return h.readModel.{{Name}}(ctx /* TODO: pass query params */)
}
```

## Update `app/app.go`

Add to the `Queries` struct:

```go
type Queries struct {
	// ... existing handlers ...
	{{Name}} query.{{Name}}Handler
}
```

## Update `service/application.go`

Wire the handler. The read model is typically implemented by the same repository adapter or a dedicated read adapter:

```go
Queries: app.Queries{
	// ... existing handlers ...
	{{Name}}: query.New{{Name}}Handler(
		{{entity}}Repository,  // implements {{Name}}ReadModel
		logger,
		metricsClient,
	),
},
```

## Implement ReadModel on Adapter

Add the read model method to your repository adapter:

```go
// In adapters/
func (r *Memory{{Entity}}Repository) {{Name}}(ctx context.Context /* params */) ({{Result}}, error) {
	// TODO: Implement query against storage
}
```
