package service

import (
	"database/sql"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	authhttp "bitmerchant/internal/auth/ports/http"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/common/http/middleware"
	dashboardhttp "bitmerchant/internal/dashboard/ports/http"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	menuhttp "bitmerchant/internal/menu/ports/http"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	placesCmd "bitmerchant/internal/places/app/command"
	placesQuery "bitmerchant/internal/places/app/query"
	placeshttp "bitmerchant/internal/places/ports/http"
	restCmd "bitmerchant/internal/restaurant/app/command"
	restQuery "bitmerchant/internal/restaurant/app/query"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"
)

// Commands groups write use-cases for composition-root wiring.
type Commands struct {
	CreateOrder        orderCmd.CreateOrderHandler
	MarkOrderPaid      orderCmd.MarkOrderPaidHandler
	MarkOrderPreparing orderCmd.MarkOrderPreparingHandler
	MarkOrderReady     orderCmd.MarkOrderReadyHandler

	CreateRestaurant     restCmd.CreateRestaurantHandler
	ToggleRestaurantOpen restCmd.ToggleRestaurantOpenHandler
	UpdateTableCount     restCmd.UpdateRestaurantTableCountHandler

	CreateMenuCategory      menuCmd.CreateMenuCategoryHandler
	CreateMenuItem          menuCmd.CreateMenuItemHandler
	UpdateMenuItem          menuCmd.UpdateMenuItemHandler
	UpdateMenuCategory      menuCmd.UpdateMenuCategoryHandler
	ToggleMenuItemAvailable menuCmd.ToggleMenuItemAvailabilityHandler
	UploadMenuPhoto         menuCmd.UploadMenuItemPhotoHandler
	ReorderMenuCategories   menuCmd.ReorderMenuCategoriesHandler
	ReorderMenuItems        menuCmd.ReorderMenuItemsHandler

	RecordMenuVisit placesCmd.RecordMenuVisitHandler
}

// Queries groups read use-cases for composition-root wiring.
type Queries struct {
	GetMenu                menuQuery.MenuForCustomerHandler
	GetMenuForAdmin        menuQuery.MenuForAdminHandler
	GetCustomerOrder       orderQuery.CustomerOrderByLookupHandler
	GetCustomerOrders      orderQuery.CustomerOrdersForSessionHandler
	GetKitchenOrders       orderQuery.ActiveKitchenOrdersHandler
	ListVisitedRestaurants placesQuery.SessionVisitedPlacesHandler
	GenerateRestaurantQR   restQuery.RestaurantTableQRImageHandler
}

// Ports groups inbound adapters used by HTTP routing.
type Ports struct {
	Menu      *menuhttp.MenuHandler
	Cart      *orderinghttp.CartHandler
	Order     *orderinghttp.OrderHandler
	Places    *placeshttp.PlacesHandler
	Kitchen   *orderinghttp.KitchenHandler
	Admin     *restauranthttp.AdminHandler
	Owner     *restauranthttp.OwnerHandler
	Dashboard *dashboardhttp.DashboardHandler
	Auth      *authhttp.AuthHandler
	SSE       *commonhttp.SSEHandler

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
