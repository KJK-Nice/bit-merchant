package service

import (
	"encoding/json"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	orderCart "bitmerchant/internal/ordering/app/cart"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderevent "bitmerchant/internal/ordering/app/event"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
	orderinghttp "bitmerchant/internal/ordering/ports/http"
	ordersse "bitmerchant/internal/ordering/ports/sse"
	"bitmerchant/internal/wiring"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Ordering bundles cart, order workflows, kitchen HTTP, and SSE event subscriptions for order events.
type Ordering struct {
	CartService *orderCart.CartService

	CreateOrder        orderCmd.CreateOrderHandler
	MarkOrderPaid      orderCmd.MarkOrderPaidHandler
	MarkOrderPreparing orderCmd.MarkOrderPreparingHandler
	MarkOrderReady     orderCmd.MarkOrderReadyHandler
	MarkOrderCompleted orderCmd.MarkOrderCompletedHandler

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
	eventBus common.EventBus,
	logger *logging.Logger,
	vapidPublicKey string,
) Ordering {
	cartService := orderCart.NewCartService()
	createOrderUC := orderCmd.NewCreateOrderHandler(repos.Order, repos.Restaurant, eventBus, logger.Logger, nil)
	getCustomerOrderByNumberUC := orderQuery.NewCustomerOrderByLookupHandler(repos.Order, nil, nil)
	getCustomerOrdersUC := orderQuery.NewCustomerOrdersForSessionHandler(repos.Order, nil, nil)
	getKitchenOrdersUC := orderQuery.NewActiveKitchenOrdersHandler(repos.Order, nil, nil)
	markPaidUC := orderCmd.NewMarkOrderPaidHandler(repos.Order, eventBus, logger.Logger, nil)
	markPreparingUC := orderCmd.NewMarkOrderPreparingHandler(repos.Order, eventBus, logger.Logger, nil)
	markReadyUC := orderCmd.NewMarkOrderReadyHandler(repos.Order, eventBus, logger.Logger, nil)
	markCompletedUC := orderCmd.NewMarkOrderCompletedHandler(repos.Order, eventBus, logger.Logger, nil)

	return Ordering{
		CartService:        cartService,
		CreateOrder:        createOrderUC,
		MarkOrderPaid:      markPaidUC,
		MarkOrderPreparing: markPreparingUC,
		MarkOrderReady:     markReadyUC,
		MarkOrderCompleted: markCompletedUC,
		GetCustomerOrder:   getCustomerOrderByNumberUC,
		GetCustomerOrders:  getCustomerOrdersUC,
		GetKitchenOrders:   getKitchenOrdersUC,
		CartHandler:        orderinghttp.NewCartHandler(cartService, repos.MenuItem),
		OrderHandler:       orderinghttp.NewOrderHandler(createOrderUC, getCustomerOrderByNumberUC, getCustomerOrdersUC, repos.Order, cartService, vapidPublicKey),
		KitchenHandler:     orderinghttp.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC, markCompletedUC, repos.Restaurant, repos.Membership, vapidPublicKey),
	}
}

// RegisterOrderSSEHandlers connects order domain events to SSE projection handlers through Watermill Router.
func RegisterOrderSSEHandlers(router *message.Router, subscriber message.Subscriber, logger *logging.Logger, sseHandler *commonhttp.SSEHandler, orderRepo order.Repository) {
	orderCreatedHandler := ordersse.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := ordersse.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := ordersse.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := ordersse.NewOrderReadyHandler(logger, sseHandler, orderRepo)
	orderCompletedHandler := ordersse.NewOrderCompletedHandler(logger, sseHandler, orderRepo)

	router.AddConsumerHandler("sse_order_created", common.EventOrderCreated, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderCreated
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order created event", "error", err)
			return nil
		}
		return orderCreatedHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_order_paid", common.EventOrderPaid, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderPaid
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order paid event", "error", err)
			return nil
		}
		return orderPaidHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_order_preparing", common.EventOrderPreparing, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderPreparing
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order preparing event", "error", err)
			return nil
		}
		return orderPreparingHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_order_ready", common.EventOrderReady, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderReady
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order ready event", "error", err)
			return nil
		}
		return orderReadyHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_order_completed", common.EventOrderCompleted, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderCompleted
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order completed event", "error", err)
			return nil
		}
		return orderCompletedHandler.Handle(msg.Context(), event)
	})
}
