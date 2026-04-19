package app

import (
	"bitmerchant/internal/ordering/app/command"
	"bitmerchant/internal/ordering/app/query"
)

// Application bundles ordering bounded-context handlers.
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands are write-side handlers.
type Commands struct {
	CreateOrder        command.CreateOrderHandler
	MarkOrderPaid      command.MarkOrderPaidHandler
	MarkOrderPreparing command.MarkOrderPreparingHandler
	MarkOrderReady     command.MarkOrderReadyHandler
	MarkOrderCompleted command.MarkOrderCompletedHandler
}

// Queries are read-side handlers.
type Queries struct {
	OrderByNumberForRestaurant query.OrderByNumberForRestaurantHandler
	CustomerOrdersForSession   query.CustomerOrdersForSessionHandler
	CustomerOrderByLookup      query.CustomerOrderByLookupHandler
	ActiveKitchenOrders        query.ActiveKitchenOrdersHandler
}
