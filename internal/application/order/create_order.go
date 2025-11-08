package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bitmerchant/internal/domain"
)

// CreateOrderUseCase creates order from paid payment
type CreateOrderUseCase struct {
	orderRepo   domain.OrderRepository
	paymentRepo domain.PaymentRepository
	itemRepo    domain.MenuItemRepository
	eventBus    interface {
		Publish(ctx context.Context, topic string, event interface{}) error
	}
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(
	orderRepo domain.OrderRepository,
	paymentRepo domain.PaymentRepository,
	itemRepo domain.MenuItemRepository,
	eventBus interface {
		Publish(ctx context.Context, topic string, event interface{}) error
	},
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		itemRepo:    itemRepo,
		eventBus:    eventBus,
	}
}

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	PaymentID domain.PaymentID
	Items     []OrderItemRequest
}

// OrderItemRequest represents order item in request
type OrderItemRequest struct {
	MenuItemID domain.ItemID
	Quantity   int
}

// CreateOrderResponse represents order creation response
type CreateOrderResponse struct {
	OrderID     domain.OrderID
	OrderNumber domain.OrderNumber
}

// Execute creates order from paid payment
func (uc *CreateOrderUseCase) Execute(req CreateOrderRequest) (*CreateOrderResponse, error) {
	// Get payment
	payment, err := uc.paymentRepo.FindByID(req.PaymentID)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	if payment.Status != domain.PaymentStatusPaid {
		return nil, errors.New("payment is not paid")
	}

	// Generate order number
	orderNumber := domain.OrderNumber(fmt.Sprintf("ORD-%03d", time.Now().Unix()%1000))
	orderID := domain.OrderID(fmt.Sprintf("ord_%d", time.Now().UnixNano()))

	// Build order items
	orderItems := make([]domain.OrderItem, 0, len(req.Items))
	for _, itemReq := range req.Items {
		menuItem, err := uc.itemRepo.FindByID(itemReq.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("menu item %s not found", itemReq.MenuItemID)
		}

		// Calculate satoshis per unit
		unitPriceSatoshis := int64(menuItem.Price * payment.ExchangeRate * 100000000) // Convert to satoshis

		orderItemID := domain.OrderItemID(fmt.Sprintf("oi_%d_%d", time.Now().UnixNano(), len(orderItems)))
		orderItem, err := domain.NewOrderItem(
			orderItemID,
			orderID,
			itemReq.MenuItemID,
			itemReq.Quantity,
			menuItem.Price,
			unitPriceSatoshis,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		orderItems = append(orderItems, *orderItem)
	}

	// Create order
	order, err := domain.NewOrder(
		orderID,
		orderNumber,
		payment.RestaurantID,
		orderItems,
		payment.AmountSatoshis,
		payment.AmountFiat,
		payment.InvoiceID,
		payment.Invoice,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Save order
	if err := uc.orderRepo.Save(order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Update payment with order ID if not already set
	if payment.OrderID == nil {
		payment.OrderID = &orderID
		_ = uc.paymentRepo.Update(payment)
	}

	// Publish OrderPaid event
	orderPaidEvent := domain.OrderPaid{
		OrderID:      orderID,
		RestaurantID: payment.RestaurantID,
		OrderNumber:  orderNumber,
		TotalAmount:  payment.AmountSatoshis,
		CreatedAt:    time.Now(),
	}
	_ = uc.eventBus.Publish(context.Background(), "OrderPaid", orderPaidEvent)

	return &CreateOrderResponse{
		OrderID:     orderID,
		OrderNumber: orderNumber,
	}, nil
}
