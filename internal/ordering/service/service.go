package service

import (
	"context"
	"encoding/json"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	orderCart "bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderevent "bitmerchant/internal/ordering/app/event"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	ordersse "bitmerchant/internal/ordering/ports/sse"
	"bitmerchant/internal/wiring"
)

// Ordering bundles cart, order workflows, kitchen HTTP, and SSE event subscriptions for order events.
type Ordering struct {
	CartService *orderCart.CartService

	CreateOrder        orderCmd.CreateOrderHandler
	MarkOrderPaid      orderCmd.MarkOrderPaidHandler
	MarkOrderPreparing orderCmd.MarkOrderPreparingHandler
	MarkOrderReady     orderCmd.MarkOrderReadyHandler

	GetCustomerOrder  orderQuery.CustomerOrderByLookupHandler
	GetCustomerOrders orderQuery.CustomerOrdersForSessionHandler
	GetKitchenOrders  orderQuery.ActiveKitchenOrdersHandler

	CartHandler    *orderinghttp.CartHandler
	OrderHandler   *orderinghttp.OrderHandler
	KitchenHandler *orderinghttp.KitchenHandler
}

// New wires ordering bounded-context handlers and HTTP ports.
func New(
	repos wiring.Repositories,
	eventBus *events.EventBus,
	logger *logging.Logger,
) Ordering {
	cartService := orderCart.NewCartService()
	createOrderUC := orderCmd.NewCreateOrderHandler(repos.Order, repos.Restaurant, eventBus, logger.Logger, nil)
	getCustomerOrderByNumberUC := orderQuery.NewCustomerOrderByLookupHandler(repos.Order, nil, nil)
	getCustomerOrdersUC := orderQuery.NewCustomerOrdersForSessionHandler(repos.Order, nil, nil)
	getKitchenOrdersUC := orderQuery.NewActiveKitchenOrdersHandler(repos.Order, nil, nil)
	markPaidUC := orderCmd.NewMarkOrderPaidHandler(repos.Order, eventBus, logger.Logger, nil)
	markPreparingUC := orderCmd.NewMarkOrderPreparingHandler(repos.Order, eventBus, logger.Logger, nil)
	markReadyUC := orderCmd.NewMarkOrderReadyHandler(repos.Order, eventBus, logger.Logger, nil)

	return Ordering{
		CartService:        cartService,
		CreateOrder:        createOrderUC,
		MarkOrderPaid:      markPaidUC,
		MarkOrderPreparing: markPreparingUC,
		MarkOrderReady:     markReadyUC,
		GetCustomerOrder:   getCustomerOrderByNumberUC,
		GetCustomerOrders:  getCustomerOrdersUC,
		GetKitchenOrders:   getKitchenOrdersUC,
		CartHandler:        orderinghttp.NewCartHandler(cartService, repos.MenuItem),
		OrderHandler:       orderinghttp.NewOrderHandler(createOrderUC, getCustomerOrderByNumberUC, getCustomerOrdersUC, cartService),
		KitchenHandler:     orderinghttp.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC, repos.Restaurant, repos.Membership),
	}
}

// RegisterOrderSSESubscriptions connects order domain events to SSE projection handlers.
func RegisterOrderSSESubscriptions(eventBus *events.EventBus, logger *logging.Logger, sseHandler *commonhttp.SSEHandler, orderRepo order.Repository) {
	orderCreatedHandler := ordersse.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := ordersse.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := ordersse.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := ordersse.NewOrderReadyHandler(logger, sseHandler, orderRepo)

	subscribe(eventBus, common.EventOrderCreated, logger, func(msg []byte) {
		var event orderevent.OrderCreated
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderCreatedHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPaid, logger, func(msg []byte) {
		var event orderevent.OrderPaid
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPaidHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderPreparing, logger, func(msg []byte) {
		var event orderevent.OrderPreparing
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderPreparingHandler.Handle(context.Background(), event)
		}
	})

	subscribe(eventBus, common.EventOrderReady, logger, func(msg []byte) {
		var event orderevent.OrderReady
		if err := json.Unmarshal(msg, &event); err == nil {
			_ = orderReadyHandler.Handle(context.Background(), event)
		}
	})
}

func subscribe(bus *events.EventBus, topic string, logger *logging.Logger, handlerFunc func([]byte)) {
	go func() {
		msgs, err := bus.Subscribe(context.Background(), topic)
		if err != nil {
			logger.Error("Failed to subscribe", "topic", topic, "error", err)
			return
		}
		for msg := range msgs {
			handlerFunc(msg.Payload)
			msg.Ack()
		}
	}()
}
