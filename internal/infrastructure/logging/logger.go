package logging

import (
	"log/slog"
	"os"

	"github.com/ThreeDotsLabs/humanslog"
)

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger.
// Uses JSON output when APP_ENV=production, and humanslog pretty output otherwise.
func NewLogger() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if os.Getenv("LOG_LEVEL") == "debug" {
		opts.Level = slog.LevelDebug
	}

	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = humanslog.NewHandler(os.Stdout, &humanslog.Options{
			HandlerOptions:  opts,
			SortKeys:        true,
			NewLineAfterLog: false,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return &Logger{logger}
}
