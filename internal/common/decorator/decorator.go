package decorator

import "context"

// CommandHandler handles a single command type (write use-case).
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

// QueryHandler handles a single query type (read use-case).
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}
