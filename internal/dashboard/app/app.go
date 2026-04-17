package app

import "bitmerchant/internal/dashboard/app/query"

// Application bundles dashboard (analytics) queries.
type Application struct {
	Queries Queries
}

type Queries struct {
	RestaurantDashboardStats query.RestaurantDashboardStatsHandler
	PaidOrdersForRestaurant  query.PaidOrdersForRestaurantHandler
	TopSellingMenuItems      query.TopSellingMenuItemsHandler
}
