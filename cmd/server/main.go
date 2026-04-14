package main

import (
	"context"
	"os"

	"bitmerchant/internal/infrastructure/server"
	"bitmerchant/internal/interfaces/http/middleware"
	"bitmerchant/internal/service"

	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to load server config: " + err.Error() + "\n")
		os.Exit(1)
	}

	application, cleanup, err := service.NewApplication(context.Background(), service.Config{
		PublicBaseURL:          cfg.PublicBaseURL,
		CustomerBaseURL:        cfg.CustomerBaseURL,
		MerchantBaseURL:        cfg.MerchantBaseURL,
		RPID:                   cfg.RPID,
		ForceSecureCookie:      cfg.ForceSecureCookie,
		DatabaseURL:            cfg.DatabaseURL,
		S3BucketName:           cfg.S3BucketName,
		AWSRegion:              cfg.AWSRegion,
		S3Endpoint:             cfg.S3Endpoint,
		S3UsePathStyle:         cfg.S3UsePathStyle,
		S3PublicBaseURL:        cfg.S3PublicBaseURL,
		S3PresignGetExpiresSec: cfg.S3PresignGetExpiresSec,
	})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to initialize application: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer cleanup()

	err = server.RunHTTPServer(server.HTTPConfig{
		Port:             cfg.Port,
		PublicBaseURL:    cfg.PublicBaseURL,
		CustomerBaseURL:  cfg.CustomerBaseURL,
		MerchantBaseURL:  cfg.MerchantBaseURL,
		DisableRateLimit: cfg.DisableRateLimit,
	}, application.Infra.Logger, func(e *echo.Echo) {
		e.Use(middleware.SessionMiddlewareWithReposAndOptions(application.Ports.SessionRepo, application.Ports.UserRepo, application.Ports.SessionOptions))

		registerRoutes(e, routeHandlers{
			Menu:      application.Ports.Menu,
			Cart:      application.Ports.Cart,
			Order:     application.Ports.Order,
			Places:    application.Ports.Places,
			Kitchen:   application.Ports.Kitchen,
			Admin:     application.Ports.Admin,
			Owner:     application.Ports.Owner,
			Dashboard: application.Ports.Dashboard,
			Auth:      application.Ports.Auth,
			SSE:       application.Ports.SSE,
		}, application.Ports.MembershipRepo)
	})
	if err != nil {
		application.Infra.Logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
