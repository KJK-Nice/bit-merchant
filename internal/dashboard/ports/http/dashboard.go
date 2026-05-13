package http

import (
	"bitmerchant/internal/auth/domain/membership"
	commonhttp "bitmerchant/internal/common/http"
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
	getStatsUC     dashboard.RestaurantDashboardStatsHandler
	getHistoryUC   dashboard.PaidOrdersForRestaurantHandler
	getTopItemsUC  dashboard.TopSellingMenuItemsHandler
	getStalledUC   dashboard.StalledOrdersHandler
	toggleOpenUC   restaurantCmd.ToggleRestaurantOpenHandler
	restaurantRepo restaurant.Repository
	membershipRepo membership.Repository
	logger         *slog.Logger
}

const dashboardFlashStatusUpdateFailed = "status_update_failed"

func dashboardFlashMessage(flashCode string) string {
	switch flashCode {
	case dashboardFlashStatusUpdateFailed:
		return "We could not update restaurant status. Please try again."
	default:
		return ""
	}
}

func NewDashboardHandler(
	getStatsUC dashboard.RestaurantDashboardStatsHandler,
	getHistoryUC dashboard.PaidOrdersForRestaurantHandler,
	getTopItemsUC dashboard.TopSellingMenuItemsHandler,
	getStalledUC dashboard.StalledOrdersHandler,
	toggleOpenUC restaurantCmd.ToggleRestaurantOpenHandler,
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
		getStalledUC:   getStalledUC,
		toggleOpenUC:   toggleOpenUC,
		restaurantRepo: restaurantRepo,
		membershipRepo: membershipRepo,
		logger:         logger,
	}
}

func (h *DashboardHandler) Dashboard(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	stats, err := h.getStatsUC.Handle(c.Request().Context(), dashboard.RestaurantDashboardStats{
		RestaurantID: restaurantID,
		Range:        dashboard.DateRangeToday,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load stats: "+err.Error())
	}

	history, err := h.getHistoryUC.Handle(c.Request().Context(), dashboard.PaidOrdersForRestaurant{
		RestaurantID: restaurantID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load history: "+err.Error())
	}

	topItems, err := h.getTopItemsUC.Handle(c.Request().Context(), dashboard.TopSellingMenuItems{
		RestaurantID: restaurantID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load top items: "+err.Error())
	}

	stalled, err := h.getStalledUC.Handle(c.Request().Context(), dashboard.StalledOrders{
		RestaurantID: restaurantID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load stalled orders: "+err.Error())
	}

	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load restaurant: "+err.Error())
	}

	statusErr := dashboardFlashMessage(c.QueryParam("flash"))

	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		h.logger.Error("Dashboard switcher data failed", "error", sErr)
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.DashboardPage(stats, history, topItems, stalled, rest, commonhttp.CSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate, statusErr).Render(c.Request().Context(), c.Response())
}

func (h *DashboardHandler) ToggleOpen(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	closedMsg := c.FormValue("closed_message")
	reopen := c.FormValue("reopening_hours")

	_, err = h.toggleOpenUC.Handle(c.Request().Context(), restaurantCmd.ToggleRestaurantOpen{
		RestaurantID:   restaurantID,
		ClosedMessage:  closedMsg,
		ReopeningHours: reopen,
	})
	if err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?flash="+url.QueryEscape(dashboardFlashStatusUpdateFailed))
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
