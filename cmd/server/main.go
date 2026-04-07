package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"bitmerchant/internal/common"
	dashQuery "bitmerchant/internal/dashboard/app/query"
	"bitmerchant/internal/infrastructure/events"
	eventHandlers "bitmerchant/internal/infrastructure/events/handlers"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/qr"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	orderCart "bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	payAdapters "bitmerchant/internal/payment/adapters"
	placesCmd "bitmerchant/internal/places/app/command"
	placesQuery "bitmerchant/internal/places/app/query"
	restCmd "bitmerchant/internal/restaurant/app/command"
	restQuery "bitmerchant/internal/restaurant/app/query"

	authInfra "bitmerchant/internal/auth/adapters"
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

	cartService := orderCart.NewCartService()
	_ = payAdapters.NewCashPaymentMethod()
	sseHandler := handler.NewSSEHandler()

	getMenuUC := menuQuery.NewGetMenuUseCase(repos.MenuCategory, repos.MenuItem, repos.Restaurant, photoStorage, menuQuery.PhotoSignerConfig{
		Bucket:        cfg.S3BucketName,
		Endpoint:      cfg.S3Endpoint,
		PublicBaseURL: cfg.S3PublicBaseURL,
	})
	getMenuAdminUC := menuQuery.NewGetMenuForAdminUseCase(repos.MenuCategory, repos.MenuItem, repos.Restaurant, photoStorage, menuQuery.PhotoSignerConfig{
		Bucket:        cfg.S3BucketName,
		Endpoint:      cfg.S3Endpoint,
		PublicBaseURL: cfg.S3PublicBaseURL,
	})
	updateMenuItemUC := menuCmd.NewUpdateMenuItemUseCase(repos.MenuItem, repos.MenuCategory)
	updateMenuCategoryUC := menuCmd.NewUpdateMenuCategoryUseCase(repos.MenuCategory)
	toggleItemAvailUC := menuCmd.NewToggleMenuItemAvailabilityUseCase(repos.MenuItem)
	createOrderUC := orderCmd.NewCreateOrderUseCase(repos.Order, repos.Restaurant, eventBus, logger)
	getCustomerOrderByNumberUC := orderQuery.NewGetCustomerOrderByNumberUseCase(repos.Order)
	getCustomerOrdersUC := orderQuery.NewGetCustomerOrdersUseCase(repos.Order)
	recordMenuVisitUC := placesCmd.NewRecordMenuVisitUseCase(repos.Restaurant, repos.SessionRestaurantVisits)
	listVisitedUC := placesQuery.NewListVisitedRestaurantsUseCase(repos.SessionRestaurantVisits, repos.Restaurant, repos.Order)

	getKitchenOrdersUC := orderQuery.NewGetKitchenOrdersUseCase(repos.Order)
	markPaidUC := orderCmd.NewMarkOrderPaidUseCase(repos.Order, eventBus)
	markPreparingUC := orderCmd.NewMarkOrderPreparingUseCase(repos.Order, eventBus)
	markReadyUC := orderCmd.NewMarkOrderReadyUseCase(repos.Order, eventBus)

	createRestUC := restCmd.NewCreateRestaurantUseCase(repos.Restaurant)
	createCatUC := menuCmd.NewCreateMenuCategoryUseCase(repos.MenuCategory)
	createItemUC := menuCmd.NewCreateMenuItemUseCase(repos.MenuItem)
	uploadPhotoUC := menuCmd.NewUploadPhotoUseCase(repos.MenuItem, photoStorage)
	reorderCategoriesUC := menuCmd.NewReorderMenuCategoriesUseCase(repos.MenuCategory)
	reorderItemsUC := menuCmd.NewReorderMenuItemsUseCase(repos.MenuItem, repos.MenuCategory)

	getStatsUC := dashQuery.NewGetDashboardStatsUseCase(repos.Order)
	getHistoryUC := dashQuery.NewGetOrderHistoryUseCase(repos.Order)
	getTopItemsUC := dashQuery.NewGetTopSellingItemsUseCase(repos.Order)
	toggleOpenUC := restCmd.NewToggleRestaurantOpenUseCase(repos.Restaurant)
	updateTableCountUC := restCmd.NewUpdateRestaurantTableCountUseCase(repos.Restaurant)
	generateQRUC := restQuery.NewGenerateRestaurantQRUseCase(qrService, cfg.BaseURL, repos.Restaurant)

	sessionOpts := middleware.SessionOptions{
		SecureCookie: middleware.ShouldUseSecureCookies(cfg.BaseURL, cfg.ForceSecureCookie),
	}
	webauthnSvc, err := authInfra.NewWebAuthnService(cfg.RPID, "BitMerchant", []string{cfg.BaseURL})
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", "error", err)
		os.Exit(1)
	}

	menuHandler := handler.NewMenuHandler(getMenuUC, cartService, recordMenuVisitUC)
	cartHandler := handler.NewCartHandler(cartService, repos.MenuItem)
	orderHandler := handler.NewOrderHandler(createOrderUC, getCustomerOrderByNumberUC, getCustomerOrdersUC, cartService)
	placesHandler := handler.NewPlacesHandler(listVisitedUC)
	kitchenHandler := handler.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC, repos.Restaurant, repos.Membership)
	adminHandler := handler.NewAdminHandler(createRestUC, createCatUC, createItemUC, getMenuAdminUC, updateMenuItemUC, updateMenuCategoryUC, toggleItemAvailUC, uploadPhotoUC, reorderCategoriesUC, reorderItemsUC, repos.MenuItem, updateTableCountUC, generateQRUC, repos.Membership, repos.Restaurant)
	ownerHandler := handler.NewOwnerHandler(createRestUC)
	dashboardHandler := handler.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, toggleOpenUC, repos.Restaurant, repos.Membership, logger.Logger)
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
		Places:    placesHandler,
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

func setupEventSubscriptions(eventBus *events.EventBus, logger *logging.Logger, sseHandler *handler.SSEHandler, orderRepo order.Repository) {
	orderCreatedHandler := eventHandlers.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := eventHandlers.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := eventHandlers.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := eventHandlers.NewOrderReadyHandler(logger, sseHandler, orderRepo)

	subscribe(eventBus, common.EventOrderCreated, logger, func(msg []byte) {
		var event order.OrderCreated
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderCreatedHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPaid, logger, func(msg []byte) {
		var event order.OrderPaid
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPaidHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPreparing, logger, func(msg []byte) {
		var event order.OrderPreparing
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPreparingHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderReady, logger, func(msg []byte) {
		var event order.OrderReady
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
