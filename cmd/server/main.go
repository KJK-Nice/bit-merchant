package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/repositories/memory"
	appmiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize repositories (will be used when handlers are added)
	_ = memory.NewMemoryRestaurantRepository()
	_ = memory.NewMemoryMenuCategoryRepository()
	_ = memory.NewMemoryMenuItemRepository()
	_ = memory.NewMemoryOrderRepository()
	_ = memory.NewMemoryPaymentRepository()

	// Initialize event bus
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(appmiddleware.LoggingMiddleware())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = appmiddleware.ErrorHandler

	// Static files
	e.Static("/static", "static")

	// Routes will be added here
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "BitMerchant API")
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	fmt.Println("Server stopped")
}
