package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			c.Logger().Infof(
				"method=%s path=%s status=%d duration=%v",
				c.Request().Method,
				c.Request().URL.Path,
				c.Response().Status,
				duration,
			)

			return err
		}
	}
}
