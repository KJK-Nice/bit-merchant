package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
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
	"bitmerchant/internal/infrastructure/migrations"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	postgresRepos "bitmerchant/internal/infrastructure/repositories/postgres"
	s3Storage "bitmerchant/internal/infrastructure/storage/s3"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	var (
		userRepo       domain.UserRepository       = memory.NewMemoryUserRepository()
		membershipRepo domain.MembershipRepository = memory.NewMemoryMembershipRepository()
		invitationRepo domain.InvitationRepository = memory.NewMemoryInvitationRepository()
		sessionRepo    domain.SessionRepository    = memory.NewMemorySessionRepository()
	)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Second)
		createdDB, bootstrapErr := migrations.EnsureDatabaseExists(bootstrapCtx, databaseURL)
		bootstrapCancel()
		if bootstrapErr != nil {
			logger.Error("Failed to ensure database exists", "error", bootstrapErr)
			os.Exit(1)
		}
		if createdDB {
			logger.Info("Created missing database from DATABASE_URL")
		} else {
			logger.Info("Database from DATABASE_URL already exists")
		}

		db, dbErr := sql.Open("pgx", databaseURL)
		if dbErr != nil {
			logger.Error("Failed to open database connection", "error", dbErr)
			os.Exit(1)
		}

		// Retry ping because postgres container may be starting up.
		const (
			maxPingAttempts = 15
			pingDelay       = 2 * time.Second
		)
		var pingErr error
		for attempt := 1; attempt <= maxPingAttempts; attempt++ {
			pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			pingErr = db.PingContext(pingCtx)
			cancel()
			if pingErr == nil {
				break
			}
			logger.Warn("Database ping failed, retrying", "attempt", attempt, "maxAttempts", maxPingAttempts, "error", pingErr)
			time.Sleep(pingDelay)
		}
		if pingErr != nil {
			logger.Error("Failed to ping database after retries", "error", pingErr)
			_ = db.Close()
			os.Exit(1)
		}
		if migrationErr := migrations.Up(context.Background(), db); migrationErr != nil {
			logger.Error("Failed to run goose migrations", "error", migrationErr)
			_ = db.Close()
			os.Exit(1)
		}
		logger.Info("Using PostgreSQL repositories for auth persistence")
		userRepo = postgresRepos.NewUserRepository(db)
		membershipRepo = postgresRepos.NewMembershipRepository(db)
		invitationRepo = postgresRepos.NewInvitationRepository(db)
		sessionRepo = postgresRepos.NewSessionRepository(db)
		defer db.Close()
	}

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
	getMenuUC := menu.NewGetMenuUseCase(menuCatRepo, menuItemRepo, restRepo)
	createOrderUC := order.NewCreateOrderUseCase(orderRepo, paymentRepo, restRepo, eventBus, paymentMethod, logger)
	getOrderUC := order.NewGetOrderByNumberUseCase(orderRepo)
	getCustomerOrdersUC := order.NewGetCustomerOrdersUseCase(orderRepo) // Added

	// Kitchen Use Cases
	getKitchenOrdersUC := kitchen.NewGetKitchenOrdersUseCase(orderRepo)
	markPaidUC := kitchen.NewMarkOrderPaidUseCase(orderRepo, eventBus)
	markPreparingUC := kitchen.NewMarkOrderPreparingUseCase(orderRepo, eventBus)
	markReadyUC := kitchen.NewMarkOrderReadyUseCase(orderRepo, eventBus)

	// Owner/Admin Use Cases
	createRestUC := restaurant.NewCreateRestaurantUseCase(restRepo)
	createCatUC := menu.NewCreateMenuCategoryUseCase(menuCatRepo)
	createItemUC := menu.NewCreateMenuItemUseCase(menuItemRepo)
	uploadPhotoUC := menu.NewUploadPhotoUseCase(menuItemRepo, photoStorage)

	// Dashboard/Analytics Use Cases
	getStatsUC := dashboard.NewGetDashboardStatsUseCase(orderRepo)
	getHistoryUC := dashboard.NewGetOrderHistoryUseCase(orderRepo)
	getTopItemsUC := dashboard.NewGetTopSellingItemsUseCase(orderRepo)
	toggleOpenUC := restaurant.NewToggleRestaurantOpenUseCase(restRepo)

	// For QR generation, we need base URL. For dev it's localhost.
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	generateQRUC := restaurant.NewGenerateRestaurantQRUseCase(qrService, baseURL)

	parsedBaseURL, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		logger.Error("Failed to parse BASE_URL", "error", parseErr)
		os.Exit(1)
	}
	rpID := parsedBaseURL.Hostname()
	if rpID == "" {
		rpID = "localhost"
	}
	forceSecureCookies := os.Getenv("COOKIE_SECURE") == "true"
	sessionOpts := middleware.SessionOptions{
		SecureCookie: middleware.ShouldUseSecureCookies(baseURL, forceSecureCookies),
	}
	webauthnSvc, err := authInfra.NewWebAuthnService(rpID, "BitMerchant", []string{baseURL})
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", "error", err)
		os.Exit(1)
	}

	// 3. Handlers
	menuHandler := handler.NewMenuHandler(getMenuUC, cartService)
	cartHandler := handler.NewCartHandler(cartService, menuItemRepo)
	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, getCustomerOrdersUC, cartService)
	kitchenHandler := handler.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC)
	adminHandler := handler.NewAdminHandler(createRestUC, createCatUC, createItemUC, getMenuUC, uploadPhotoUC, generateQRUC)
	ownerHandler := handler.NewOwnerHandler(createRestUC)
	dashboardHandler := handler.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, toggleOpenUC)
	authHandler := handler.NewAuthHandler(webauthnSvc, userRepo, membershipRepo, invitationRepo, sessionRepo, createRestUC, logger.Logger, sessionOpts)

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
	e.Use(middleware.SessionMiddlewareWithReposAndOptions(sessionRepo, userRepo, sessionOpts))
	e.Use(middleware.PerformanceMiddleware(logger, 200*time.Millisecond))
	e.Use(middleware.RateLimitMiddleware())
	e.Use(middleware.CSRFMiddleware())
	// e.Use(middleware.LoggingMiddleware())

	// Static files
	e.Static("/static", "static")
	e.Static("/assets", "assets")
	e.File("/sw.js", "static/pwa/sw.js")

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
	e.GET("/order/lookup", orderHandler.GetLookup)
	e.POST("/order/lookup", orderHandler.PostLookup)
	e.GET("/order/confirm", orderHandler.GetConfirmOrder)
	e.POST("/order/create", orderHandler.CreateOrder)
	e.GET("/order/:orderNumber", orderHandler.GetOrder)
	e.GET("/order/:orderNumber/stream", sseHandler.OrderStatusStream)

	// Kitchen
	kitchenGroup := e.Group("/kitchen")
	kitchenGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner, domain.RoleKitchenStaff))
	kitchenGroup.GET("", kitchenHandler.GetKitchen)
	kitchenGroup.GET("/stream", sseHandler.KitchenStream)
	kitchenGroup.POST("/order/:id/mark-paid", kitchenHandler.MarkPaid)
	kitchenGroup.POST("/order/:id/mark-preparing", kitchenHandler.MarkPreparing)
	kitchenGroup.POST("/order/:id/mark-ready", kitchenHandler.MarkReady)

	// Admin/Owner Menu Management
	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner))
	adminGroup.GET("/dashboard", adminHandler.Dashboard)
	adminGroup.POST("/category", adminHandler.CreateCategory)
	adminGroup.POST("/item", adminHandler.CreateItem)
	adminGroup.POST("/item/:id/photo", adminHandler.UploadPhoto)
	adminGroup.GET("/qr", adminHandler.GenerateQR)

	// Owner Signup
	e.GET("/owner/signup", ownerHandler.GetSignup)
	e.POST("/owner/signup", ownerHandler.PostSignup)

	// Auth
	e.GET("/auth/signup", authHandler.GetSignup)
	e.GET("/auth/login", authHandler.GetLogin)
	e.GET("/auth/invite/:token", authHandler.GetInvite)
	e.POST("/auth/register/begin", authHandler.BeginRegistration)
	e.POST("/auth/register/finish", authHandler.FinishRegistration)
	e.POST("/auth/login/begin", authHandler.BeginLogin)
	e.POST("/auth/login/finish", authHandler.FinishLogin)
	e.POST("/auth/logout", authHandler.Logout)

	// Dashboard Menu Management (US3)
	dashboardGroup := e.Group("/dashboard")
	dashboardGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner))
	dashboardGroup.GET("", dashboardHandler.Dashboard)
	dashboardGroup.GET("/menu", adminHandler.GetMenu)
	dashboardGroup.POST("/menu/category", adminHandler.CreateMenuCategory)
	dashboardGroup.POST("/menu/item", adminHandler.CreateMenuItem)
	dashboardGroup.POST("/menu/item/:id/photo", adminHandler.UploadMenuItemPhoto)
	dashboardGroup.GET("/qr-code", adminHandler.GetQRCode)
	dashboardGroup.POST("/toggle-open", dashboardHandler.ToggleOpen)
	dashboardGroup.POST("/invite", authHandler.CreateInvitation)

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
