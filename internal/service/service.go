package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	authInfra "bitmerchant/internal/auth/adapters"
	authservice "bitmerchant/internal/auth/service"
	dashboardservice "bitmerchant/internal/dashboard/service"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/qr"
	menuservice "bitmerchant/internal/menu/service"
	"bitmerchant/internal/notification"
	notifwebpush "bitmerchant/internal/notification/webpush"
	"bitmerchant/internal/ordering/domain/order"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	ordernotif "bitmerchant/internal/ordering/ports/notification"
	orderingservice "bitmerchant/internal/ordering/service"
	payAdapters "bitmerchant/internal/payment/adapters"
	placeservice "bitmerchant/internal/places/service"
	restaurantservice "bitmerchant/internal/restaurant/service"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/wiring"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	wmmiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
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
	eventBus, err := events.NewEventBusWithConfig(eventBusConfig(cfg))
	if err != nil {
		return Application{}, nil, fmt.Errorf("init event bus: %w", err)
	}

	var db *sql.DB
	var orderEventsRouter *message.Router
	cleanupResources := func() {
		if orderEventsRouter != nil {
			_ = orderEventsRouter.Close()
		}
		_ = eventBus.Close()
		if db != nil {
			_ = db.Close()
		}
	}

	photoStorage, err := wiring.InitPhotoStorage(cfg, logger)
	if err != nil {
		cleanupResources()
		return Application{}, nil, fmt.Errorf("init photo storage: %w", err)
	}

	db, err = wiring.ConnectDatabase(cfg, logger)
	if err != nil {
		cleanupResources()
		return Application{}, nil, fmt.Errorf("connect database: %w", err)
	}

	repos := wiring.NewMemoryRepositories()
	if db != nil {
		repos = wiring.NewPostgresRepositories(db)
	}

	var pushRepo notifwebpush.Repository = notifwebpush.NewMemoryRepository()
	if db != nil {
		pushRepo = notifwebpush.NewPostgresRepository(db)
	}

	wiring.SeedData(ctx, repos)

	qrService := qr.NewQRCodeService()
	_ = payAdapters.NewCashPaymentMethod()
	sseHandler := commonhttp.NewSSEHandler()

	placesSvc := placeservice.New(repos)
	orderingSvc := orderingservice.New(repos, eventBus, logger, cfg.VAPIDPublicKey)
	menuSvc := menuservice.New(repos, photoStorage, cfg, orderingSvc.CartService, placesSvc.RecordMenuVisit)
	restaurantSvc := restaurantservice.New(repos, cfg, qrService, menuSvc)
	dashboardSvc := dashboardservice.New(repos, restaurantSvc.ToggleRestaurantOpen, logger.Logger)

	sessionOpts := newSessionOptions(cfg)
	webauthnSvc, err := authInfra.NewWebAuthnService(cfg.RPID, "BitMerchant", []string{cfg.MerchantBaseURL})
	if err != nil {
		cleanupResources()
		return Application{}, nil, fmt.Errorf("init webauthn: %w", err)
	}

	authSvc := authservice.New(repos, webauthnSvc, logger.Logger, sessionOpts, restaurantSvc.CreateRestaurant)

	vapidCfg := notifwebpush.VAPIDConfig{
		PublicKey:  cfg.VAPIDPublicKey,
		PrivateKey: cfg.VAPIDPrivateKey,
		Subject:    cfg.VAPIDSubject,
	}
	orderEventsRouter, err = startOrderEventsRouter(ctx, cfg, eventBus, logger, sseHandler, repos.Order, pushRepo, vapidCfg)
	if err != nil {
		cleanupResources()
		return Application{}, nil, fmt.Errorf("init order events router: %w", err)
	}

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
			Push:           orderinghttp.NewPushHandler(pushRepo),
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
		cleanupResources()
	}

	return application, cleanup, nil
}

func eventBusConfig(cfg Config) events.Config {
	return events.Config{
		Backend:           cfg.EventBusBackend,
		NATSURL:           cfg.NATSURL,
		NATSAutoProvision: cfg.NATSAutoProvision,
		NATSAckWait:       cfg.NATSAckWait,
		NATSCloseTimeout:  cfg.NATSCloseTimeout,
		NATSSubscribers:   cfg.NATSSubscribersCount,
		NATSInstanceID:    cfg.NATSInstanceID,
	}
}

func resolveRouterCloseTimeout(cfg Config) time.Duration {
	if cfg.NATSCloseTimeout <= 0 {
		return 30 * time.Second
	}
	return cfg.NATSCloseTimeout
}

func newSessionOptions(cfg Config) middleware.SessionOptions {
	secureCookie := middleware.ShouldUseSecureCookies(cfg.PublicBaseURL, cfg.ForceSecureCookie) ||
		middleware.ShouldUseSecureCookies(cfg.CustomerBaseURL, cfg.ForceSecureCookie) ||
		middleware.ShouldUseSecureCookies(cfg.MerchantBaseURL, cfg.ForceSecureCookie)

	return middleware.SessionOptions{
		SecureCookie:       secureCookie,
		CookieName:         middleware.MerchantSessionCookieName,
		MerchantCookieName: middleware.MerchantSessionCookieName,
		CustomerCookieName: middleware.CustomerSessionCookieName,
		LegacyCookieName:   middleware.SessionCookieName,
	}
}

func startOrderEventsRouter(
	ctx context.Context,
	cfg Config,
	eventBus *events.EventBus,
	logger *logging.Logger,
	sseHandler *commonhttp.SSEHandler,
	orderRepo order.Repository,
	pushRepo notifwebpush.Repository,
	vapidCfg notifwebpush.VAPIDConfig,
) (*message.Router, error) {
	wmLogger := watermill.NewStdLogger(false, false)
	orderEventsRouter, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: resolveRouterCloseTimeout(cfg),
	}, wmLogger)
	if err != nil {
		return nil, err
	}

	orderEventsRouter.AddMiddleware(
		wmmiddleware.Recoverer,
		wmmiddleware.Retry{
			MaxRetries:      3,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
			Logger:          wmLogger,
		}.Middleware,
	)
	orderingservice.RegisterOrderSSEHandlers(orderEventsRouter, eventBus.Subscriber(), logger, sseHandler, orderRepo)

	webPushNotifier := notifwebpush.NewNotifier(pushRepo, vapidCfg)
	notifSvc := notification.NewService(logger, webPushNotifier)
	ordernotif.RegisterOrderNotificationHandlers(orderEventsRouter, eventBus.Subscriber(), logger, notifSvc)

	routerErrors := make(chan error, 1)
	go func() {
		routerErrors <- orderEventsRouter.Run(ctx)
	}()

	select {
	case runErr := <-routerErrors:
		_ = orderEventsRouter.Close()
		return nil, fmt.Errorf("run order events router: %w", runErr)
	case <-orderEventsRouter.Running():
		return orderEventsRouter, nil
	case <-ctx.Done():
		_ = orderEventsRouter.Close()
		return nil, fmt.Errorf("application context cancelled while starting order events router: %w", ctx.Err())
	}
}
