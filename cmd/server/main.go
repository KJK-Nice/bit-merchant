package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/dashboard"
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	authInfra "bitmerchant/internal/infrastructure/auth"
	"bitmerchant/internal/infrastructure/events"
	eventHandlers "bitmerchant/internal/infrastructure/events/handlers"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/qr"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	logger := logging.NewLogger()
	cfg, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load server config", "error", err)
		os.Exit(1)
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	photoStorage, err := initPhotoStorage(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize S3 storage", "error", err)
		os.Exit(1)
	}

	db, err := connectDatabase(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer func() { _ = db.Close() }()
		logger.Info("Using PostgreSQL repositories for core and auth persistence")
	}

	repos := newMemoryRepositories()
	if db != nil {
		repos = newPostgresRepositories(db)
	}

	seedData(repos)

	qrService := qr.NewQRCodeService()

	cartService := cart.NewCartService()
	paymentMethod := cash.NewCashPaymentMethod()
	sseHandler := handler.NewSSEHandler()

	getMenuUC := menu.NewGetMenuUseCase(repos.MenuCategory, repos.MenuItem, repos.Restaurant)
	createOrderUC := order.NewCreateOrderUseCase(repos.Order, repos.Payment, repos.Restaurant, eventBus, paymentMethod, logger)
	getOrderUC := order.NewGetOrderByNumberUseCase(repos.Order)
	getCustomerOrdersUC := order.NewGetCustomerOrdersUseCase(repos.Order)

	getKitchenOrdersUC := kitchen.NewGetKitchenOrdersUseCase(repos.Order)
	markPaidUC := kitchen.NewMarkOrderPaidUseCase(repos.Order, eventBus)
	markPreparingUC := kitchen.NewMarkOrderPreparingUseCase(repos.Order, eventBus)
	markReadyUC := kitchen.NewMarkOrderReadyUseCase(repos.Order, eventBus)

	createRestUC := restaurant.NewCreateRestaurantUseCase(repos.Restaurant)
	createCatUC := menu.NewCreateMenuCategoryUseCase(repos.MenuCategory)
	createItemUC := menu.NewCreateMenuItemUseCase(repos.MenuItem)
	uploadPhotoUC := menu.NewUploadPhotoUseCase(repos.MenuItem, photoStorage)

	getStatsUC := dashboard.NewGetDashboardStatsUseCase(repos.Order)
	getHistoryUC := dashboard.NewGetOrderHistoryUseCase(repos.Order)
	getTopItemsUC := dashboard.NewGetTopSellingItemsUseCase(repos.Order)
	toggleOpenUC := restaurant.NewToggleRestaurantOpenUseCase(repos.Restaurant)
	generateQRUC := restaurant.NewGenerateRestaurantQRUseCase(qrService, cfg.BaseURL)

	sessionOpts := middleware.SessionOptions{
		SecureCookie: middleware.ShouldUseSecureCookies(cfg.BaseURL, cfg.ForceSecureCookie),
	}
	webauthnSvc, err := authInfra.NewWebAuthnService(cfg.RPID, "BitMerchant", []string{cfg.BaseURL})
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", "error", err)
		os.Exit(1)
	}

	menuHandler := handler.NewMenuHandler(getMenuUC, cartService)
	cartHandler := handler.NewCartHandler(cartService, repos.MenuItem)
	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, getCustomerOrdersUC, cartService)
	kitchenHandler := handler.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC)
	adminHandler := handler.NewAdminHandler(createRestUC, createCatUC, createItemUC, getMenuUC, uploadPhotoUC, generateQRUC)
	ownerHandler := handler.NewOwnerHandler(createRestUC)
	dashboardHandler := handler.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, toggleOpenUC)
	authHandler := handler.NewAuthHandler(webauthnSvc, repos.User, repos.Membership, repos.Invitation, repos.Session, repos.Restaurant, createRestUC, logger.Logger, sessionOpts)

	setupEventSubscriptions(eventBus, logger, sseHandler, repos.Order)

	e := echo.New()

	e.Use(echoMiddleware.Recover())
	e.Use(middleware.SessionMiddlewareWithReposAndOptions(repos.Session, repos.User, sessionOpts))
	e.Use(middleware.PerformanceMiddleware(logger, 200*time.Millisecond))
	e.Use(middleware.RateLimitMiddleware())
	e.Use(middleware.CSRFMiddleware())

	e.Static("/static", "static")
	e.Static("/assets", "assets")
	e.File("/sw.js", "static/pwa/sw.js")

	registerRoutes(e, routeHandlers{
		Menu:      menuHandler,
		Cart:      cartHandler,
		Order:     orderHandler,
		Kitchen:   kitchenHandler,
		Admin:     adminHandler,
		Owner:     ownerHandler,
		Dashboard: dashboardHandler,
		Auth:      authHandler,
		SSE:       sseHandler,
	}, repos.Membership)

	logger.Info("Starting server on port " + cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}

func setupEventSubscriptions(eventBus *events.EventBus, logger *logging.Logger, sseHandler *handler.SSEHandler, orderRepo domain.OrderRepository) {
	orderCreatedHandler := eventHandlers.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := eventHandlers.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := eventHandlers.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := eventHandlers.NewOrderReadyHandler(logger, sseHandler, orderRepo)

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
