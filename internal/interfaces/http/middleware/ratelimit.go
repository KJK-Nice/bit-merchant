package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RateLimitMiddleware returns rate limiting middleware
func RateLimitMiddleware() echo.MiddlewareFunc {
	// 20 requests per second
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
}

