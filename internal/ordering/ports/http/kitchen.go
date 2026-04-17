package http

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/interfaces/templates/components"
	orderCmd "bitmerchant/internal/ordering/app/command"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"net/http"
)

type KitchenHandler struct {
	getOrdersUC     orderQuery.ActiveKitchenOrdersHandler
	markPaidUC      orderCmd.MarkOrderPaidHandler
	markPreparingUC orderCmd.MarkOrderPreparingHandler
	markReadyUC     orderCmd.MarkOrderReadyHandler
	restaurantRepo  restaurant.Repository
	membershipRepo  membership.Repository
}

func NewKitchenHandler(
	getOrdersUC orderQuery.ActiveKitchenOrdersHandler,
	markPaidUC orderCmd.MarkOrderPaidHandler,
	markPreparingUC orderCmd.MarkOrderPreparingHandler,
	markReadyUC orderCmd.MarkOrderReadyHandler,
	restaurantRepo restaurant.Repository,
	membershipRepo membership.Repository,
) *KitchenHandler {
	return &KitchenHandler{
		getOrdersUC:     getOrdersUC,
		markPaidUC:      markPaidUC,
		markPreparingUC: markPreparingUC,
		markReadyUC:     markReadyUC,
		restaurantRepo:  restaurantRepo,
		membershipRepo:  membershipRepo,
	}
}

func (h *KitchenHandler) GetKitchen(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	orders, err := h.getOrdersUC.Handle(c.Request().Context(), orderQuery.ActiveKitchenOrders{RestaurantID: restaurantID})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.KitchenPage(orders, commonhttp.CSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPaid(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPaidUC.Handle(c.Request().Context(), orderCmd.MarkOrderPaid{OrderID: common.OrderID(id)})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPreparing(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPreparingUC.Handle(c.Request().Context(), orderCmd.MarkOrderPreparing{OrderID: common.OrderID(id)})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkReady(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markReadyUC.Handle(c.Request().Context(), orderCmd.MarkOrderReady{OrderID: common.OrderID(id)})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}
