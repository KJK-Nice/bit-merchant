package service

import (
	"context"
	"encoding/json"
	"fmt"

	"bitmerchant/internal/app"
	authInfra "bitmerchant/internal/auth/adapters"
	"bitmerchant/internal/common"
	dashQuery "bitmerchant/internal/dashboard/app/query"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/qr"
	ifaceevents "bitmerchant/internal/interfaces/events"
	eventHandlers "bitmerchant/internal/interfaces/events/handlers"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"
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
)

// Config mirrors runtime configuration required by composition root.
type Config struct {
	PublicBaseURL          string
	CustomerBaseURL        string
	MerchantBaseURL        string
	RPID                   string
	ForceSecureCookie      bool
	DatabaseURL            string
	S3BucketName           string
	AWSRegion              string
	S3Endpoint             string
	S3UsePathStyle         bool
	S3PublicBaseURL        string
	S3PresignGetExpiresSec int
}

func NewApplication(ctx context.Context, cfg Config) (app.Application, func(), error) {
	logger := logging.NewLogger()
	return newApplication(ctx, cfg, logger)
}

func NewComponentTestApplication(ctx context.Context) app.Application {
	logger := logging.NewLogger()
	application, _, _ := newApplication(ctx, Config{}, logger)
	return application
}

func newApplication(ctx context.Context, cfg Config, logger *logging.Logger) (app.Application, func(), error) {
	eventBus := events.NewEventBus()

	photoStorage, err := initPhotoStorage(cfg, logger)
	if err != nil {
		_ = eventBus.Close()
		return app.Application{}, nil, fmt.Errorf("init photo storage: %w", err)
	}

	db, err := connectDatabase(cfg, logger)
	if err != nil {
		_ = eventBus.Close()
		return app.Application{}, nil, fmt.Errorf("connect database: %w", err)
	}

	repos := newMemoryRepositories()
	if db != nil {
		repos = newPostgresRepositories(db)
	}

	seedData(ctx, repos)

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
	generateQRUC := restQuery.NewGenerateRestaurantQRUseCase(qrService, cfg.CustomerBaseURL, repos.Restaurant)

	secureCookie := middleware.ShouldUseSecureCookies(cfg.PublicBaseURL, cfg.ForceSecureCookie) ||
		middleware.ShouldUseSecureCookies(cfg.CustomerBaseURL, cfg.ForceSecureCookie) ||
		middleware.ShouldUseSecureCookies(cfg.MerchantBaseURL, cfg.ForceSecureCookie)
	sessionOpts := middleware.SessionOptions{
		SecureCookie:       secureCookie,
		CookieName:         middleware.MerchantSessionCookieName,
		MerchantCookieName: middleware.MerchantSessionCookieName,
		CustomerCookieName: middleware.CustomerSessionCookieName,
		LegacyCookieName:   middleware.SessionCookieName,
	}
	webauthnSvc, err := authInfra.NewWebAuthnService(cfg.RPID, "BitMerchant", []string{cfg.MerchantBaseURL})
	if err != nil {
		if db != nil {
			_ = db.Close()
		}
		_ = eventBus.Close()
		return app.Application{}, nil, fmt.Errorf("init webauthn: %w", err)
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

	application := app.Application{
		Commands: app.Commands{
			CreateOrder:             createOrderUC,
			MarkOrderPaid:           markPaidUC,
			MarkOrderPreparing:      markPreparingUC,
			MarkOrderReady:          markReadyUC,
			CreateRestaurant:        createRestUC,
			ToggleRestaurantOpen:    toggleOpenUC,
			UpdateTableCount:        updateTableCountUC,
			CreateMenuCategory:      createCatUC,
			CreateMenuItem:          createItemUC,
			UpdateMenuItem:          updateMenuItemUC,
			UpdateMenuCategory:      updateMenuCategoryUC,
			ToggleMenuItemAvailable: toggleItemAvailUC,
			UploadMenuPhoto:         uploadPhotoUC,
			ReorderMenuCategories:   reorderCategoriesUC,
			ReorderMenuItems:        reorderItemsUC,
			RecordMenuVisit:         recordMenuVisitUC,
		},
		Queries: app.Queries{
			GetMenu:                getMenuUC,
			GetMenuForAdmin:        getMenuAdminUC,
			GetCustomerOrder:       getCustomerOrderByNumberUC,
			GetCustomerOrders:      getCustomerOrdersUC,
			GetKitchenOrders:       getKitchenOrdersUC,
			ListVisitedRestaurants: listVisitedUC,
			GenerateRestaurantQR:   generateQRUC,
		},
		Ports: app.Ports{
			Menu:           menuHandler,
			Cart:           cartHandler,
			Order:          orderHandler,
			Places:         placesHandler,
			Kitchen:        kitchenHandler,
			Admin:          adminHandler,
			Owner:          ownerHandler,
			Dashboard:      dashboardHandler,
			Auth:           authHandler,
			SSE:            sseHandler,
			MembershipRepo: repos.Membership,
			SessionRepo:    repos.Session,
			UserRepo:       repos.User,
			SessionOptions: sessionOpts,
		},
		Infra: app.Infra{
			Logger:   logger,
			EventBus: eventBus,
			DB:       db,
		},
	}

	cleanup := func() {
		_ = eventBus.Close()
		if db != nil {
			_ = db.Close()
		}
	}

	return application, cleanup, nil
}

func setupEventSubscriptions(eventBus *events.EventBus, logger *logging.Logger, sseHandler *handler.SSEHandler, orderRepo order.Repository) {
	orderCreatedHandler := eventHandlers.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := eventHandlers.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := eventHandlers.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := eventHandlers.NewOrderReadyHandler(logger, sseHandler, orderRepo)

	subscribe(eventBus, common.EventOrderCreated, logger, func(msg []byte) {
		var event ifaceevents.OrderCreated
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderCreatedHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPaid, logger, func(msg []byte) {
		var event ifaceevents.OrderPaid
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPaidHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPreparing, logger, func(msg []byte) {
		var event ifaceevents.OrderPreparing
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPreparingHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderReady, logger, func(msg []byte) {
		var event ifaceevents.OrderReady
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
