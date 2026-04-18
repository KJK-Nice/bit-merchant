package logging

import (
	"context"
	"log/slog"
)

type logContextKey struct{}

// FromContext returns the logger stored in ctx, or slog.Default() if none.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(logContextKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// ToContext returns a new context with logger attached.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, logContextKey{}, logger)
}
