package main

import (
	"net/http"

	"bitmerchant/internal/domain"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
)

type routeHandlers struct {
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
}

func registerRoutes(e *echo.Echo, handlers routeHandlers, membershipRepo domain.MembershipRepository) {
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/menu")
	})

	e.GET("/menu", handlers.Menu.GetMenu)
	e.GET("/my-places", handlers.Places.GetMyPlaces)

	e.GET("/cart", handlers.Cart.GetCart)
	e.POST("/cart/add", handlers.Cart.AddToCart)
	e.POST("/cart/remove", handlers.Cart.RemoveFromCart)

	e.GET("/order/lookup", handlers.Order.GetLookup)
	e.POST("/order/lookup", handlers.Order.PostLookup)
	e.GET("/order/confirm", handlers.Order.GetConfirmOrder)
	e.POST("/order/create", handlers.Order.CreateOrder)
	e.GET("/order/:orderNumber", handlers.Order.GetOrder)
	e.GET("/order/:orderNumber/stream", handlers.SSE.OrderStatusStream)

	kitchenGroup := e.Group("/kitchen")
	kitchenGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner, domain.RoleKitchenStaff))
	kitchenGroup.GET("", handlers.Kitchen.GetKitchen)
	kitchenGroup.GET("/stream", handlers.SSE.KitchenStream)
	kitchenGroup.POST("/order/:id/mark-paid", handlers.Kitchen.MarkPaid)
	kitchenGroup.POST("/order/:id/mark-preparing", handlers.Kitchen.MarkPreparing)
	kitchenGroup.POST("/order/:id/mark-ready", handlers.Kitchen.MarkReady)

	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner))
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
	dashboardGroup.Use(middleware.RequireAuth(), middleware.RequireRole(membershipRepo, domain.RoleOwner))
	dashboardGroup.GET("", handlers.Dashboard.Dashboard)
	dashboardGroup.GET("/menu", handlers.Admin.GetMenu)
	dashboardGroup.GET("/qr-code", handlers.Admin.GetQRCode)
	dashboardGroup.POST("/toggle-open", handlers.Dashboard.ToggleOpen)
	dashboardGroup.POST("/invite", handlers.Auth.CreateInvitation)
}
