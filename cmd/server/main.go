package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	eventHandlers "bitmerchant/internal/infrastructure/events/handlers"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	s3Storage "bitmerchant/internal/infrastructure/storage/s3"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Infrastructure
	logger := logging.NewLogger()
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// S3 Storage
	bucketName := os.Getenv("S3_BUCKET_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	var photoStorage domain.PhotoStorage
	var err error

	if bucketName != "" && awsRegion != "" {
		photoStorage, err = s3Storage.NewS3Storage(context.Background(), bucketName, awsRegion)
		if err != nil {
			logger.Error("Failed to initialize S3 storage", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("S3 config missing, photo uploads will fail")
		// For dev/testing without S3, we could use a no-op or local storage.
		// For now, nil is fine, app will panic if upload attempted, or we can handle it.
	}

	// QR Service
	qrService := qr.NewQRCodeService()

	// Repositories
	restRepo := memory.NewMemoryRestaurantRepository()
	menuCatRepo := memory.NewMemoryMenuCategoryRepository()
	menuItemRepo := memory.NewMemoryMenuItemRepository()
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()

	// Services
	cartService := cart.NewCartService()
	paymentMethod := cash.NewCashPaymentMethod()
	sseHandler := handler.NewSSEHandler()

	// --- Seeding Data (MVP) ---
	// Restaurant
	restaurantID := domain.RestaurantID("restaurant_1") // Corrected ID to match tests/admin
	restaurantObj, _ := domain.NewRestaurant(restaurantID, "BitMerchant Cafe")
	_ = restRepo.Save(restaurantObj)

	// Categories
	cat1, _ := domain.NewMenuCategory("cat_1", restaurantID, "Appetizers", 1)
	cat2, _ := domain.NewMenuCategory("cat_2", restaurantID, "Mains", 2)
	cat3, _ := domain.NewMenuCategory("cat_3", restaurantID, "Drinks", 3)
	_ = menuCatRepo.Save(cat1)
	_ = menuCatRepo.Save(cat2)
	_ = menuCatRepo.Save(cat3)

	// Items
	item1, _ := domain.NewMenuItem("item_1", "cat_1", restaurantID, "Bruschetta", 8.50)
	_ = item1.SetDescription("Toasted bread with tomatoes and basil")
	_ = menuItemRepo.Save(item1)

	item2, _ := domain.NewMenuItem("item_2", "cat_2", restaurantID, "Bitcoin Burger", 15.00)
	_ = item2.SetDescription("Premium beef patty with cheese")
	_ = menuItemRepo.Save(item2)

	item3, _ := domain.NewMenuItem("item_3", "cat_3", restaurantID, "Satoshi Soda", 3.00)
	_ = menuItemRepo.Save(item3)
	// --------------------------

	// 2. Use Cases
	getMenuUC := menu.NewGetMenuUseCase(menuCatRepo, menuItemRepo)
	createOrderUC := order.NewCreateOrderUseCase(orderRepo, paymentRepo, eventBus, paymentMethod, logger)
	getOrderUC := order.NewGetOrderByNumberUseCase(orderRepo)

	// Kitchen Use Cases
	getKitchenOrdersUC := kitchen.NewGetKitchenOrdersUseCase(orderRepo)
	markPaidUC := kitchen.NewMarkOrderPaidUseCase(orderRepo, eventBus)
	markPreparingUC := kitchen.NewMarkOrderPreparingUseCase(orderRepo, eventBus)
	markReadyUC := kitchen.NewMarkOrderReadyUseCase(orderRepo, eventBus)

	// Owner Use Cases
	createRestUC := restaurant.NewCreateRestaurantUseCase(restRepo)
	createCatUC := menu.NewCreateMenuCategoryUseCase(menuCatRepo)
	createItemUC := menu.NewCreateMenuItemUseCase(menuItemRepo)
	uploadPhotoUC := menu.NewUploadPhotoUseCase(menuItemRepo, photoStorage)

	// For QR generation, we need base URL. For dev it's localhost.
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	generateQRUC := restaurant.NewGenerateRestaurantQRUseCase(qrService, baseURL)

	// 3. Handlers
	menuHandler := handler.NewMenuHandler(getMenuUC, cartService)
	cartHandler := handler.NewCartHandler(cartService, menuItemRepo)
	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, cartService)
	kitchenHandler := handler.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC)
	adminHandler := handler.NewAdminHandler(createRestUC, createCatUC, createItemUC, getMenuUC, uploadPhotoUC, generateQRUC)

	// 4. Event Handlers
	orderCreatedHandler := eventHandlers.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := eventHandlers.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := eventHandlers.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := eventHandlers.NewOrderReadyHandler(logger, sseHandler, orderRepo)

	// Subscribe
	subscribe(eventBus, "OrderCreated", logger, func(msg []byte) {
		var event domain.OrderCreated
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderCreatedHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, "OrderPaid", logger, func(msg []byte) {
		var event domain.OrderPaid
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPaidHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, "OrderPreparing", logger, func(msg []byte) {
		var event domain.OrderPreparing
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPreparingHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, "OrderReady", logger, func(msg []byte) {
		var event domain.OrderReady
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderReadyHandler.Handle(context.Background(), event)
		}
	})

	// 5. Server Setup
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.SessionMiddleware())
	// e.Use(middleware.LoggingMiddleware())

	// Static files
	e.Static("/static", "static")
	e.Static("/assets", "assets")

	// 6. Routes

	// Redirect root to menu
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/menu")
	})

	// Menu
	e.GET("/menu", menuHandler.GetMenu)

	// Cart
	e.GET("/cart", cartHandler.GetCart)
	e.POST("/cart/add", cartHandler.AddToCart)
	e.POST("/cart/remove", cartHandler.RemoveFromCart)

	// Order
	e.GET("/order/confirm", orderHandler.GetConfirmOrder)
	e.POST("/order/create", orderHandler.CreateOrder)
	e.GET("/order/:orderNumber", orderHandler.GetOrder)
	e.GET("/order/:orderNumber/stream", sseHandler.OrderStatusStream)

	// Kitchen
	e.GET("/kitchen", kitchenHandler.GetKitchen)
	e.GET("/kitchen/stream", sseHandler.KitchenStream)
	e.POST("/kitchen/order/:id/mark-paid", kitchenHandler.MarkPaid)
	e.POST("/kitchen/order/:id/mark-preparing", kitchenHandler.MarkPreparing)
	e.POST("/kitchen/order/:id/mark-ready", kitchenHandler.MarkReady)

	// Admin/Owner
	e.GET("/admin/dashboard", adminHandler.Dashboard)
	e.POST("/admin/category", adminHandler.CreateCategory)
	e.POST("/admin/item", adminHandler.CreateItem)
	e.POST("/admin/item/:id/photo", adminHandler.UploadPhoto)
	e.GET("/admin/qr", adminHandler.GenerateQR)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info("Starting server on port " + port)
	e.Logger.Fatal(e.Start(":" + port))
}

func subscribe(bus *events.EventBus, topic string, logger *logging.Logger, handlerFunc func([]byte)) {
	go func() {
		msgs, err := bus.Subscribe(context.Background(), topic)
		if err != nil {
			logger.Error("Failed to subscribe", "topic", topic, "error", err)
			return
		}
		for msg := range msgs {
			handlerFunc(msg.Payload)
			msg.Ack()
		}
	}()
}
