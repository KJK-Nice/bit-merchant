package service

import (
	"encoding/json"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
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

	CreateOrder         orderCmd.CreateOrderHandler
	MarkOrderPaid       orderCmd.MarkOrderPaidHandler
	MarkOrderPreparing  orderCmd.MarkOrderPreparingHandler
	MarkOrderReady      orderCmd.MarkOrderReadyHandler
	MarkOrderCompleted  orderCmd.MarkOrderCompletedHandler
	ToggleOrderItemPrep orderCmd.ToggleOrderItemPrepHandler
	RequestServer       orderCmd.RequestServerHandler
	RequestBill         orderCmd.RequestBillHandler

	GetCustomerOrder  orderQuery.CustomerOrderByLookupHandler
	GetCustomerOrders orderQuery.CustomerOrdersForSessionHandler
	GetKitchenOrders  orderQuery.ActiveKitchenOrdersHandler
	GetUnpaidServer   orderQuery.UnpaidServerOrdersHandler

	CartHandler    *orderinghttp.CartHandler
	OrderHandler   *orderinghttp.OrderHandler
	KitchenHandler *orderinghttp.KitchenHandler
	ServerHandler  *orderinghttp.ServerHandler
}

// New wires ordering bounded-context handlers and HTTP ports.
//
// photoStorage may be nil; when missing, the customer item-detail page falls
// back to rendering raw PhotoURLs (e.g. dev environments without S3).
func New(
	repos wiring.Repositories,
	eventBus common.EventBus,
	logger *logging.Logger,
	vapidPublicKey string,
	photoStorage menu.PhotoStorage,
	cfg wiring.Config,
) Ordering {
	cartService := orderCart.NewCartService()
	createOrderUC := orderCmd.NewCreateOrderHandler(repos.Order, repos.Restaurant, eventBus, logger.Logger, nil)
	getCustomerOrderByNumberUC := orderQuery.NewCustomerOrderByLookupHandler(repos.Order, nil, nil)
	getCustomerOrdersUC := orderQuery.NewCustomerOrdersForSessionHandler(repos.Order, nil, nil)
	getKitchenOrdersUC := orderQuery.NewActiveKitchenOrdersHandler(repos.Order, nil, nil)
	getUnpaidServerUC := orderQuery.NewUnpaidServerOrdersHandler(repos.Order, nil, nil)
	markPaidUC := orderCmd.NewMarkOrderPaidHandler(repos.Order, eventBus, logger.Logger, nil)
	markPreparingUC := orderCmd.NewMarkOrderPreparingHandler(repos.Order, eventBus, logger.Logger, nil)
	markReadyUC := orderCmd.NewMarkOrderReadyHandler(repos.Order, eventBus, logger.Logger, nil)
	markCompletedUC := orderCmd.NewMarkOrderCompletedHandler(repos.Order, eventBus, logger.Logger, nil)
	toggleItemPrepUC := orderCmd.NewToggleOrderItemPrepHandler(repos.Order, eventBus, logger.Logger, nil)
	requestServerUC := orderCmd.NewRequestServerHandler(repos.Order, eventBus, logger.Logger, nil)
	requestBillUC := orderCmd.NewRequestBillHandler(repos.Order, eventBus, logger.Logger, nil)

	return Ordering{
		CartService:         cartService,
		CreateOrder:         createOrderUC,
		MarkOrderPaid:       markPaidUC,
		MarkOrderPreparing:  markPreparingUC,
		MarkOrderReady:      markReadyUC,
		MarkOrderCompleted:  markCompletedUC,
		ToggleOrderItemPrep: toggleItemPrepUC,
		RequestServer:       requestServerUC,
		RequestBill:         requestBillUC,
		GetCustomerOrder:    getCustomerOrderByNumberUC,
		GetCustomerOrders:   getCustomerOrdersUC,
		GetKitchenOrders:    getKitchenOrdersUC,
		GetUnpaidServer:     getUnpaidServerUC,
		CartHandler: orderinghttp.NewCartHandler(cartService, repos.MenuItem, photoStorage, menuQuery.PhotoSignerConfig{
			Bucket:        cfg.S3BucketName,
			Endpoint:      cfg.S3Endpoint,
			PublicBaseURL: cfg.S3PublicBaseURL,
		}),
		OrderHandler:   orderinghttp.NewOrderHandler(createOrderUC, getCustomerOrderByNumberUC, getCustomerOrdersUC, requestServerUC, requestBillUC, repos.Order, repos.Restaurant, cartService, vapidPublicKey),
		KitchenHandler: orderinghttp.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC, markCompletedUC, toggleItemPrepUC, repos.Restaurant, repos.Membership, vapidPublicKey),
		ServerHandler:  orderinghttp.NewServerHandler(getUnpaidServerUC, markPaidUC, repos.Restaurant, repos.Membership),
	}
}

// RegisterOrderSSEHandlers connects order domain events to SSE projection handlers through Watermill Router.
func RegisterOrderSSEHandlers(router *message.Router, subscriber message.Subscriber, logger *logging.Logger, sseHandler *commonhttp.SSEHandler, orderRepo order.Repository) {
	orderCreatedHandler := ordersse.NewOrderCreatedHandler(logger, sseHandler, orderRepo)
	orderPaidHandler := ordersse.NewOrderPaidHandler(logger, sseHandler, orderRepo)
	orderPreparingHandler := ordersse.NewOrderPreparingHandler(logger, sseHandler, orderRepo)
	orderReadyHandler := ordersse.NewOrderReadyHandler(logger, sseHandler, orderRepo)
	orderCompletedHandler := ordersse.NewOrderCompletedHandler(logger, sseHandler, orderRepo)
	orderItemPrepToggledHandler := ordersse.NewOrderItemPrepToggledHandler(logger, sseHandler, orderRepo)
	serverCalledHandler := ordersse.NewServerCalledHandler(logger, sseHandler, orderRepo)
	billRequestedHandler := ordersse.NewBillRequestedHandler(logger, sseHandler, orderRepo)

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

	router.AddConsumerHandler("sse_order_item_prep_toggled", common.EventOrderItemPrepToggled, subscriber, func(msg *message.Message) error {
		var event orderevent.OrderItemPrepToggled
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed order item prep toggled event", "error", err)
			return nil
		}
		return orderItemPrepToggledHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_server_called", common.EventServerCalled, subscriber, func(msg *message.Message) error {
		var event orderevent.ServerCalled
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed server called event", "error", err)
			return nil
		}
		return serverCalledHandler.Handle(msg.Context(), event)
	})

	router.AddConsumerHandler("sse_bill_requested", common.EventBillRequested, subscriber, func(msg *message.Message) error {
		var event orderevent.BillRequested
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Warn("Skipping malformed bill requested event", "error", err)
			return nil
		}
		return billRequestedHandler.Handle(msg.Context(), event)
	})
}
