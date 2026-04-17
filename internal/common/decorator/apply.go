package decorator

import (
	"context"
	"log/slog"
	"reflect"
	"time"
)

// ApplyCommandDecorators wraps a command handler with logging and metrics.
func ApplyCommandDecorators[C any](h CommandHandler[C], log *slog.Logger, metrics MetricsClient) CommandHandler[C] {
	if metrics == nil {
		metrics = NoopMetrics{}
	}
	return commandDecorated[C]{inner: h, log: log, metrics: metrics}
}

// ApplyQueryDecorators wraps a query handler with logging and metrics.
func ApplyQueryDecorators[Q any, R any](h QueryHandler[Q, R], log *slog.Logger, metrics MetricsClient) QueryHandler[Q, R] {
	if metrics == nil {
		metrics = NoopMetrics{}
	}
	return queryDecorated[Q, R]{inner: h, log: log, metrics: metrics}
}

func typeName(v any) string {
	if v == nil {
		return "unknown"
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

type commandDecorated[C any] struct {
	inner   CommandHandler[C]
	log     *slog.Logger
	metrics MetricsClient
}

func (d commandDecorated[C]) Handle(ctx context.Context, cmd C) error {
	name := typeName(cmd)
	start := time.Now()
	err := d.inner.Handle(ctx, cmd)
	if d.log != nil {
		d.log.DebugContext(ctx, "command", "name", name, "duration_ms", time.Since(start).Milliseconds(), "err", err)
	}
	d.metrics.IncCommand(name)
	return err
}

type queryDecorated[Q any, R any] struct {
	inner   QueryHandler[Q, R]
	log     *slog.Logger
	metrics MetricsClient
}

func (d queryDecorated[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	name := typeName(q)
	start := time.Now()
	r, err := d.inner.Handle(ctx, q)
	if d.log != nil {
		d.log.DebugContext(ctx, "query", "name", name, "duration_ms", time.Since(start).Milliseconds(), "err", err)
	}
	d.metrics.IncQuery(name)
	return r, err
}
