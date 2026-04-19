package http

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/menu/domain/menu"

	"bitmerchant/internal/ordering/app/cart"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type CartHandler struct {
	cartService *cart.CartService
	itemRepo    menu.ItemRepository
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartService *cart.CartService, itemRepo menu.ItemRepository) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		itemRepo:    itemRepo,
	}
}

// writeCartSSE writes a Datastar SSE response with updated cart fragments + per-item qty signals.
// zeroedIDs are item IDs that were just removed; they are emitted with qty=0 so the menu CTA resets.
func (h *CartHandler) writeCartSSE(c echo.Context, updatedCart *cart.Cart, zeroedIDs ...common.ItemID) error {
	ctx := c.Request().Context()
	w := c.Response()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Build HTML fragments.
	var summaryBuf bytes.Buffer
	if err := components.CartSummary(updatedCart, true).Render(ctx, &summaryBuf); err != nil {
		return err
	}
	var floatBuf bytes.Buffer
	if err := components.CartFloatingButton(updatedCart).Render(ctx, &floatBuf); err != nil {
		return err
	}

	// Emit element patches.
	fmt.Fprintf(w, "event: datastar-patch-elements\ndata: elements %s\n\n", summaryBuf.String())
	fmt.Fprintf(w, "event: datastar-patch-elements\ndata: elements %s\n\n", floatBuf.String())

	// Emit per-item qty signal patch so the menu page CTAs can react.
	qtyMap := map[string]int{}
	for _, item := range updatedCart.Items {
		qtyMap[string(item.ItemID)] = item.Quantity
	}
	// Explicitly zero out removed items so the "Add to Cart" CTA reappears.
	for _, id := range zeroedIDs {
		if _, stillInCart := qtyMap[string(id)]; !stillInCart {
			qtyMap[string(id)] = 0
		}
	}
	signalBytes, err := json.Marshal(map[string]any{"cartItemQty": qtyMap})
	if err == nil {
		fmt.Fprintf(w, "event: datastar-patch-signals\ndata: signals %s\n\n", string(signalBytes))
	}

	w.Flush()
	return nil
}

// AddToCart handles POST /cart/add
func (h *CartHandler) AddToCart(c echo.Context) error {
	type AddToCartRequest struct {
		ItemID   string `json:"itemID" form:"itemID"`
		Quantity string `json:"quantity" form:"quantity"`
	}

	req := new(AddToCartRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

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

	item, err := h.itemRepo.FindByID(common.ItemID(req.ItemID))
	if err != nil {
		return c.String(http.StatusBadRequest, "Item not found")
	}

	if err := h.cartService.AddItem(sessionID, item, quantity); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	updatedCart := h.cartService.GetCart(sessionID)
	return h.writeCartSSE(c, updatedCart)
}

// DecrementFromCart handles POST /cart/decrement — reduces qty by 1, removes item at 0.
func (h *CartHandler) DecrementFromCart(c echo.Context) error {
	type DecrementRequest struct {
		ItemID string `json:"itemID" form:"itemID"`
	}
	req := new(DecrementRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if req.ItemID == "" {
		req.ItemID = c.QueryParam("itemID")
	}

	sessionID := c.Get("sessionID").(string)

	if err := h.cartService.DecrementItem(sessionID, common.ItemID(req.ItemID)); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	updatedCart := h.cartService.GetCart(sessionID)
	return h.writeCartSSE(c, updatedCart, common.ItemID(req.ItemID))
}

// RemoveFromCart handles POST /cart/remove — removes the entire line regardless of qty.
func (h *CartHandler) RemoveFromCart(c echo.Context) error {
	type RemoveFromCartRequest struct {
		ItemID string `json:"itemID" form:"itemID"`
	}
	req := new(RemoveFromCartRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}

	if req.ItemID == "" {
		req.ItemID = c.QueryParam("itemID")
	}

	sessionID := c.Get("sessionID").(string)

	if err := h.cartService.RemoveItem(sessionID, common.ItemID(req.ItemID)); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	updatedCart := h.cartService.GetCart(sessionID)
	return h.writeCartSSE(c, updatedCart, common.ItemID(req.ItemID))
}

// GetCart handles GET /cart — returns the cart summary fragment for Datastar to patch.
func (h *CartHandler) GetCart(c echo.Context) error {
	sessionID := c.Get("sessionID").(string)
	updatedCart := h.cartService.GetCart(sessionID)
	return h.writeCartSSE(c, updatedCart)
}
