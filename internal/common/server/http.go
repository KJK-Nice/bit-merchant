// Package server is the shared HTTP transport for cmd/server: Echo bootstrap, global middleware,
// static assets, graceful shutdown. Use [Component] from the composition root or call [RunHTTPServer] directly.
package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/logging"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// HTTPConfig defines transport-level server configuration.
type HTTPConfig struct {
	Port             string
	PublicBaseURL    string
	CustomerBaseURL  string
	MerchantBaseURL  string
	DisableRateLimit bool
	// S3Endpoint is the S3-compatible storage endpoint (AWS_ENDPOINT_URL /
	// S3_ENDPOINT). Its origin (scheme+host) is added to the CSP img-src
	// allowlist so presigned image URLs served from that host are not blocked.
	// Leave empty when object storage is not configured.
	S3Endpoint string
}

// Component is the HTTP transport adapter used by the application composition root (cmd/server).
type Component struct {
	Config HTTPConfig
}

// Run applies shared middleware, registers routes, and blocks until ctx is cancelled or Listen fails.
func (c Component) Run(ctx context.Context, logger *logging.Logger, register func(e *echo.Echo)) error {
	return RunHTTPServer(ctx, c.Config, logger, register)
}

// RunHTTPServer applies shared middleware, registers routes, and runs until ctx is cancelled.
func RunHTTPServer(ctx context.Context, cfg HTTPConfig, logger *logging.Logger, register func(e *echo.Echo)) error {
	e := echo.New()

	// Skip the request timeout for SSE stream endpoints — they are long-lived by design.
	e.Use(echoMiddleware.ContextTimeoutWithConfig(echoMiddleware.ContextTimeoutConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasSuffix(c.Request().URL.Path, "/stream")
		},
		Timeout: 30 * time.Second,
	}))
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.RequestIDMiddleware())
	e.Use(middleware.SurfaceRoutingMiddleware(middleware.SurfaceConfig{
		PublicBaseURL:   cfg.PublicBaseURL,
		CustomerBaseURL: cfg.CustomerBaseURL,
		MerchantBaseURL: cfg.MerchantBaseURL,
	}))
	e.Use(middleware.LoggingMiddleware())
	e.Use(middleware.PerformanceMiddleware(200 * time.Millisecond))
	if !cfg.DisableRateLimit {
		e.Use(middleware.RateLimitMiddleware())
	}
	e.Use(middleware.CSRFMiddleware())
	e.Use(middleware.CSPMiddleware(cfg.S3Endpoint))

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.Static("/static", "static")
	e.Static("/assets", "assets")
	e.GET("/sw.js", serveSW)
	e.GET("/offline", serveOffline)
	e.GET("/sw-kill.js", serveKillSwitch)
	e.POST("/api/pwa/events", servePWAEvents)

	register(e)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("Starting server on port " + cfg.Port)
		if err := e.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := e.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}
}
