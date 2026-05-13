package http

import (
	"bitmerchant/internal/auth/domain/membership"
	commonhttp "bitmerchant/internal/common/http"
	dashboard "bitmerchant/internal/dashboard/app/query"

	"bitmerchant/internal/interfaces/templates"
	"bitmerchant/internal/ordering/domain/order"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type DashboardHandler struct {
	getStatsUC     dashboard.RestaurantDashboardStatsHandler
	getHistoryUC   dashboard.PaidOrdersForRestaurantHandler
	getTopItemsUC  dashboard.TopSellingMenuItemsHandler
	getStalledUC   dashboard.StalledOrdersHandler
	getByHourUC    dashboard.OrdersByHourHandler
	toggleOpenUC   restaurantCmd.ToggleRestaurantOpenHandler
	pauseUC        restaurantCmd.PauseRestaurantHandler
	restaurantRepo restaurant.Repository
	orderRepo      order.Repository
	membershipRepo membership.Repository
	logger         *slog.Logger
}

const (
	dashboardFlashStatusUpdateFailed = "status_update_failed"
	dashboardFlashPauseFailed        = "pause_failed"
	dashboardFlashPaused             = "paused"
	dashboardFlashResumed            = "resumed"
)

func dashboardFlashMessage(flashCode string) string {
	switch flashCode {
	case dashboardFlashStatusUpdateFailed:
		return "We could not update restaurant status. Please try again."
	case dashboardFlashPauseFailed:
		return "We could not pause the restaurant. Please try again."
	default:
		return ""
	}
}

// dashboardHistoryPageSize caps each Recent Orders page.
const dashboardHistoryPageSize = 10

func NewDashboardHandler(
	getStatsUC dashboard.RestaurantDashboardStatsHandler,
	getHistoryUC dashboard.PaidOrdersForRestaurantHandler,
	getTopItemsUC dashboard.TopSellingMenuItemsHandler,
	getStalledUC dashboard.StalledOrdersHandler,
	getByHourUC dashboard.OrdersByHourHandler,
	toggleOpenUC restaurantCmd.ToggleRestaurantOpenHandler,
	pauseUC restaurantCmd.PauseRestaurantHandler,
	restaurantRepo restaurant.Repository,
	orderRepo order.Repository,
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
		getByHourUC:    getByHourUC,
		toggleOpenUC:   toggleOpenUC,
		pauseUC:        pauseUC,
		restaurantRepo: restaurantRepo,
		orderRepo:      orderRepo,
		membershipRepo: membershipRepo,
		logger:         logger,
	}
}

// parseRange normalises the ?range= query param to today/week/month, with
// today as the safe default.
func parseRange(raw string) dashboard.DateRange {
	switch raw {
	case string(dashboard.DateRangeWeek), "7d":
		return dashboard.DateRangeWeek
	case string(dashboard.DateRangeMonth), "30d":
		return dashboard.DateRangeMonth
	default:
		return dashboard.DateRangeToday
	}
}

func (h *DashboardHandler) Dashboard(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	rng := parseRange(c.QueryParam("range"))
	statusFilter := c.QueryParam("status")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	stats, err := h.getStatsUC.Handle(c.Request().Context(), dashboard.RestaurantDashboardStats{
		RestaurantID: restaurantID,
		Range:        rng,
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
	filtered := filterAndPage(history, statusFilter, page, dashboardHistoryPageSize)

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

	hourly, err := h.getByHourUC.Handle(c.Request().Context(), dashboard.OrdersByHour{
		RestaurantID: restaurantID,
		Range:        rng,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load hourly stats: "+err.Error())
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
	view := templates.DashboardView{
		Stats:        stats,
		History:      filtered.Items,
		TotalHistory: filtered.Total,
		Page:         page,
		PageSize:     dashboardHistoryPageSize,
		StatusFilter: statusFilter,
		TopItems:     topItems,
		Stalled:      stalled,
		Hourly:       hourly,
		Range:        rng,
		Restaurant:   rest,
		Now:          time.Now(),
		CSRFToken:    commonhttp.CSRFToken(c),
		ActiveLabel:  label,
		DisplayName:  dn,
		Subtitle:     st,
		Initials:     ini,
		Switcher:     switchOpts,
		ActiveRole:   activeRole,
		CanCreate:    canCreate,
		StatusError:  statusErr,
	}
	return templates.DashboardPage(view).Render(c.Request().Context(), c.Response())
}

// HistoryFilterResult holds the filtered + paginated slice plus the
// pre-pagination total so the template can render a "Page X of Y" counter.
type HistoryFilterResult struct {
	Items []*order.Order
	Total int
}

func filterAndPage(orders []*order.Order, status string, page, pageSize int) HistoryFilterResult {
	if pageSize <= 0 {
		pageSize = dashboardHistoryPageSize
	}
	filtered := make([]*order.Order, 0, len(orders))
	for _, o := range orders {
		if status != "" && string(o.FulfillmentStatus) != status {
			continue
		}
		filtered = append(filtered, o)
	}
	total := len(filtered)
	startIdx := (page - 1) * pageSize
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= total {
		return HistoryFilterResult{Items: nil, Total: total}
	}
	endIdx := startIdx + pageSize
	if endIdx > total {
		endIdx = total
	}
	return HistoryFilterResult{Items: filtered[startIdx:endIdx], Total: total}
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

// OrderDetail renders the owner-facing order detail panel for the click
// target from the Recent Orders table. Scoped to the active restaurant —
// querying another restaurant's order returns 404.
func (h *DashboardHandler) OrderDetail(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		return c.String(http.StatusBadRequest, "Order number required")
	}
	o, err := h.orderRepo.FindByOrderNumber(restaurantID, orderNumber)
	if err != nil {
		if err.Error() == "order not found" {
			return c.String(http.StatusNotFound, "Order not found")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	return templates.DashboardOrderDetail(o, rest, label, dn, st, ini, commonhttp.CSRFToken(c), switchOpts, activeRole, canCreate).Render(c.Request().Context(), c.Response())
}

// Pause applies a quick-pause window (15/30/60 minutes are typical). A
// non-positive minutes value resumes any active pause immediately.
func (h *DashboardHandler) Pause(c echo.Context) error {
	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	minutes, _ := strconv.Atoi(c.FormValue("minutes"))
	cmd := restaurantCmd.PauseRestaurant{
		RestaurantID: restaurantID,
		Duration:     time.Duration(minutes) * time.Minute,
	}
	if err := h.pauseUC.Handle(c.Request().Context(), cmd); err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?flash="+url.QueryEscape(dashboardFlashPauseFailed))
	}
	if minutes <= 0 {
		return c.Redirect(http.StatusFound, "/dashboard?flash="+url.QueryEscape(dashboardFlashResumed))
	}
	return c.Redirect(http.StatusFound, "/dashboard?flash="+url.QueryEscape(dashboardFlashPaused))
}
