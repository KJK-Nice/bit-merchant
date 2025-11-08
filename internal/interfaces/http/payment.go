package http

import (
	"net/http"

	"bitmerchant/internal/application/payment"
	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	createInvoiceUseCase *payment.CreatePaymentInvoiceUseCase
	checkStatusUseCase   *payment.CheckPaymentStatusUseCase
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(
	createInvoiceUseCase *payment.CreatePaymentInvoiceUseCase,
	checkStatusUseCase *payment.CheckPaymentStatusUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		createInvoiceUseCase: createInvoiceUseCase,
		checkStatusUseCase:   checkStatusUseCase,
	}
}

// CreateInvoiceRequest represents create invoice request
type CreateInvoiceRequest struct {
	Items []CartItemRequest `json:"items"`
}

// CartItemRequest represents cart item in request
type CartItemRequest struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

// CreateInvoice handles POST /payment/create-invoice
func (h *PaymentHandler) CreateInvoice(c echo.Context) error {
	var req CreateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Convert request items
	items := make([]payment.CartItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = payment.CartItemRequest{
			ItemID:   domain.ItemID(item.ItemID),
			Quantity: item.Quantity,
		}
	}

	// Get restaurant ID (v1.0 single tenant)
	restaurantID := domain.RestaurantID("rest_001") // TODO: Get from config

	paymentReq := payment.CreateInvoiceRequest{
		SessionID:    getSessionID(c),
		RestaurantID: restaurantID,
		CartItems:    items,
	}

	result, err := h.createInvoiceUseCase.Execute(c.Request().Context(), paymentReq)
	if err != nil {
		if err.Error() == "restaurant not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "restaurant is closed" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// CheckStatus handles GET /payment/status/:invoiceId
func (h *PaymentHandler) CheckStatus(c echo.Context) error {
	invoiceID := c.Param("invoiceId")
	if invoiceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invoice ID is required"})
	}

	result, err := h.checkStatusUseCase.Execute(c.Request().Context(), invoiceID)
	if err != nil {
		if err.Error() == "payment not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
