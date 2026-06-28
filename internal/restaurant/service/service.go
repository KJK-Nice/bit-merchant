package service

import (
	"bitmerchant/internal/infrastructure/qr"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	menuservice "bitmerchant/internal/menu/service"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"
	"bitmerchant/internal/wiring"
)

// Restaurant bundles restaurant lifecycle commands, QR query, and merchant HTTP ports (admin/owner).
type Restaurant struct {
	CreateRestaurant        restaurantCmd.CreateRestaurantHandler
	ToggleRestaurantOpen    restaurantCmd.ToggleRestaurantOpenHandler
	PauseRestaurant         restaurantCmd.PauseRestaurantHandler
	UpdateTableCount        restaurantCmd.UpdateRestaurantTableCountHandler
	UpdateKitchenThresholds restaurantCmd.UpdateKitchenThresholdsHandler
	GenerateRestaurantQR    restaurantQuery.RestaurantTableQRImageHandler
	Admin                   *restauranthttp.AdminHandler
	Owner                   *restauranthttp.OwnerHandler
}

// New wires restaurant bounded-context handlers and admin/owner HTTP adapters.
func New(
	repos wiring.Repositories,
	cfg wiring.Config,
	qrService *qr.QRCodeService,
	menuSvc menuservice.Menu,
	photoStorage menu.PhotoStorage,
) Restaurant {
	createRestUC := restaurantCmd.NewCreateRestaurantHandler(repos.Restaurant, nil, nil)
	toggleOpenUC := restaurantCmd.NewToggleRestaurantOpenHandler(repos.Restaurant, nil, nil)
	pauseRestUC := restaurantCmd.NewPauseRestaurantHandler(repos.Restaurant, nil, nil)
	updateTableCountUC := restaurantCmd.NewUpdateRestaurantTableCountHandler(repos.Restaurant, nil, nil)
	updateKitchenThresholdsUC := restaurantCmd.NewUpdateKitchenThresholdsHandler(repos.Restaurant, nil, nil)
	generateQRUC := restaurantQuery.NewRestaurantTableQRImageHandler(qrService, cfg.CustomerBaseURL, repos.Restaurant, nil, nil)

	adminHandler := restauranthttp.NewAdminHandler(
		createRestUC,
		menuSvc.CreateMenuCategory,
		menuSvc.CreateMenuItem,
		menuSvc.GetMenuForAdmin,
		menuSvc.UpdateMenuItem,
		menuSvc.UpdateMenuCategory,
		menuSvc.ToggleItemAvailability,
		menuSvc.UploadMenuPhoto,
		menuSvc.ReorderMenuCategories,
		menuSvc.ReorderMenuItems,
		repos.MenuItem,
		photoStorage,
		menuQuery.PhotoSignerConfig{
			Bucket:        cfg.S3BucketName,
			Endpoint:      cfg.S3Endpoint,
			PublicBaseURL: cfg.S3PublicBaseURL,
		},
		updateTableCountUC,
		updateKitchenThresholdsUC,
		generateQRUC,
		repos.Membership,
		repos.Restaurant,
	)
	ownerHandler := restauranthttp.NewOwnerHandler(createRestUC)

	return Restaurant{
		CreateRestaurant:        createRestUC,
		ToggleRestaurantOpen:    toggleOpenUC,
		PauseRestaurant:         pauseRestUC,
		UpdateTableCount:        updateTableCountUC,
		UpdateKitchenThresholds: updateKitchenThresholdsUC,
		GenerateRestaurantQR:    generateQRUC,
		Admin:                   adminHandler,
		Owner:                   ownerHandler,
	}
}
