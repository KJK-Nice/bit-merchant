package http

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"

	"bitmerchant/internal/interfaces/templates/admin"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strconv"
)

const adminMenuDashboardPath = "/admin/dashboard"
const adminQRPath = "/admin/qr"

const (
	adminFlashMenuActionFailed  = "menu_action_failed"
	adminFlashMenuItemNotFound  = "menu_item_not_found"
	adminFlashMenuImageRequired = "menu_image_required"
	adminFlashQRActionFailed    = "qr_action_failed"
	adminFlashQRSettingsSaved   = "qr_settings_saved"
)

func adminMenuRedirect(flashCode string) string {
	if flashCode == "" {
		return adminMenuDashboardPath
	}
	return adminMenuDashboardPath + "?flash=" + url.QueryEscape(flashCode)
}

func adminQRRedirect(flashCode string) string {
	if flashCode == "" {
		return adminQRPath
	}
	return adminQRPath + "?flash=" + url.QueryEscape(flashCode)
}

func adminMenuFlashMessage(flashCode string) string {
	switch flashCode {
	case adminFlashMenuItemNotFound:
		return "Menu item not found."
	case adminFlashMenuImageRequired:
		return "Please choose an image file before uploading."
	case adminFlashMenuActionFailed:
		return "We could not update the menu. Please try again."
	default:
		return ""
	}
}

func adminQRFlashState(flashCode string) (qrError string, saved bool) {
	switch flashCode {
	case adminFlashQRSettingsSaved:
		return "", true
	case adminFlashQRActionFailed:
		return "We could not update QR settings. Please try again.", false
	default:
		return "", false
	}
}

func (h *AdminHandler) restaurantID(c echo.Context) (common.RestaurantID, error) {
	return commonhttp.RestaurantIDFromContext(c)
}

func (h *AdminHandler) renderAdminDashboard(c echo.Context, menuData *menuQuery.MenuResponse, restaurantID common.RestaurantID) error {
	activeLabel := string(restaurantID)
	if menuData != nil && menuData.Restaurant != nil && menuData.Restaurant.Name != "" {
		activeLabel = menuData.Restaurant.Name
	}
	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	menuError := adminMenuFlashMessage(c.QueryParam("flash"))
	return admin.Dashboard(menuData, commonhttp.CSRFToken(c), activeLabel, dn, st, ini, switchOpts, activeRole, canCreate, menuError).Render(c.Request().Context(), c.Response())
}

type AdminHandler struct {
	createRestaurantUC  restaurantCmd.CreateRestaurantHandler
	createCategoryUC    menuCmd.CreateMenuCategoryHandler
	createItemUC        menuCmd.CreateMenuItemHandler
	getMenuAdminUC      menuQuery.MenuForAdminHandler
	updateItemUC        menuCmd.UpdateMenuItemHandler
	updateCategoryUC    menuCmd.UpdateMenuCategoryHandler
	toggleItemAvailUC   menuCmd.ToggleMenuItemAvailabilityHandler
	uploadPhotoUC       menuCmd.UploadMenuItemPhotoHandler
	reorderCategoriesUC menuCmd.ReorderMenuCategoriesHandler
	reorderItemsUC      menuCmd.ReorderMenuItemsHandler
	itemRepo            menu.ItemRepository
	updateTableCountUC  restaurantCmd.UpdateRestaurantTableCountHandler
	generateQRUC        restaurantQuery.RestaurantTableQRImageHandler
	membershipRepo      membership.Repository
	restaurantRepo      restaurant.Repository
}

