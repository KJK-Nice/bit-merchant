package http

import (
	"net/http"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/interfaces/templates"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
)

// ServerHandler renders the front-of-house tablet view (unpaid orders) and
// owns the Mark Paid transition. Cooks are intentionally not authorized for
// this surface.
type ServerHandler struct {
	getUnpaidUC    orderQuery.UnpaidServerOrdersHandler
	markPaidUC     orderCmd.MarkOrderPaidHandler
	restaurantRepo restaurant.Repository
	membershipRepo membership.Repository
}

func NewServerHandler(
	getUnpaidUC orderQuery.UnpaidServerOrdersHandler,
	markPaidUC orderCmd.MarkOrderPaidHandler,
	restaurantRepo restaurant.Repository,
	membershipRepo membership.Repository,
) *ServerHandler {
	return &ServerHandler{
		getUnpaidUC:    getUnpaidUC,
		markPaidUC:     markPaidUC,
		restaurantRepo: restaurantRepo,
		membershipRepo: membershipRepo,
	}
}

func (h *ServerHandler) GetServer(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	orders, err := h.getUnpaidUC.Handle(c.Request().Context(), orderQuery.UnpaidServerOrders{RestaurantID: restaurantID})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.ServerPage(orders, commonhttp.CSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate).Render(c.Request().Context(), c.Response())
}

// MarkPaid handles POST /server/order/:id/mark-paid. Returns an empty 200 — the
// SSE broadcast removes the card from the FOH view.
func (h *ServerHandler) MarkPaid(c echo.Context) error {
	id := c.Param("id")
	if _, err := h.markPaidUC.Handle(c.Request().Context(), orderCmd.MarkOrderPaid{OrderID: common.OrderID(id)}); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
