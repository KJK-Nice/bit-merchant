package decorator

import (
	"context"
	"log/slog"
	"time"
)

// CommandResultHandler handles a command that returns a result (e.g. newly created aggregate).
// Prefer CommandHandler when no result is needed.
type CommandResultHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// ApplyCommandResultDecorators wraps a command+result handler with logging and metrics.
func ApplyCommandResultDecorators[C any, R any](h CommandResultHandler[C, R], log *slog.Logger, metrics MetricsClient) CommandResultHandler[C, R] {
	if metrics == nil {
		metrics = NoopMetrics{}
	}
	return commandResultDecorated[C, R]{inner: h, log: log, metrics: metrics}
}

type commandResultDecorated[C any, R any] struct {
	inner   CommandResultHandler[C, R]
	log     *slog.Logger
	metrics MetricsClient
}

func (d commandResultDecorated[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	name := typeName(cmd)
	start := time.Now()
	r, err := d.inner.Handle(ctx, cmd)
	if d.log != nil {
		d.log.DebugContext(ctx, "command", "name", name, "duration_ms", time.Since(start).Milliseconds(), "err", err)
	}
	d.metrics.IncCommand(name)
	return r, err
}