// NewAdminHandler constructs the admin HTTP handler.
func NewAdminHandler(
	createRestaurantUC restaurantCmd.CreateRestaurantHandler,
	createCategoryUC menuCmd.CreateMenuCategoryHandler,
	createItemUC menuCmd.CreateMenuItemHandler,
	getMenuAdminUC menuQuery.MenuForAdminHandler,
	updateItemUC menuCmd.UpdateMenuItemHandler,
	updateCategoryUC menuCmd.UpdateMenuCategoryHandler,
	toggleItemAvailUC menuCmd.ToggleMenuItemAvailabilityHandler,
	uploadPhotoUC menuCmd.UploadMenuItemPhotoHandler,
	reorderCategoriesUC menuCmd.ReorderMenuCategoriesHandler,
	reorderItemsUC menuCmd.ReorderMenuItemsHandler,
	itemRepo menu.ItemRepository,
	updateTableCountUC restaurantCmd.UpdateRestaurantTableCountHandler,
	generateQRUC restaurantQuery.RestaurantTableQRImageHandler,
	membershipRepo membership.Repository,
	restaurantRepo restaurant.Repository,
) *AdminHandler {
	return &AdminHandler{
		createRestaurantUC:  createRestaurantUC,
		createCategoryUC:    createCategoryUC,
		createItemUC:        createItemUC,
		getMenuAdminUC:      getMenuAdminUC,
		updateItemUC:        updateItemUC,
		updateCategoryUC:    updateCategoryUC,
		toggleItemAvailUC:   toggleItemAvailUC,
		uploadPhotoUC:       uploadPhotoUC,
		reorderCategoriesUC: reorderCategoriesUC,
		reorderItemsUC:      reorderItemsUC,
		itemRepo:            itemRepo,
		updateTableCountUC:  updateTableCountUC,
		generateQRUC:        generateQRUC,
		membershipRepo:      membershipRepo,
		restaurantRepo:      restaurantRepo,
	}
}

// Dashboard handles GET /admin/dashboard
func (h *AdminHandler) Dashboard(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	menuData, err := h.getMenuAdminUC.Handle(c.Request().Context(), menuQuery.MenuForAdmin{RestaurantID: restaurantID})
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

	req := menuCmd.CreateMenuCategory{
		RestaurantID: restaurantID,
		Name:         name,
		DisplayOrder: displayOrder,
	}

	if _, err = h.createCategoryUC.Handle(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// UpdateCategory handles POST /admin/category/:id/update
func (h *AdminHandler) UpdateCategory(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	categoryID := common.CategoryID(c.Param("id"))
	name := c.FormValue("name")
	displayOrder, _ := strconv.Atoi(c.FormValue("displayOrder"))
	isActive := c.FormValue("isActive") == "on"

	req := menuCmd.UpdateMenuCategory{
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         name,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
	}
	if err := h.updateCategoryUC.Handle(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// CreateItem handles POST /admin/item
func (h *AdminHandler) CreateItem(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	categoryID := common.CategoryID(c.FormValue("categoryID"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)

	available := c.FormValue("available") == "on"
	isVegetarian := c.FormValue("is_vegetarian") == "on"
	isGlutenFree := c.FormValue("is_gluten_free") == "on"
	isSpicy := c.FormValue("is_spicy") == "on"

	currencyCode := h.restaurantCurrencyCode(restaurantID)

	req := menuCmd.CreateMenuItem{
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         name,
		Description:  description,
		Price:        price,
		CurrencyCode: currencyCode,
		Available:    available,
		IsVegetarian: isVegetarian,
		IsGlutenFree: isGlutenFree,
		IsSpicy:      isSpicy,
	}

	if _, err = h.createItemUC.Handle(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// restaurantCurrencyCode returns the restaurant's base currency code, or
// "USD" if the restaurant cannot be loaded (the menu commands re-validate
// via money.Parse so an empty/unknown code is rejected upstream too).
func (h *AdminHandler) restaurantCurrencyCode(restaurantID common.RestaurantID) string {
	if h.restaurantRepo == nil {
		return ""
	}
	rest, err := h.restaurantRepo.FindByID(restaurantID)
	if err != nil || rest == nil {
		return ""
	}
	return rest.BaseCurrency.Code
}

// UpdateItem handles POST /admin/item/:id/update
func (h *AdminHandler) UpdateItem(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := common.ItemID(c.Param("id"))
	categoryID := common.CategoryID(c.FormValue("categoryID"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
	available := c.FormValue("available") == "on"
	isVegetarian := c.FormValue("is_vegetarian") == "on"
	isGlutenFree := c.FormValue("is_gluten_free") == "on"
	isSpicy := c.FormValue("is_spicy") == "on"

	req := menuCmd.UpdateMenuItem{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		CategoryID:   categoryID,
		Name:         name,
		Description:  description,
		Price:        price,
		Available:    available,
		IsVegetarian: isVegetarian,
		IsGlutenFree: isGlutenFree,
		IsSpicy:      isSpicy,
	}
	if err := h.updateItemUC.Handle(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// ToggleItemAvailability handles POST /admin/item/:id/toggle-availability
func (h *AdminHandler) ToggleItemAvailability(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := common.ItemID(c.Param("id"))
	if err := h.toggleItemAvailUC.Handle(c.Request().Context(), menuCmd.ToggleMenuItemAvailability{
		RestaurantID: restaurantID,
		ItemID:       itemID,
	}); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}
	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// UploadPhoto handles POST /admin/item/:id/photo
func (h *AdminHandler) UploadPhoto(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := common.ItemID(c.Param("id"))

	existing, err := h.itemRepo.FindByID(itemID)
	if err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuItemNotFound))
	}
	if existing.RestaurantID != restaurantID {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuItemNotFound))
	}

	file, err := c.FormFile("photo")
	if err != nil {
		if existing.PhotoURL != "" {
			return c.Redirect(http.StatusFound, adminMenuDashboardPath)
		}
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuImageRequired))
	}
	src, err := file.Open()
	if err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}
	defer src.Close()

	if h.uploadPhotoUC == nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}

	req := menuCmd.UploadMenuItemPhoto{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		File:         src,
		Filename:     file.Filename,
		ContentType:  file.Header.Get("Content-Type"),
	}

	if _, err = h.uploadPhotoUC.Handle(c.Request().Context(), req); err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuActionFailed))
	}

	return c.Redirect(http.StatusFound, adminMenuDashboardPath)
}

