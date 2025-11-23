package http

import (
	"net/http"
	"strconv"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates/components"

	"github.com/labstack/echo/v4"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	cartService *cart.CartService
	itemRepo    domain.MenuItemRepository // Need item repo to get item details for AddItem
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartService *cart.CartService, itemRepo domain.MenuItemRepository) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		itemRepo:    itemRepo,
	}
}

// AddToCart handles POST /cart/add
func (h *CartHandler) AddToCart(c echo.Context) error {
	// Datastar sends JSON by default or form encoded?
	// If using @post('/cart/add', { ... }), it sends JSON payload if header not set.
	// However, c.FormValue works for both if Echo binder is used or content type is form.
	// Datastar fetch defaults to JSON.
	// Echo's c.FormValue does NOT parse JSON body automatically without Bind.
	// But we can use a struct to bind.

	type AddToCartRequest struct {
		ItemID   string `json:"itemID" form:"itemID"`
		Quantity string `json:"quantity" form:"quantity"` // Datastar sends strings usually in params object
	}

	req := new(AddToCartRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	// Fallback to query params if body is empty (e.g. Datastar @post without signals)
	if req.ItemID == "" {
		req.ItemID = c.QueryParam("itemID")
	}
	if req.Quantity == "" {
		req.Quantity = c.QueryParam("quantity")
	}

	quantity, _ := strconv.Atoi(req.Quantity)
	if quantity <= 0 {
		quantity = 1
	}

	sessionID := c.Get("sessionID").(string)

	item, err := h.itemRepo.FindByID(domain.ItemID(req.ItemID))
	if err != nil {
		return c.String(http.StatusBadRequest, "Item not found")
	}

	if err := h.cartService.AddItem(sessionID, item, quantity); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return updated CartSummary fragment
	updatedCart := h.cartService.GetCart(sessionID)
	return components.CartSummary(updatedCart, true).Render(c.Request().Context(), c.Response())
}

// RemoveFromCart handles POST /cart/remove
func (h *CartHandler) RemoveFromCart(c echo.Context) error {
	type RemoveFromCartRequest struct {
		ItemID string `json:"itemID" form:"itemID"`
	}
	req := new(RemoveFromCartRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	// Fallback to query params
	if req.ItemID == "" {
		req.ItemID = c.QueryParam("itemID")
	}

	sessionID := c.Get("sessionID").(string)

	if err := h.cartService.RemoveItem(sessionID, domain.ItemID(req.ItemID)); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return updated CartSummary fragment
	updatedCart := h.cartService.GetCart(sessionID)
	return components.CartSummary(updatedCart, true).Render(c.Request().Context(), c.Response())
}

// GetCart handles GET /cart
func (h *CartHandler) GetCart(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	return components.CartSummary(cart, true).Render(c.Request().Context(), c.Response())
}
