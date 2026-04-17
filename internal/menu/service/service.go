package service

import (
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	menuhttp "bitmerchant/internal/menu/ports/http"
	orderCart "bitmerchant/internal/ordering/app/cart"
	placesCmd "bitmerchant/internal/places/app/command"
	"bitmerchant/internal/wiring"
)

// Menu bundles menu catalog handlers and the customer HTTP port.
type Menu struct {
	GetMenu                menuQuery.MenuForCustomerHandler
	GetMenuForAdmin        menuQuery.MenuForAdminHandler
	UpdateMenuItem         menuCmd.UpdateMenuItemHandler
	UpdateMenuCategory     menuCmd.UpdateMenuCategoryHandler
	ToggleItemAvailability menuCmd.ToggleMenuItemAvailabilityHandler
	CreateMenuCategory     menuCmd.CreateMenuCategoryHandler
	CreateMenuItem         menuCmd.CreateMenuItemHandler
	UploadMenuPhoto        menuCmd.UploadMenuItemPhotoHandler
	ReorderMenuCategories  menuCmd.ReorderMenuCategoriesHandler
	ReorderMenuItems       menuCmd.ReorderMenuItemsHandler
	HTTP                   *menuhttp.MenuHandler
}

// New wires menu bounded-context command/query handlers and the public menu HTTP handler.
func New(
	repos wiring.Repositories,
	photoStorage menu.PhotoStorage,
	cfg wiring.Config,
	cartService *orderCart.CartService,
	recordMenuVisitUC placesCmd.RecordMenuVisitHandler,
) Menu {
	signer := menuQuery.PhotoSignerConfig{
		Bucket:        cfg.S3BucketName,
		Endpoint:      cfg.S3Endpoint,
		PublicBaseURL: cfg.S3PublicBaseURL,
	}
	getMenuUC := menuQuery.NewMenuForCustomerHandler(repos.MenuCategory, repos.MenuItem, repos.Restaurant, photoStorage, signer, nil, nil)
	getMenuAdminUC := menuQuery.NewMenuForAdminHandler(repos.MenuCategory, repos.MenuItem, repos.Restaurant, photoStorage, signer, nil, nil)
	updateMenuItemUC := menuCmd.NewUpdateMenuItemHandler(repos.MenuItem, repos.MenuCategory, nil, nil)
	updateMenuCategoryUC := menuCmd.NewUpdateMenuCategoryHandler(repos.MenuCategory, nil, nil)
	toggleItemAvailUC := menuCmd.NewToggleMenuItemAvailabilityHandler(repos.MenuItem, nil, nil)
	createCatUC := menuCmd.NewCreateMenuCategoryHandler(repos.MenuCategory, nil, nil)
	createItemUC := menuCmd.NewCreateMenuItemHandler(repos.MenuItem, nil, nil)
	uploadPhotoUC := menuCmd.NewUploadMenuItemPhotoHandler(repos.MenuItem, photoStorage, nil, nil)
	reorderCategoriesUC := menuCmd.NewReorderMenuCategoriesHandler(repos.MenuCategory, nil, nil)
	reorderItemsUC := menuCmd.NewReorderMenuItemsHandler(repos.MenuItem, repos.MenuCategory, nil, nil)

	return Menu{
		GetMenu:                getMenuUC,
		GetMenuForAdmin:        getMenuAdminUC,
		UpdateMenuItem:         updateMenuItemUC,
		UpdateMenuCategory:     updateMenuCategoryUC,
		ToggleItemAvailability: toggleItemAvailUC,
		CreateMenuCategory:     createCatUC,
		CreateMenuItem:         createItemUC,
		UploadMenuPhoto:        uploadPhotoUC,
		ReorderMenuCategories:  reorderCategoriesUC,
		ReorderMenuItems:       reorderItemsUC,
		HTTP:                   menuhttp.NewMenuHandler(getMenuUC, cartService, recordMenuVisitUC),
	}
}
