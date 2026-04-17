package service

import (
	"bitmerchant/internal/infrastructure/qr"
	menuservice "bitmerchant/internal/menu/service"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"
	"bitmerchant/internal/wiring"
)

// Restaurant bundles restaurant lifecycle commands, QR query, and merchant HTTP ports (admin/owner).
type Restaurant struct {
	CreateRestaurant     restaurantCmd.CreateRestaurantHandler
	ToggleRestaurantOpen restaurantCmd.ToggleRestaurantOpenHandler
	UpdateTableCount     restaurantCmd.UpdateRestaurantTableCountHandler
	GenerateRestaurantQR restaurantQuery.RestaurantTableQRImageHandler
	Admin                *restauranthttp.AdminHandler
	Owner                *restauranthttp.OwnerHandler
}

// New wires restaurant bounded-context handlers and admin/owner HTTP adapters.
func New(
	repos wiring.Repositories,
	cfg wiring.Config,
	qrService *qr.QRCodeService,
	menu menuservice.Menu,
) Restaurant {
	createRestUC := restaurantCmd.NewCreateRestaurantHandler(repos.Restaurant, nil, nil)
	toggleOpenUC := restaurantCmd.NewToggleRestaurantOpenHandler(repos.Restaurant, nil, nil)
	updateTableCountUC := restaurantCmd.NewUpdateRestaurantTableCountHandler(repos.Restaurant, nil, nil)
	generateQRUC := restaurantQuery.NewRestaurantTableQRImageHandler(qrService, cfg.CustomerBaseURL, repos.Restaurant, nil, nil)

	adminHandler := restauranthttp.NewAdminHandler(
		createRestUC,
		menu.CreateMenuCategory,
		menu.CreateMenuItem,
		menu.GetMenuForAdmin,
		menu.UpdateMenuItem,
		menu.UpdateMenuCategory,
		menu.ToggleItemAvailability,
		menu.UploadMenuPhoto,
		menu.ReorderMenuCategories,
		menu.ReorderMenuItems,
		repos.MenuItem,
		updateTableCountUC,
		generateQRUC,
		repos.Membership,
		repos.Restaurant,
	)
	ownerHandler := restauranthttp.NewOwnerHandler(createRestUC)

	return Restaurant{
		CreateRestaurant:     createRestUC,
		ToggleRestaurantOpen: toggleOpenUC,
		UpdateTableCount:     updateTableCountUC,
		GenerateRestaurantQR: generateQRUC,
		Admin:                adminHandler,
		Owner:                ownerHandler,
	}
}
