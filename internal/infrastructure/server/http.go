package server

import (
	"fmt"
	"time"

	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/http/middleware"

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
}

// RunHTTPServer applies shared middleware/transport setup and starts the server.
func RunHTTPServer(cfg HTTPConfig, logger *logging.Logger, register func(e *echo.Echo)) error {
	e := echo.New()

	e.Use(echoMiddleware.Recover())
	e.Use(middleware.SurfaceRoutingMiddleware(middleware.SurfaceConfig{
		PublicBaseURL:   cfg.PublicBaseURL,
		CustomerBaseURL: cfg.CustomerBaseURL,
		MerchantBaseURL: cfg.MerchantBaseURL,
	}))
	e.Use(middleware.PerformanceMiddleware(logger, 200*time.Millisecond))
	if !cfg.DisableRateLimit {
		e.Use(middleware.RateLimitMiddleware())
	}
	e.Use(middleware.CSRFMiddleware())

	e.Static("/static", "static")
	e.Static("/assets", "assets")
	e.File("/sw.js", "static/pwa/sw.js")

	register(e)

	logger.Info("Starting server on port " + cfg.Port)
	if err := e.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		return err
	}
	return nil
}
