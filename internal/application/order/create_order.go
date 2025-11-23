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
	eventBus      *events.EventBus
	paymentMethod domain.PaymentMethod
	logger        *logging.Logger
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(
	orderRepo domain.OrderRepository,
	paymentRepo domain.PaymentRepository,
	eventBus *events.EventBus,
	paymentMethod domain.PaymentMethod,
	logger *logging.Logger,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		eventBus:      eventBus,
		paymentMethod: paymentMethod,
		logger:        logger,
	}
}

// Execute creates an order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// 1. Create Order Items from Cart
	var orderItems []domain.OrderItem
	var totalAmount int64 // assuming cents/satoshis

	// Generate IDs
	orderID := domain.OrderID(fmt.Sprintf("ord_%d", time.Now().UnixNano()))
	// Simple random order number for MVP
	orderNumber := domain.OrderNumber(fmt.Sprintf("%04d", rand.Intn(10000)))

	for _, item := range req.Cart.Items {
		orderItemID := domain.OrderItemID(fmt.Sprintf("oi_%d_%s", time.Now().UnixNano(), item.ItemID))

		// Create OrderItem
		// Note: domain.NewOrderItem now takes unit price as float64 and calculates subtotal
		orderItem, err := domain.NewOrderItem(
			orderItemID,
			orderID,
			item.ItemID,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, *orderItem)
		// For total amount in int64 (e.g. cents), we need a conversion strategy.
		// Assuming float64 Price is in Dollars/Euros etc.
		// int64 totalAmount is typically for smallest unit.
		// For now, let's assume simple x100 conversion for MVP if currency is USD/EUR.
		// Or if we just store float64 total in Order struct?
		// Order struct has: TotalAmount int64, FiatAmount float64.
		// Let's populate FiatAmount from Cart Total.
		// And TotalAmount.. maybe keep 0 for cash if not using crypto?
		// Or use it for cents.
	}

	fiatAmount := req.Cart.Total
	totalAmount = int64(fiatAmount * 100) // Simple cents conversion

	// 2. Create Order
	order, err := domain.NewOrder(
		orderID,
		orderNumber,
		req.RestaurantID,
		orderItems,
		totalAmount,
		req.PaymentMethod,
	)
	if err != nil {
		return nil, err
	}
	order.FiatAmount = fiatAmount

	// 3. Save Order
	if err := uc.orderRepo.Save(order); err != nil {
		return nil, err
	}

	// 4. Process Payment (Initial Pending State)
	if uc.paymentMethod.GetPaymentMethodType() == req.PaymentMethod {
		payment, err := uc.paymentMethod.ProcessPayment(ctx, order)
		if err != nil {
			return nil, err
		}
		if err := uc.paymentRepo.Save(payment); err != nil {
			return nil, err
		}
	}

	// 5. Publish Event
	event := domain.OrderCreated{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		OrderNumber:  order.OrderNumber,
		TotalAmount:  order.TotalAmount,
		CreatedAt:    order.CreatedAt,
	}
	if err := uc.eventBus.Publish(ctx, "OrderCreated", event); err != nil {
		uc.logger.Error("Failed to publish OrderCreated event", "error", err)
		// Don't fail request, just log
	}

	uc.logger.Info("Order created", "orderID", order.ID, "amount", order.FiatAmount)

	return &CreateOrderResponse{
		OrderID:     order.ID,
		OrderNumber: order.OrderNumber,
	}, nil
}
