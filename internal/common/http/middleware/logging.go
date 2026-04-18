package middleware

import (
	"time"

	"bitmerchant/internal/infrastructure/logging"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware logs each HTTP request as a structured slog entry.
// Picks up the request_id-enriched logger set by RequestIDMiddleware.
func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			logging.FromContext(c.Request().Context()).Info("request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration_ms", duration.Milliseconds(),
			)

			return err
		}
	}
}
