package command

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/ordering/app/cart"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type CreateOrderRequest struct {
	RestaurantID  common.RestaurantID
	SessionID     string
	Cart          *cart.Cart
	PaymentMethod common.PaymentMethodType
}

type CreateOrderResponse struct {
	OrderID     common.OrderID
	OrderNumber common.OrderNumber
}

type CreateOrderUseCase struct {
	orderRepo order.Repository
	restRepo  restaurant.Repository
	eventBus  *events.EventBus
	logger    *logging.Logger
}

func NewCreateOrderUseCase(
	orderRepo order.Repository,
	restRepo restaurant.Repository,
	eventBus *events.EventBus,
	logger *logging.Logger,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo: orderRepo,
		restRepo:  restRepo,
		eventBus:  eventBus,
		logger:    logger,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	rest, err := uc.restRepo.FindByID(req.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !rest.IsOpen {
		return nil, fmt.Errorf("restaurant is currently closed")
	}

	orderID := common.OrderID(fmt.Sprintf("ord_%d", time.Now().UnixNano()))
	orderNumber := common.OrderNumber(fmt.Sprintf("%04d", rand.Intn(10000)))

	orderItems, err := uc.createOrderItems(req.Cart.Items, orderID)
	if err != nil {
		return nil, err
	}

	fiatAmount := req.Cart.Total
	totalAmount := int64(fiatAmount * 100)

	o, err := order.NewOrder(orderID, orderNumber, req.RestaurantID, req.SessionID, orderItems, totalAmount, req.PaymentMethod)
	if err != nil {
		return nil, err
	}
	o.FiatAmount = fiatAmount

	if err := uc.orderRepo.Save(o); err != nil {
		return nil, err
	}

	uc.publishOrderCreatedEvent(ctx, o)
	uc.logger.Info("Order created", "orderID", o.ID, "amount", o.FiatAmount)

	return &CreateOrderResponse{
		OrderID:     o.ID,
		OrderNumber: o.OrderNumber,
	}, nil
}

func (uc *CreateOrderUseCase) createOrderItems(cartItems []cart.CartItem, orderID common.OrderID) ([]order.OrderItem, error) {
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

func (uc *CreateOrderUseCase) publishOrderCreatedEvent(ctx context.Context, o *order.Order) {
	event := order.OrderCreated{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		CreatedAt:    o.CreatedAt,
	}
	if err := uc.eventBus.Publish(ctx, common.EventOrderCreated, event); err != nil {
		uc.logger.Warn("Failed to publish order created event", "orderID", o.ID, "error", err)
	}
}
