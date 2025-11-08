package http

import (
	"net/http"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	addToCartUseCase      *cart.AddToCartUseCase
	removeFromCartUseCase *cart.RemoveFromCartUseCase
	getCartUseCase        *cart.GetCartUseCase
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(
	addToCartUseCase *cart.AddToCartUseCase,
	removeFromCartUseCase *cart.RemoveFromCartUseCase,
	getCartUseCase *cart.GetCartUseCase,
) *CartHandler {
	return &CartHandler{
		addToCartUseCase:      addToCartUseCase,
		removeFromCartUseCase: removeFromCartUseCase,
		getCartUseCase:        getCartUseCase,
	}
}

// AddToCartRequest represents add to cart request
type AddToCartRequest struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

// RemoveFromCartRequest represents remove from cart request
type RemoveFromCartRequest struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

// AddToCart handles POST /cart/add
func (h *CartHandler) AddToCart(c echo.Context) error {
	var req AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	sessionID := getSessionID(c)
	cartResult, err := h.addToCartUseCase.Execute(sessionID, domain.ItemID(req.ItemID), req.Quantity)
	if err != nil {
		if err.Error() == "item not found" || err.Error() == "item is not available" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "restaurant is closed" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"cart": cartResult})
}

// RemoveFromCart handles POST /cart/remove
func (h *CartHandler) RemoveFromCart(c echo.Context) error {
	var req RemoveFromCartRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	sessionID := getSessionID(c)
	cartResult, err := h.removeFromCartUseCase.Execute(sessionID, domain.ItemID(req.ItemID), req.Quantity)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"cart": cartResult})
}

// GetCart handles GET /cart
func (h *CartHandler) GetCart(c echo.Context) error {
	sessionID := getSessionID(c)
	cartResult := h.getCartUseCase.Execute(sessionID)

	return c.JSON(http.StatusOK, map[string]interface{}{"cart": cartResult})
}

// getSessionID extracts session ID from request (cookie or header)
func getSessionID(c echo.Context) string {
	// Try cookie first
	cookie, err := c.Cookie("session_id")
	if err == nil && cookie != nil {
		return cookie.Value
	}

	// Try header
	sessionID := c.Request().Header.Get("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}

	// Generate new session ID if none exists
	// In production, this would be handled by session middleware
	return c.RealIP() + "_" + c.Request().UserAgent()
}
