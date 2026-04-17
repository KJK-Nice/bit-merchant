package app

import (
	"bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/app/query"
)

// Application bundles restaurant bounded-context use cases.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands are write-side handlers.
type Commands struct {
	CreateRestaurant           command.CreateRestaurantHandler
	ToggleRestaurantOpen       command.ToggleRestaurantOpenHandler
	UpdateRestaurantTableCount command.UpdateRestaurantTableCountHandler
}

// Queries are read-side handlers.
type Queries struct {
	RestaurantTableQRImage query.RestaurantTableQRImageHandler
}
