package logging

import (
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger
func NewLogger() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	// Check env for debug level
	if os.Getenv("LOG_LEVEL") == "debug" {
		opts.Level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	return &Logger{logger}
}

