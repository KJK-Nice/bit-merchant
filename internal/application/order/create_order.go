package order

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
)

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	RestaurantID  domain.RestaurantID
	SessionID     string
	Cart          *cart.Cart
	PaymentMethod domain.PaymentMethodType
}

// CreateOrderResponse represents order creation response
type CreateOrderResponse struct {
	OrderID     domain.OrderID
	OrderNumber domain.OrderNumber
}

// CreateOrderUseCase handles order creation
type CreateOrderUseCase struct {
	orderRepo     domain.OrderRepository
	paymentRepo   domain.PaymentRepository
	restRepo      domain.RestaurantRepository
	eventBus      *events.EventBus
	paymentMethod domain.PaymentMethod
	logger        *logging.Logger
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(
	orderRepo domain.OrderRepository,
	paymentRepo domain.PaymentRepository,
	restRepo domain.RestaurantRepository,
	eventBus *events.EventBus,
	paymentMethod domain.PaymentMethod,
	logger *logging.Logger,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		restRepo:      restRepo,
		eventBus:      eventBus,
		paymentMethod: paymentMethod,
		logger:        logger,
	}
}

// Execute creates an order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// Check if restaurant is open
	restaurant, err := uc.restRepo.FindByID(req.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !restaurant.IsOpen {
		return nil, fmt.Errorf("restaurant is currently closed")
	}

	orderID := uc.generateOrderID()
	orderNumber := uc.generateOrderNumber()

	orderItems, err := uc.createOrderItems(req.Cart.Items, orderID)
	if err != nil {
		return nil, err
	}

	fiatAmount := req.Cart.Total
	totalAmount := int64(fiatAmount * 100) // Simple cents conversion

	order, err := uc.createOrder(orderID, orderNumber, req.RestaurantID, orderItems, totalAmount, fiatAmount, req.PaymentMethod)
	if err != nil {
		return nil, err
	}

	if err := uc.orderRepo.Save(order); err != nil {
		return nil, err
	}

	if err := uc.processPayment(ctx, order, req.PaymentMethod); err != nil {
		return nil, err
	}

	uc.publishOrderCreatedEvent(ctx, order)
	uc.logger.Info("Order created", "orderID", order.ID, "amount", order.FiatAmount)

	return &CreateOrderResponse{
		OrderID:     order.ID,
		OrderNumber: order.OrderNumber,
	}, nil
}

func (uc *CreateOrderUseCase) generateOrderID() domain.OrderID {
	return domain.OrderID(fmt.Sprintf("ord_%d", time.Now().UnixNano()))
}

func (uc *CreateOrderUseCase) generateOrderNumber() domain.OrderNumber {
	return domain.OrderNumber(fmt.Sprintf("%04d", rand.Intn(10000)))
}

func (uc *CreateOrderUseCase) createOrderItems(cartItems []cart.CartItem, orderID domain.OrderID) ([]domain.OrderItem, error) {
	var orderItems []domain.OrderItem
	for _, item := range cartItems {
		orderItemID := domain.OrderItemID(fmt.Sprintf("oi_%d_%s", time.Now().UnixNano(), item.ItemID))
		orderItem, err := domain.NewOrderItem(
			orderItemID,
			orderID,
			item.ItemID,
			item.Name,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, *orderItem)
	}
	return orderItems, nil
}

func (uc *CreateOrderUseCase) createOrder(orderID domain.OrderID, orderNumber domain.OrderNumber, restaurantID domain.RestaurantID, orderItems []domain.OrderItem, totalAmount int64, fiatAmount float64, paymentMethod domain.PaymentMethodType) (*domain.Order, error) {
	order, err := domain.NewOrder(
		orderID,
		orderNumber,
		restaurantID,
		orderItems,
		totalAmount,
		paymentMethod,
	)
	if err != nil {
		return nil, err
	}
	order.FiatAmount = fiatAmount
	return order, nil
}

func (uc *CreateOrderUseCase) processPayment(ctx context.Context, order *domain.Order, paymentMethod domain.PaymentMethodType) error {
	if uc.paymentMethod.GetPaymentMethodType() != paymentMethod {
		return nil
	}
	payment, err := uc.paymentMethod.ProcessPayment(ctx, order)
	if err != nil {
		return err
	}
	return uc.paymentRepo.Save(payment)
}

func (uc *CreateOrderUseCase) publishOrderCreatedEvent(ctx context.Context, order *domain.Order) {
	event := domain.OrderCreated{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		OrderNumber:  order.OrderNumber,
		TotalAmount:  order.TotalAmount,
		CreatedAt:    order.CreatedAt,
	}
	if err := uc.eventBus.Publish(ctx, "OrderCreated", event); err != nil {
		uc.logger.Error("Failed to publish OrderCreated event", "error", err)
	}
}
