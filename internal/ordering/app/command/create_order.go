package command

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/app/cart"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

// CreateOrder places a new order from a session cart.
type CreateOrder struct {
	RestaurantID  common.RestaurantID
	SessionID     string
	Cart          *cart.Cart
	PaymentMethod common.PaymentMethodType
}

// CreateOrderResult is returned after a successful create.
type CreateOrderResult struct {
	OrderID     common.OrderID
	OrderNumber common.OrderNumber
}

type CreateOrderHandler decorator.CommandResultHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo order.Repository
	restRepo  restaurant.Repository
	eventBus  common.EventBus
	log       *slog.Logger
}

func NewCreateOrderHandler(
	orderRepo order.Repository,
	restRepo restaurant.Repository,
	eventBus common.EventBus,
	log *slog.Logger,
	metrics decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil || restRepo == nil {
		panic("nil repository")
	}
	h := createOrderHandler{
		orderRepo: orderRepo,
		restRepo:  restRepo,
		eventBus:  eventBus,
		log:       log,
	}
	return decorator.ApplyCommandResultDecorators[CreateOrder, *CreateOrderResult](h, log, metrics)
}

func (h createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	rest, err := h.restRepo.FindByID(cmd.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !rest.IsOpen {
		return nil, fmt.Errorf("restaurant is currently closed")
	}

	orderID := common.OrderID(fmt.Sprintf("ord_%d", time.Now().UnixNano()))
	orderNumber := common.OrderNumber(fmt.Sprintf("%04d", rand.Intn(10000)))

	orderItems, err := h.createOrderItems(cmd.Cart.Items, orderID)
	if err != nil {
		return nil, err
	}

	fiatAmount := cmd.Cart.Total
	totalAmount := int64(fiatAmount * 100)

	o, err := order.NewOrder(orderID, orderNumber, cmd.RestaurantID, cmd.SessionID, orderItems, totalAmount, cmd.PaymentMethod)
	if err != nil {
		return nil, err
	}
	o.FiatAmount = fiatAmount

	if err := h.orderRepo.Save(o); err != nil {
		return nil, err
	}

	h.publishOrderCreatedEvent(ctx, o)
	if h.log != nil {
		h.log.InfoContext(ctx, "Order created", "orderID", o.ID, "amount", o.FiatAmount)
	}

	return &CreateOrderResult{
		OrderID:     o.ID,
		OrderNumber: o.OrderNumber,
	}, nil
}

func (h createOrderHandler) createOrderItems(cartItems []cart.CartItem, orderID common.OrderID) ([]order.OrderItem, error) {
	var orderItems []order.OrderItem
	for _, item := range cartItems {
		orderItemID := common.OrderItemID(fmt.Sprintf("oi_%d_%s", time.Now().UnixNano(), item.ItemID))
		oi, err := order.NewOrderItem(orderItemID, orderID, item.ItemID, item.Name, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, *oi)
	}
	return orderItems, nil
}

func (h createOrderHandler) publishOrderCreatedEvent(ctx context.Context, o *order.Order) {
	ev := event.OrderCreated{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		TotalAmount:  o.TotalAmount,
		CreatedAt:    o.CreatedAt,
	}
	if err := h.eventBus.Publish(ctx, common.EventOrderCreated, ev); err != nil && h.log != nil {
		h.log.WarnContext(ctx, "Failed to publish order created event", "orderID", o.ID, "error", err)
	}
}
