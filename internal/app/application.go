package app

import (
	"database/sql"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	placesCmd "bitmerchant/internal/places/app/command"
	placesQuery "bitmerchant/internal/places/app/query"
	restCmd "bitmerchant/internal/restaurant/app/command"
	restQuery "bitmerchant/internal/restaurant/app/query"
)

// Commands groups write use-cases for composition-root wiring.
type Commands struct {
	CreateOrder        *orderCmd.CreateOrderUseCase
	MarkOrderPaid      *orderCmd.MarkOrderPaidUseCase
	MarkOrderPreparing *orderCmd.MarkOrderPreparingUseCase
	MarkOrderReady     *orderCmd.MarkOrderReadyUseCase

	CreateRestaurant     *restCmd.CreateRestaurantUseCase
	ToggleRestaurantOpen *restCmd.ToggleRestaurantOpenUseCase
	UpdateTableCount     *restCmd.UpdateRestaurantTableCountUseCase

	CreateMenuCategory      *menuCmd.CreateMenuCategoryUseCase
	CreateMenuItem          *menuCmd.CreateMenuItemUseCase
	UpdateMenuItem          *menuCmd.UpdateMenuItemUseCase
	UpdateMenuCategory      *menuCmd.UpdateMenuCategoryUseCase
	ToggleMenuItemAvailable *menuCmd.ToggleMenuItemAvailabilityUseCase
	UploadMenuPhoto         *menuCmd.UploadPhotoUseCase
	ReorderMenuCategories   *menuCmd.ReorderMenuCategoriesUseCase
	ReorderMenuItems        *menuCmd.ReorderMenuItemsUseCase

	RecordMenuVisit *placesCmd.RecordMenuVisitUseCase
}

// Queries groups read use-cases for composition-root wiring.
type Queries struct {
	GetMenu                *menuQuery.GetMenuUseCase
	GetMenuForAdmin        *menuQuery.GetMenuForAdminUseCase
	GetCustomerOrder       *orderQuery.GetCustomerOrderByNumberUseCase
	GetCustomerOrders      *orderQuery.GetCustomerOrdersUseCase
	GetKitchenOrders       *orderQuery.GetKitchenOrdersUseCase
	ListVisitedRestaurants *placesQuery.ListVisitedRestaurantsUseCase
	GenerateRestaurantQR   *restQuery.GenerateRestaurantQRUseCase
}

// Ports groups inbound adapters used by HTTP routing.
type Ports struct {
	Menu      *handler.MenuHandler
	Cart      *handler.CartHandler
	Order     *handler.OrderHandler
	Places    *handler.PlacesHandler
	Kitchen   *handler.KitchenHandler
	Admin     *handler.AdminHandler
	Owner     *handler.OwnerHandler
	Dashboard *handler.DashboardHandler
	Auth      *handler.AuthHandler
	SSE       *handler.SSEHandler

	MembershipRepo membership.Repository
	SessionRepo    session.Repository
	UserRepo       user.Repository
	SessionOptions middleware.SessionOptions
}

// Infra groups infrastructure managed by composition root.
type Infra struct {
	Logger   *logging.Logger
	EventBus *events.EventBus
	DB       *sql.DB
}

// Application is the composed runtime application.
type Application struct {
	Commands Commands
	Queries  Queries
	Ports    Ports
	Infra    Infra
}
