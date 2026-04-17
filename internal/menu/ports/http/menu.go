package http

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/interfaces/templates"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/ordering/app/cart"
	placesCmd "bitmerchant/internal/places/app/command"

	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"net/url"
)

type MenuHandler struct {
	getMenu       menuQuery.MenuForCustomerHandler
	cartService   *cart.CartService
	recordVisitUC placesCmd.RecordMenuVisitHandler
}

func NewMenuHandler(getMenu menuQuery.MenuForCustomerHandler, cartService *cart.CartService, recordVisitUC placesCmd.RecordMenuVisitHandler) *MenuHandler {
	return &MenuHandler{
		getMenu:       getMenu,
		cartService:   cartService,
		recordVisitUC: recordVisitUC,
	}
}

func (h *MenuHandler) GetMenu(c echo.Context) error {
	restaurantID := c.QueryParam("restaurantID")
	if restaurantID == "" {
		return c.Redirect(http.StatusFound, "/?reason="+url.QueryEscape("restaurant_required"))
	}

	// Get Menu Data
	menuData, err := h.getMenu.Handle(c.Request().Context(), menuQuery.MenuForCustomer{RestaurantID: common.RestaurantID(restaurantID)})
	if err != nil {
		if err.Error() == "restaurant not found" {
			return c.Redirect(http.StatusFound, "/?reason="+url.QueryEscape("restaurant_not_found"))
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get Cart Data
	sessionID := c.Get("sessionID").(string)
	cart := h.cartService.GetCart(sessionID)

	if h.recordVisitUC != nil {
		if err := h.recordVisitUC.Handle(c.Request().Context(), placesCmd.RecordMenuVisit{
			SessionID:    sessionID,
			RestaurantID: common.RestaurantID(restaurantID),
		}); err != nil {
			slog.Warn("record menu visit failed", "error", err, "restaurantID", restaurantID)
		}
	}

	tableLabel := c.QueryParam("table")

	// Prevent caching so back button always fetches fresh state (updated cart)
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	return templates.MenuPage(menuData, cart, tableLabel).Render(c.Request().Context(), c.Response())
}
