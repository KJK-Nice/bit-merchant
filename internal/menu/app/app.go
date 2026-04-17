package app

import (
	"bitmerchant/internal/menu/app/command"
	"bitmerchant/internal/menu/app/query"
)

// Application bundles menu bounded-context handlers.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands are write-side handlers.
type Commands struct {
	CreateMenuCategory      command.CreateMenuCategoryHandler
	CreateMenuItem          command.CreateMenuItemHandler
	UpdateMenuCategory      command.UpdateMenuCategoryHandler
	UpdateMenuItem          command.UpdateMenuItemHandler
	ToggleMenuItemAvailable command.ToggleMenuItemAvailabilityHandler
	UploadMenuItemPhoto     command.UploadMenuItemPhotoHandler
	ReorderMenuCategories   command.ReorderMenuCategoriesHandler
	ReorderMenuItems        command.ReorderMenuItemsHandler
}

// Queries are read-side handlers.
type Queries struct {
	MenuForCustomer query.MenuForCustomerHandler
	MenuForAdmin    query.MenuForAdminHandler
}
