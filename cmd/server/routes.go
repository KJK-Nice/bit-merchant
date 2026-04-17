package main

import (
	"bitmerchant/internal/auth/domain/membership"
	authhttp "bitmerchant/internal/auth/ports/http"
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/common/http/middleware"
	dashboardhttp "bitmerchant/internal/dashboard/ports/http"
	menuhttp "bitmerchant/internal/menu/ports/http"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	placeshttp "bitmerchant/internal/places/ports/http"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"

	"github.com/labstack/echo/v4"
)

type routeHandlers struct {
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
}

func registerRoutes(e *echo.Echo, handlers routeHandlers, membershipRepo membership.Repository) {
	e.GET("/", handlers.Places.GetEntry)

	e.GET("/menu", handlers.Menu.GetMenu)
	e.GET("/my-places", handlers.Places.GetMyPlaces)
	e.GET("/scan", handlers.Places.GetScanQR)

	e.GET("/cart", handlers.Cart.GetCart)
	e.POST("/cart/add", handlers.Cart.AddToCart)
	e.POST("/cart/decrement", handlers.Cart.DecrementFromCart)
	e.POST("/cart/remove", handlers.Cart.RemoveFromCart)

	e.GET("/order/lookup", handlers.Order.GetLookup)
	e.POST("/order/lookup", handlers.Order.PostLookup)
	e.GET("/order/confirm", handlers.Order.GetConfirmOrder)
	e.POST("/order/create", handlers.Order.CreateOrder)
	e.GET("/order/:orderNumber", handlers.Order.GetOrder)
	e.GET("/order/:orderNumber/stream", handlers.SSE.OrderStatusStream)

	kitchenGroup := e.Group("/kitchen")
	kitchenGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, common.RoleOwner, common.RoleKitchenStaff))
	kitchenGroup.GET("", handlers.Kitchen.GetKitchen)
	kitchenGroup.GET("/stream", handlers.SSE.KitchenStream)
	kitchenGroup.POST("/order/:id/mark-paid", handlers.Kitchen.MarkPaid)
	kitchenGroup.POST("/order/:id/mark-preparing", handlers.Kitchen.MarkPreparing)
	kitchenGroup.POST("/order/:id/mark-ready", handlers.Kitchen.MarkReady)

	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, common.RoleOwner))
	adminGroup.GET("/dashboard", handlers.Admin.Dashboard)
	adminGroup.POST("/category", handlers.Admin.CreateCategory)
	adminGroup.POST("/category/:id/update", handlers.Admin.UpdateCategory)
	adminGroup.POST("/item", handlers.Admin.CreateItem)
	adminGroup.POST("/item/:id/update", handlers.Admin.UpdateItem)
	adminGroup.POST("/item/:id/toggle-availability", handlers.Admin.ToggleItemAvailability)
	adminGroup.POST("/item/:id/photo", handlers.Admin.UploadPhoto)
	adminGroup.POST("/menu/reorder-categories", handlers.Admin.PostReorderCategories)
	adminGroup.POST("/menu/reorder-items", handlers.Admin.PostReorderItems)
	adminGroup.GET("/qr", handlers.Admin.GetQRPage)
	adminGroup.POST("/qr/settings", handlers.Admin.PostQRSettings)
	adminGroup.GET("/qr/print", handlers.Admin.GetQRPrint)
	adminGroup.GET("/qr/table/:table", handlers.Admin.GetQRTablePNG)

	e.GET("/owner/signup", handlers.Owner.GetSignup)
	e.POST("/owner/signup", handlers.Owner.PostSignup)

	e.GET("/auth/signup", handlers.Auth.GetSignup)
	e.GET("/auth/login", handlers.Auth.GetLogin)
	e.GET("/auth/invite/:token", handlers.Auth.GetInvite)
	e.POST("/auth/register/begin", handlers.Auth.BeginRegistration)
	e.POST("/auth/register/finish", handlers.Auth.FinishRegistration)
	e.POST("/auth/login/begin", handlers.Auth.BeginLogin)
	e.POST("/auth/login/finish", handlers.Auth.FinishLogin)
	e.POST("/auth/logout", handlers.Auth.Logout)

	authSelectionGroup := e.Group("/auth")
	authSelectionGroup.Use(middleware.RequireAuth())
	authSelectionGroup.GET("/profile", handlers.Auth.GetProfile)
	authSelectionGroup.GET("/restaurants/new", handlers.Auth.GetNewRestaurant)
	authSelectionGroup.POST("/restaurants", handlers.Auth.PostNewRestaurant)
	authSelectionGroup.GET("/select-restaurant", handlers.Auth.GetSelectRestaurant)
	authSelectionGroup.POST("/select-restaurant", handlers.Auth.PostSelectRestaurant)

	dashboardGroup := e.Group("/dashboard")
	dashboardGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, common.RoleOwner))
	dashboardGroup.GET("", handlers.Dashboard.Dashboard)
	dashboardGroup.GET("/menu", handlers.Admin.GetMenu)
	dashboardGroup.GET("/qr-code", handlers.Admin.GetQRCode)
	dashboardGroup.POST("/toggle-open", handlers.Dashboard.ToggleOpen)
	dashboardGroup.POST("/invite", handlers.Auth.CreateInvitation)
}
