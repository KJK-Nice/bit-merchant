package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/application/payment"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"bitmerchant/internal/infrastructure/strike"
	apphttp "bitmerchant/internal/interfaces/http"
	appmiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize repositories
	restaurantRepo := memory.NewMemoryRestaurantRepository()
	menuCategoryRepo := memory.NewMemoryMenuCategoryRepository()
	menuItemRepo := memory.NewMemoryMenuItemRepository()
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()

	// Initialize Strike client
	strikeClient := strike.NewClient(
		os.Getenv("STRIKE_API_KEY"),
		os.Getenv("STRIKE_API_URL"),
	)

	// Initialize event bus
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Initialize cart store
	cartStore := cart.NewCartStore()

	// Initialize use cases
	getMenuUseCase := menu.NewGetMenuUseCase(restaurantRepo, menuCategoryRepo, menuItemRepo)
	addToCartUseCase := cart.NewAddToCartUseCase(cartStore, menuItemRepo, restaurantRepo)
	removeFromCartUseCase := cart.NewRemoveFromCartUseCase(cartStore)
	getCartUseCase := cart.NewGetCartUseCase(cartStore)
	createInvoiceUseCase := payment.NewCreatePaymentInvoiceUseCase(strikeClient, paymentRepo, restaurantRepo, menuItemRepo)
	checkStatusUseCase := payment.NewCheckPaymentStatusUseCase(strikeClient, paymentRepo)
	getOrderUseCase := order.NewGetOrderByNumberUseCase(orderRepo)
	_ = order.NewCreateOrderUseCase(orderRepo, paymentRepo, menuItemRepo, eventBus) // Will be used by event handlers

	// Initialize SSE hub
	sseHub := apphttp.NewSSEHub()

	// Initialize handlers
	menuHandler := apphttp.NewMenuHandler(getMenuUseCase)
	cartHandler := apphttp.NewCartHandler(addToCartUseCase, removeFromCartUseCase, getCartUseCase)
	paymentHandler := apphttp.NewPaymentHandler(createInvoiceUseCase, checkStatusUseCase)
	orderHandler := apphttp.NewOrderHandler(getOrderUseCase)
	sseHandler := apphttp.NewOrderSSEHandler(sseHub)

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(appmiddleware.LoggingMiddleware())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = appmiddleware.ErrorHandler

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "BitMerchant API")
	})

	// Menu routes
	e.GET("/menu", menuHandler.GetMenu)

	// Cart routes
	e.POST("/cart/add", cartHandler.AddToCart)
	e.POST("/cart/remove", cartHandler.RemoveFromCart)
	e.GET("/cart", cartHandler.GetCart)

	// Payment routes
	e.POST("/payment/create-invoice", paymentHandler.CreateInvoice)
	e.GET("/payment/status/:invoiceId", paymentHandler.CheckStatus)

	// Order routes
	e.GET("/order/:orderNumber", orderHandler.GetOrder)
	e.GET("/order/:orderNumber/stream", sseHandler.StreamOrderStatus)

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
