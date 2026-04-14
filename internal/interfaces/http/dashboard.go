package http

import (
	"bitmerchant/internal/auth/domain/membership"
	dashboard "bitmerchant/internal/dashboard/app/query"

	"bitmerchant/internal/interfaces/templates"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"net/url"
)

type DashboardHandler struct {
	getStatsUC     *dashboard.GetDashboardStatsUseCase
	getHistoryUC   *dashboard.GetOrderHistoryUseCase
	getTopItemsUC  *dashboard.GetTopSellingItemsUseCase
	toggleOpenUC   *restaurantCmd.ToggleRestaurantOpenUseCase
	restaurantRepo restaurant.Repository
	membershipRepo membership.Repository
	logger         *slog.Logger
}

func NewDashboardHandler(
	getStatsUC *dashboard.GetDashboardStatsUseCase,
	getHistoryUC *dashboard.GetOrderHistoryUseCase,
	getTopItemsUC *dashboard.GetTopSellingItemsUseCase,
	toggleOpenUC *restaurantCmd.ToggleRestaurantOpenUseCase,
	restaurantRepo restaurant.Repository,
	membershipRepo membership.Repository,
	logger *slog.Logger,
) *DashboardHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &DashboardHandler{
		getStatsUC:     getStatsUC,
		getHistoryUC:   getHistoryUC,
		getTopItemsUC:  getTopItemsUC,
		toggleOpenUC:   toggleOpenUC,
		restaurantRepo: restaurantRepo,
		membershipRepo: membershipRepo,
		logger:         logger,
	}
}

func (h *DashboardHandler) Dashboard(c echo.Context) error {
	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	stats, err := h.getStatsUC.Execute(c.Request().Context(), restaurantID, dashboard.DateRangeToday)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load stats: "+err.Error())
	}

	history, err := h.getHistoryUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load history: "+err.Error())
	}

	topItems, err := h.getTopItemsUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load top items: "+err.Error())
	}

	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load restaurant: "+err.Error())
	}

	statusErr := c.QueryParam("error")

	dn, st, ini := LayoutUserStringsFromContext(c)
	label := ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		h.logger.Error("Dashboard switcher data failed", "error", sErr)
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.DashboardPage(stats, history, topItems, rest, getCSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate, statusErr).Render(c.Request().Context(), c.Response())
}

func (h *DashboardHandler) ToggleOpen(c echo.Context) error {
	restaurantID, err := getRestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	closedMsg := c.FormValue("closed_message")
	reopen := c.FormValue("reopening_hours")

	_, err = h.toggleOpenUC.Execute(c.Request().Context(), restaurantID, closedMsg, reopen)
	if err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?error="+url.QueryEscape(err.Error()))
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
