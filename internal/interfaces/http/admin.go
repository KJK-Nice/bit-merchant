package http

import (
	"net/http"
	"net/url"
	"strconv"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates/admin"

	"github.com/labstack/echo/v4"
)

const adminMenuDashboardPath = "/admin/dashboard"
const adminQRPath = "/admin/qr"

func adminMenuRedirect(errMsg string) string {
	if errMsg == "" {
		return adminMenuDashboardPath
	}
	return adminMenuDashboardPath + "?error=" + url.QueryEscape(errMsg)
}

func adminQRRedirect(errMsg string) string {
	if errMsg == "" {
		return adminQRPath
	}
	return adminQRPath + "?error=" + url.QueryEscape(errMsg)
}

func (h *AdminHandler) restaurantID(c echo.Context) (domain.RestaurantID, error) {
	return getRestaurantIDFromContext(c)
}

func (h *AdminHandler) renderAdminDashboard(c echo.Context, menuData *menu.MenuResponse, restaurantID domain.RestaurantID) error {
	activeLabel := string(restaurantID)
	if menuData != nil && menuData.Restaurant != nil && menuData.Restaurant.Name != "" {
		activeLabel = menuData.Restaurant.Name
	}
	dn, st, ini := LayoutUserStringsFromContext(c)
	switchOpts, activeRole, canCreate, sErr := RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	menuError := c.QueryParam("error")
	return admin.Dashboard(menuData, getCSRFToken(c), activeLabel, dn, st, ini, switchOpts, activeRole, canCreate, menuError).Render(c.Request().Context(), c.Response())
}

type AdminHandler struct {
	createRestaurantUC *restaurant.CreateRestaurantUseCase
	createCategoryUC   *menu.CreateMenuCategoryUseCase
	createItemUC       *menu.CreateMenuItemUseCase
	getMenuAdminUC     *menu.GetMenuForAdminUseCase
	updateItemUC       *menu.UpdateMenuItemUseCase
	updateCategoryUC   *menu.UpdateMenuCategoryUseCase
	toggleItemAvailUC  *menu.ToggleMenuItemAvailabilityUseCase
	uploadPhotoUC       *menu.UploadPhotoUseCase
	updateTableCountUC *restaurant.UpdateRestaurantTableCountUseCase
	generateQRUC       *restaurant.GenerateRestaurantQRUseCase
	membershipRepo     domain.MembershipRepository
	restaurantRepo     domain.RestaurantRepository
}

// NewAdminHandler constructs the admin HTTP handler.
func NewAdminHandler(
	createRestaurantUC *restaurant.CreateRestaurantUseCase,
	createCategoryUC *menu.CreateMenuCategoryUseCase,
	createItemUC *menu.CreateMenuItemUseCase,
	getMenuAdminUC *menu.GetMenuForAdminUseCase,
	updateItemUC *menu.UpdateMenuItemUseCase,
	updateCategoryUC *menu.UpdateMenuCategoryUseCase,
	toggleItemAvailUC *menu.ToggleMenuItemAvailabilityUseCase,
	uploadPhotoUC *menu.UploadPhotoUseCase,
	updateTableCountUC *restaurant.UpdateRestaurantTableCountUseCase,
	generateQRUC *restaurant.GenerateRestaurantQRUseCase,
	membershipRepo domain.MembershipRepository,
	restaurantRepo domain.RestaurantRepository,
) *AdminHandler {
	return &AdminHandler{
		createRestaurantUC: createRestaurantUC,
		createCategoryUC:   createCategoryUC,
		createItemUC:       createItemUC,
		getMenuAdminUC:     getMenuAdminUC,
		updateItemUC:       updateItemUC,
		updateCategoryUC:   updateCategoryUC,
		toggleItemAvailUC:  toggleItemAvailUC,
		uploadPhotoUC:       uploadPhotoUC,
		updateTableCountUC: updateTableCountUC,
		generateQRUC:       generateQRUC,
		membershipRepo:     membershipRepo,
		restaurantRepo:     restaurantRepo,
	}
}

// Dashboard handles GET /admin/dashboard
func (h *AdminHandler) Dashboard(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	menuData, err := h.getMenuAdminUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load dashboard: "+err.Error())
	}

	return h.renderAdminDashboard(c, menuData, restaurantID)
}

