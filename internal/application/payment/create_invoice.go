package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/strike"
)

// CreatePaymentInvoiceUseCase creates Lightning invoice for cart
type CreatePaymentInvoiceUseCase struct {
	strikeClient   *strike.Client
	paymentRepo    domain.PaymentRepository
	restaurantRepo domain.RestaurantRepository
	itemRepo       domain.MenuItemRepository
}

// NewCreatePaymentInvoiceUseCase creates a new CreatePaymentInvoiceUseCase
func NewCreatePaymentInvoiceUseCase(
	strikeClient *strike.Client,
	paymentRepo domain.PaymentRepository,
	restaurantRepo domain.RestaurantRepository,
	itemRepo domain.MenuItemRepository,
) *CreatePaymentInvoiceUseCase {
	return &CreatePaymentInvoiceUseCase{
		strikeClient:   strikeClient,
		paymentRepo:    paymentRepo,
		restaurantRepo: restaurantRepo,
		itemRepo:       itemRepo,
	}
}

// CreateInvoiceRequest represents invoice creation request
type CreateInvoiceRequest struct {
	SessionID    string
	RestaurantID domain.RestaurantID
	CartItems    []CartItemRequest
}

// CartItemRequest represents cart item in request
type CartItemRequest struct {
	ItemID   domain.ItemID
	Quantity int
}

// CreateInvoiceResponse represents invoice creation response
type CreateInvoiceResponse struct {
	PaymentID      domain.PaymentID
	InvoiceID      string
	Invoice        string
	AmountFiat     float64
	AmountSatoshis int64
}

// Execute creates Lightning invoice for cart
func (uc *CreatePaymentInvoiceUseCase) Execute(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	restaurant, err := uc.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil {
		return nil, errors.New("restaurant not found")
	}

	if !restaurant.IsOpen {
		return nil, errors.New("restaurant is closed")
	}

	// Calculate total from cart items
	totalFiat := 0.0
	for _, cartItem := range req.CartItems {
		item, err := uc.itemRepo.FindByID(cartItem.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %s not found", cartItem.ItemID)
		}
		if !item.IsAvailable {
			return nil, fmt.Errorf("item %s is not available", cartItem.ItemID)
		}
		totalFiat += item.Price * float64(cartItem.Quantity)
	}

	if totalFiat <= 0 {
		return nil, errors.New("cart total must be greater than 0")
	}

	// Get exchange rate
	rateResp, err := uc.strikeClient.GetExchangeRate(ctx, "USD")
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Calculate satoshis
	amountSatoshis := int64(totalFiat * rateResp.Rate * float64(rateResp.SatoshisPerUnit))

	// Create invoice via Strike
	invoiceReq := strike.CreateInvoiceRequest{
		Amount: strike.Amount{
			Currency: "USD",
			Amount:   fmt.Sprintf("%.2f", totalFiat),
		},
		Description: fmt.Sprintf("Order at %s", restaurant.Name),
	}

	invoiceResp, err := uc.strikeClient.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create payment record
	paymentID := domain.PaymentID(fmt.Sprintf("pay_%d", time.Now().UnixNano()))
	payment, err := domain.NewPayment(
		paymentID,
		req.RestaurantID,
		nil, // OrderID will be set after order creation
		invoiceResp.InvoiceID,
		invoiceResp.Invoice,
		amountSatoshis,
		totalFiat,
		rateResp.Rate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	if err := uc.paymentRepo.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &CreateInvoiceResponse{
		PaymentID:      paymentID,
		InvoiceID:      invoiceResp.InvoiceID,
		Invoice:        invoiceResp.Invoice,
		AmountFiat:     totalFiat,
		AmountSatoshis: amountSatoshis,
	}, nil
}
