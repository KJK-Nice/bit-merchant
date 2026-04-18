package middleware

import (
	"bitmerchant/internal/infrastructure/logging"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid/v3"
)

const correlationIDHeader = "Correlation-ID"

// RequestIDMiddleware generates or propagates a Correlation-ID for each request,
// attaches an enriched logger to the request context, and echoes the ID back in
// the response header.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get(correlationIDHeader)
			if id == "" {
				id = shortuuid.New()
			}

			enriched := slog.Default().With("request_id", id)
			ctx := logging.ToContext(c.Request().Context(), enriched)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set(correlationIDHeader, id)

			return next(c)
		}
	}
}