// GetMenu redirects legacy /dashboard/menu to the canonical admin menu URL.
func (h *AdminHandler) GetMenu(c echo.Context) error {
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// CreateCategory handles POST /admin/category
func (h *AdminHandler) CreateCategory(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	name := c.FormValue("name")
	displayOrder, _ := strconv.Atoi(c.FormValue("displayOrder"))

	req := menu.CreateMenuCategoryRequest{
		RestaurantID: restaurantID,
		Name:         name,
		DisplayOrder: displayOrder,
	}

	if _, err = h.createCategoryUC.Execute(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// UpdateCategory handles POST /admin/category/:id/update
func (h *AdminHandler) UpdateCategory(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	categoryID := domain.CategoryID(c.Param("id"))
	name := c.FormValue("name")
	displayOrder, _ := strconv.Atoi(c.FormValue("displayOrder"))
	isActive := c.FormValue("isActive") == "on"

	req := menu.UpdateMenuCategoryRequest{
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         name,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
	}
	if err := h.updateCategoryUC.Execute(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// CreateItem handles POST /admin/item
func (h *AdminHandler) CreateItem(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	categoryID := domain.CategoryID(c.FormValue("categoryID"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)

	available := c.FormValue("available") == "on"

	req := menu.CreateMenuItemRequest{
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         name,
		Description:  description,
		Price:        price,
		Available:    available,
	}

	if _, err = h.createItemUC.Execute(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// UpdateItem handles POST /admin/item/:id/update
func (h *AdminHandler) UpdateItem(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := domain.ItemID(c.Param("id"))
	categoryID := domain.CategoryID(c.FormValue("categoryID"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
	available := c.FormValue("available") == "on"

	req := menu.UpdateMenuItemRequest{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		CategoryID:   categoryID,
		Name:         name,
		Description:  description,
		Price:        price,
		Available:    available,
	}
	if err := h.updateItemUC.Execute(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// ToggleItemAvailability handles POST /admin/item/:id/toggle-availability
func (h *AdminHandler) ToggleItemAvailability(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := domain.ItemID(c.Param("id"))
	if err := h.toggleItemAvailUC.Execute(c.Request().Context(), restaurantID, itemID); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// UploadPhoto handles POST /admin/item/:id/photo
func (h *AdminHandler) UploadPhoto(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := domain.ItemID(c.Param("id"))

	file, err := c.FormFile("photo")
	if err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect("Image file required"))
	}
	src, err := file.Open()
	if err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}
	defer src.Close()

	req := menu.UploadPhotoRequest{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		File:         src,
		Filename:     file.Filename,
		ContentType:  file.Header.Get("Content-Type"),
	}

	if _, err = h.uploadPhotoUC.Execute(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(err.Error()))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

func (h *AdminHandler) renderQRPage(c echo.Context, rest *domain.Restaurant, qrError string, saved bool) error {
	dn, st, ini := LayoutUserStringsFromContext(c)
	switchOpts, activeRole, canCreate, sErr := RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	label := ActiveRestaurantLabel(c.Request().Context(), rest.ID, h.restaurantRepo)
	tc := rest.TableCount
	if tc < domain.MinTableCount {
		tc = domain.MinTableCount
	}
	tables := make([]int, tc)
	for i := range tables {
		tables[i] = i + 1
	}
	return admin.QRPage(getCSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate, rest.Name, tables, tc, qrError, saved).Render(c.Request().Context(), c.Response())
}

// GetQRPage handles GET /admin/qr
func (h *AdminHandler) GetQRPage(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load restaurant")
	}
	return h.renderQRPage(c, rest, c.QueryParam("error"), c.QueryParam("saved") == "1")
}

// PostQRSettings handles POST /admin/qr/settings
func (h *AdminHandler) PostQRSettings(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	count, _ := strconv.Atoi(c.FormValue("tableCount"))
	if err := h.updateTableCountUC.Execute(c.Request().Context(), restaurantID, count); err != nil {
		return c.Redirect(http.StatusFound, adminQRRedirect(err.Error()))
	}
	return c.Redirect(http.StatusFound, adminQRPath+"?saved=1")
}

// GetQRTablePNG handles GET /admin/qr/table/:table
func (h *AdminHandler) GetQRTablePNG(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	n, err := strconv.Atoi(c.Param("table"))
	if err != nil || n < domain.MinTableCount {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid table")
	}
	png, err := h.generateQRUC.Execute(c.Request().Context(), restaurantID, n)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Blob(http.StatusOK, "image/png", png)
}

// GetQRPrint handles GET /admin/qr/print
func (h *AdminHandler) GetQRPrint(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load restaurant")
	}
	n := rest.TableCount
	if n < domain.MinTableCount {
		n = domain.MinTableCount
	}
	tables := make([]int, n)
	for i := range tables {
		tables[i] = i + 1
	}
	return admin.QRPrintPage(rest.Name, tables).Render(c.Request().Context(), c.Response())
}

// GetQRCode handles GET /dashboard/qr-code
func (h *AdminHandler) GetQRCode(c echo.Context) error {
	return c.Redirect(http.StatusFound, adminQRPath)
}
