package service

import (
	"context"
	"fmt"

	authInfra "bitmerchant/internal/auth/adapters"
	authservice "bitmerchant/internal/auth/service"
	dashboardservice "bitmerchant/internal/dashboard/service"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/qr"
	menuservice "bitmerchant/internal/menu/service"
	orderingservice "bitmerchant/internal/ordering/service"
	payAdapters "bitmerchant/internal/payment/adapters"
	placeservice "bitmerchant/internal/places/service"
	restaurantservice "bitmerchant/internal/restaurant/service"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/wiring"
)

// Config mirrors runtime configuration required by the composition root (alias for wiring.Config).
type Config = wiring.Config

func NewApplication(ctx context.Context, cfg Config) (Application, func(), error) {
	logger := logging.NewLogger()
	return newApplication(ctx, cfg, logger)
}

func NewComponentTestApplication(ctx context.Context) Application {
	logger := logging.NewLogger()
	application, _, _ := newApplication(ctx, Config{}, logger)
	return application
}

func newApplication(ctx context.Context, cfg Config, logger *logging.Logger) (Application, func(), error) {
	eventBus := events.NewEventBus()

	photoStorage, err := wiring.InitPhotoStorage(cfg, logger)
	if err != nil {
		_ = eventBus.Close()
		return Application{}, nil, fmt.Errorf("init photo storage: %w", err)
	}

	db, err := wiring.ConnectDatabase(cfg, logger)
	if err != nil {
		_ = eventBus.Close()
		return Application{}, nil, fmt.Errorf("connect database: %w", err)
	}

	repos := wiring.NewMemoryRepositories()
	if db != nil {
		repos = wiring.NewPostgresRepositories(db)
	}

	wiring.SeedData(ctx, repos)

	qrService := qr.NewQRCodeService()
	_ = payAdapters.NewCashPaymentMethod()
	sseHandler := commonhttp.NewSSEHandler()

	placesSvc := placeservice.New(repos)
	orderingSvc := orderingservice.New(repos, eventBus, logger)
	menuSvc := menuservice.New(repos, photoStorage, cfg, orderingSvc.CartService, placesSvc.RecordMenuVisit)
	restaurantSvc := restaurantservice.New(repos, cfg, qrService, menuSvc)
	dashboardSvc := dashboardservice.New(repos, restaurantSvc.ToggleRestaurantOpen, logger.Logger)

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
		return Application{}, nil, fmt.Errorf("init webauthn: %w", err)
	}

	authSvc := authservice.New(repos, webauthnSvc, logger.Logger, sessionOpts, restaurantSvc.CreateRestaurant)

	orderingservice.RegisterOrderSSESubscriptions(eventBus, logger, sseHandler, repos.Order)

	application := Application{
		Commands: Commands{
			CreateOrder:             orderingSvc.CreateOrder,
			MarkOrderPaid:           orderingSvc.MarkOrderPaid,
			MarkOrderPreparing:      orderingSvc.MarkOrderPreparing,
			MarkOrderReady:          orderingSvc.MarkOrderReady,
			MarkOrderCompleted:      orderingSvc.MarkOrderCompleted,
			CreateRestaurant:        restaurantSvc.CreateRestaurant,
			ToggleRestaurantOpen:    restaurantSvc.ToggleRestaurantOpen,
			UpdateTableCount:        restaurantSvc.UpdateTableCount,
			CreateMenuCategory:      menuSvc.CreateMenuCategory,
			CreateMenuItem:          menuSvc.CreateMenuItem,
			UpdateMenuItem:          menuSvc.UpdateMenuItem,
			UpdateMenuCategory:      menuSvc.UpdateMenuCategory,
			ToggleMenuItemAvailable: menuSvc.ToggleItemAvailability,
			UploadMenuPhoto:         menuSvc.UploadMenuPhoto,
			ReorderMenuCategories:   menuSvc.ReorderMenuCategories,
			ReorderMenuItems:        menuSvc.ReorderMenuItems,
			RecordMenuVisit:         placesSvc.RecordMenuVisit,
		},
		Queries: Queries{
			GetMenu:                menuSvc.GetMenu,
			GetMenuForAdmin:        menuSvc.GetMenuForAdmin,
			GetCustomerOrder:       orderingSvc.GetCustomerOrder,
			GetCustomerOrders:      orderingSvc.GetCustomerOrders,
			GetKitchenOrders:       orderingSvc.GetKitchenOrders,
			ListVisitedRestaurants: placesSvc.ListVisitedRestaurants,
			GenerateRestaurantQR:   restaurantSvc.GenerateRestaurantQR,
		},
		Ports: Ports{
			Menu:           menuSvc.HTTP,
			Cart:           orderingSvc.CartHandler,
			Order:          orderingSvc.OrderHandler,
			Places:         placesSvc.HTTP,
			Kitchen:        orderingSvc.KitchenHandler,
			Admin:          restaurantSvc.Admin,
			Owner:          restaurantSvc.Owner,
			Dashboard:      dashboardSvc.HTTP,
			Auth:           authSvc.HTTP,
			SSE:            sseHandler,
			MembershipRepo: repos.Membership,
			SessionRepo:    repos.Session,
			UserRepo:       repos.User,
			SessionOptions: sessionOpts,
		},
		Infra: Infra{
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