// PostReorderCategories handles POST /admin/menu/reorder-categories (JSON body: { "categoryIDs": ["id1", ...] }).
func (h *AdminHandler) PostReorderCategories(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	var body struct {
		CategoryIDs []string `json:"categoryIDs"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	ordered := make([]common.CategoryID, len(body.CategoryIDs))
	for i, s := range body.CategoryIDs {
		ordered[i] = common.CategoryID(s)
	}
	if err := h.reorderCategoriesUC.Handle(c.Request().Context(), menuCmd.ReorderMenuCategories{
		RestaurantID:       restaurantID,
		OrderedCategoryIDs: ordered,
	}); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// PostReorderItems handles POST /admin/menu/reorder-items (JSON body: { "categoryID": "...", "itemIDs": ["id1", ...] }).
func (h *AdminHandler) PostReorderItems(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	var body struct {
		CategoryID string   `json:"categoryID"`
		ItemIDs    []string `json:"itemIDs"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	ordered := make([]common.ItemID, len(body.ItemIDs))
	for i, s := range body.ItemIDs {
		ordered[i] = common.ItemID(s)
	}
	if err := h.reorderItemsUC.Handle(c.Request().Context(), menuCmd.ReorderMenuItems{
		RestaurantID:   restaurantID,
		CategoryID:     common.CategoryID(body.CategoryID),
		OrderedItemIDs: ordered,
	}); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminHandler) renderQRPage(c echo.Context, rest *restaurant.Restaurant, qrError string, saved bool) error {
	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	label := commonhttp.ActiveRestaurantLabel(c.Request().Context(), rest.ID, h.restaurantRepo)
	tc := rest.TableCount
	if tc < restaurant.MinTableCount {
		tc = restaurant.MinTableCount
	}
	tables := make([]int, tc)
	for i := range tables {
		tables[i] = i + 1
	}
	return admin.QRPage(commonhttp.CSRFToken(c), label, dn, st, ini, switchOpts, activeRole, canCreate, rest.Name, tables, tc, qrError, saved).Render(c.Request().Context(), c.Response())
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
	qrError, saved := adminQRFlashState(c.QueryParam("flash"))
	return h.renderQRPage(c, rest, qrError, saved)
}

// PostQRSettings handles POST /admin/qr/settings
func (h *AdminHandler) PostQRSettings(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	count, _ := strconv.Atoi(c.FormValue("tableCount"))
	if err := h.updateTableCountUC.Handle(c.Request().Context(), restaurantCmd.UpdateRestaurantTableCount{
		RestaurantID: restaurantID,
		TableCount:   count,
	}); err != nil {
		return c.Redirect(http.StatusFound, adminQRRedirect(adminFlashQRActionFailed))
	}
	return c.Redirect(http.StatusFound, adminQRRedirect(adminFlashQRSettingsSaved))
}

// GetQRTablePNG handles GET /admin/qr/table/:table
func (h *AdminHandler) GetQRTablePNG(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	n, err := strconv.Atoi(c.Param("table"))
	if err != nil || n < restaurant.MinTableCount {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid table")
	}
	png, err := h.generateQRUC.Handle(c.Request().Context(), restaurantQuery.RestaurantTableQRImage{
		RestaurantID: restaurantID,
		TableNumber:  n,
	})
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
	if n < restaurant.MinTableCount {
		n = restaurant.MinTableCount
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
