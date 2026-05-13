package service

import (
	"log/slog"

	dashboardQuery "bitmerchant/internal/dashboard/app/query"
	dashboardhttp "bitmerchant/internal/dashboard/ports/http"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/wiring"
)

// Dashboard bundles dashboard read models and the owner dashboard HTTP adapter.
type Dashboard struct {
	GetStats    dashboardQuery.RestaurantDashboardStatsHandler
	GetHistory  dashboardQuery.PaidOrdersForRestaurantHandler
	GetTopItems dashboardQuery.TopSellingMenuItemsHandler
	GetStalled  dashboardQuery.StalledOrdersHandler
	GetByHour   dashboardQuery.OrdersByHourHandler
	HTTP        *dashboardhttp.DashboardHandler
}

// New wires dashboard queries and HTTP port. toggleOpen and pause must be the same handler instances used in service.Application.Commands.
func New(
	repos wiring.Repositories,
	toggleOpen restaurantCmd.ToggleRestaurantOpenHandler,
	pause restaurantCmd.PauseRestaurantHandler,
	logger *slog.Logger,
) Dashboard {
	if logger == nil {
		logger = slog.Default()
	}
	getStatsUC := dashboardQuery.NewRestaurantDashboardStatsHandler(repos.Order, nil, nil)
	getHistoryUC := dashboardQuery.NewPaidOrdersForRestaurantHandler(repos.Order, nil, nil)
	getTopItemsUC := dashboardQuery.NewTopSellingMenuItemsHandler(repos.Order, repos.MenuItem, nil, nil)
	getStalledUC := dashboardQuery.NewStalledOrdersHandler(repos.Order, nil, nil)
	getByHourUC := dashboardQuery.NewOrdersByHourHandler(repos.Order, nil, nil)
	return Dashboard{
		GetStats:    getStatsUC,
		GetHistory:  getHistoryUC,
		GetTopItems: getTopItemsUC,
		GetStalled:  getStalledUC,
		GetByHour:   getByHourUC,
		HTTP:        dashboardhttp.NewDashboardHandler(getStatsUC, getHistoryUC, getTopItemsUC, getStalledUC, getByHourUC, toggleOpen, pause, repos.Restaurant, repos.Order, repos.Membership, logger),
	}
}
