package middleware

import (
	"time"

	"bitmerchant/internal/infrastructure/logging"

	"github.com/labstack/echo/v4"
)

// PerformanceMiddleware logs requests that take longer than threshold
func PerformanceMiddleware(logger *logging.Logger, threshold time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			if duration > threshold {
				logger.Warn("Slow request detected",
					"method", c.Request().Method,
					"path", c.Path(),
					"duration_ms", duration.Milliseconds(),
				)
			}
			return err
		}
	}
}

