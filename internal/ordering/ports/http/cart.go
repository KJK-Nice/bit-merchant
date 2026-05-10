package http

import (
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/menu/domain/menu"

	"bitmerchant/internal/ordering/app/cart"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// GetItemDetail handles GET /menu/item/:itemID — full item-detail page with
// modifier groups, special instructions, and qty stepper.
func (h *CartHandler) GetItemDetail(c echo.Context) error {
	itemID := c.Param("itemID")
	if itemID == "" {
		return c.String(http.StatusBadRequest, "itemID required")
	}
	item, err := h.itemRepo.FindByID(common.ItemID(itemID))
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}
	restaurantID := c.QueryParam("restaurantID")
	tableLabel := c.QueryParam("table")
	csrfToken := commonhttp.CSRFToken(c)
	return templates.ItemDetailPage(item, restaurantID, tableLabel, csrfToken).Render(c.Request().Context(), c.Response())
}

// AddToCartAndRedirect handles POST /cart/add-redirect — used by the item-detail
// page form. Adds item (with modifiers) to cart then redirects back to the menu.
func (h *CartHandler) AddToCartAndRedirect(c echo.Context) error {
	itemID := c.FormValue("itemID")
	restaurantID := c.FormValue("restaurantID")
	tableLabel := c.FormValue("tableLabel")

	quantityStr := c.FormValue("quantity")
	quantity, _ := strconv.Atoi(quantityStr)
	if quantity <= 0 {
		quantity = 1
	}

	specialInstructions := c.FormValue("specialInstructions")
	sessionID := c.Get("sessionID").(string)

	item, err := h.itemRepo.FindByID(common.ItemID(itemID))
	if err != nil {
		return c.String(http.StatusBadRequest, "Item not found")
	}

	modifiers := parseModifiers(c, item)

	if err := h.cartService.AddItemWithModifiers(sessionID, item, quantity, modifiers, specialInstructions); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	redirectURL := fmt.Sprintf("/menu?restaurantID=%s", strings.ReplaceAll(restaurantID, " ", "+"))
	if tableLabel != "" {
		redirectURL += "&table=" + strings.ReplaceAll(tableLabel, " ", "+")
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// AddToCart handles POST /cart/add
func (h *CartHandler) AddToCart(c echo.Context) error {
	itemID := c.FormValue("itemID")
	if itemID == "" {
		itemID = c.QueryParam("itemID")
	}
	quantityStr := c.FormValue("quantity")
	if quantityStr == "" {
		quantityStr = c.QueryParam("quantity")
	}
	quantity, _ := strconv.Atoi(quantityStr)
	if quantity <= 0 {
		quantity = 1
	}

	specialInstructions := c.FormValue("specialInstructions")

	sessionID := c.Get("sessionID").(string)

	item, err := h.itemRepo.FindByID(common.ItemID(itemID))
	if err != nil {
		return c.String(http.StatusBadRequest, "Item not found")
	}

	modifiers := parseModifiers(c, item)

	if err := h.cartService.AddItemWithModifiers(sessionID, item, quantity, modifiers, specialInstructions); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	updatedCart := h.cartService.GetCart(sessionID)
	return h.writeCartSSE(c, updatedCart)
}

// parseModifiers reads form values named "mod_{groupID}" and converts them to
// CartItemModifier slices. Radio groups produce one modifier; checkbox groups
// can produce multiple (same field name, multiple values).
func parseModifiers(c echo.Context, item *menu.MenuItem) []cart.CartItemModifier {
	if len(item.OptionGroups) == 0 {
		return nil
	}

	// Index option groups and options for fast lookup.
	type optKey struct{ groupID, optionID string }
	type optInfo struct {
		groupName, optName string
		delta              float64
	}
	byOpt := map[optKey]optInfo{}
	for _, g := range item.OptionGroups {
		for _, o := range g.Options {
			byOpt[optKey{g.ID, o.ID}] = optInfo{g.Name, o.Name, o.PriceDelta}
		}
	}

	if err := c.Request().ParseForm(); err != nil {
		return nil
	}

	var mods []cart.CartItemModifier
	for key, vals := range c.Request().Form {
		if !strings.HasPrefix(key, "mod_") {
			continue
		}
		groupID := strings.TrimPrefix(key, "mod_")
		for _, optionID := range vals {
			info, ok := byOpt[optKey{groupID, optionID}]
			if !ok {
				continue
			}
			mods = append(mods, cart.CartItemModifier{
				GroupID:    groupID,
				GroupName:  info.groupName,
				OptionID:   optionID,
				OptionName: info.optName,
				PriceDelta: info.delta,
			})
		}
	}
	return mods
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
