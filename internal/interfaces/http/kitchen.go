package http

import (
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/interfaces/templates/components"
	"net/http"

	"github.com/labstack/echo/v4"
)

type KitchenHandler struct {
	getOrdersUC     *kitchen.GetKitchenOrdersUseCase
	markPaidUC      *kitchen.MarkOrderPaidUseCase
	markPreparingUC *kitchen.MarkOrderPreparingUseCase
	markReadyUC     *kitchen.MarkOrderReadyUseCase
}

func NewKitchenHandler(
	getOrdersUC *kitchen.GetKitchenOrdersUseCase,
	markPaidUC *kitchen.MarkOrderPaidUseCase,
	markPreparingUC *kitchen.MarkOrderPreparingUseCase,
	markReadyUC *kitchen.MarkOrderReadyUseCase,
) *KitchenHandler {
	return &KitchenHandler{
		getOrdersUC:     getOrdersUC,
		markPaidUC:      markPaidUC,
		markPreparingUC: markPreparingUC,
		markReadyUC:     markReadyUC,
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
	return templates.KitchenPage(orders, string(restaurantID)).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPaid(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPaidUC.Execute(c.Request().Context(), domain.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkPreparing(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markPreparingUC.Execute(c.Request().Context(), domain.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}

func (h *KitchenHandler) MarkReady(c echo.Context) error {
	id := c.Param("id")
	order, err := h.markReadyUC.Execute(c.Request().Context(), domain.OrderID(id))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return components.OrderCard(order).Render(c.Request().Context(), c.Response())
}
