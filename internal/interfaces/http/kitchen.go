package http

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/interfaces/templates/components"
	kitchenCmd "bitmerchant/internal/ordering/app/command"
	kitchenQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"net/http"
)

type KitchenHandler struct {
	getOrdersUC     *kitchenQuery.GetKitchenOrdersUseCase
	markPaidUC      *kitchenCmd.MarkOrderPaidUseCase
	markPreparingUC *kitchenCmd.MarkOrderPreparingUseCase
	markReadyUC     *kitchenCmd.MarkOrderReadyUseCase
	restaurantRepo  restaurant.Repository
	membershipRepo  membership.Repository
}

func NewKitchenHandler(
	getOrdersUC *kitchenQuery.GetKitchenOrdersUseCase,
	markPaidUC *kitchenCmd.MarkOrderPaidUseCase,
	markPreparingUC *kitchenCmd.MarkOrderPreparingUseCase,
	markReadyUC *kitchenCmd.MarkOrderReadyUseCase,
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
	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	orders, err := h.getOrdersUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	dn, st, ini := LayoutUserStringsFromContext(c)
	label := ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.KitchenPage(orders, getCSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPaid(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPaidUC.Execute(c.Request().Context(), common.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPreparing(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPreparingUC.Execute(c.Request().Context(), common.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkReady(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markReadyUC.Execute(c.Request().Context(), common.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}
