package http

import (
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/interfaces/templates/admin"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	createRestaurantUC *restaurant.CreateRestaurantUseCase
	createCategoryUC   *menu.CreateMenuCategoryUseCase
	createItemUC       *menu.CreateMenuItemUseCase
	getMenuUC          *menu.GetMenuUseCase
	uploadPhotoUC      *menu.UploadPhotoUseCase
	generateQRUC       *restaurant.GenerateRestaurantQRUseCase
}

func NewAdminHandler(
	createRestaurantUC *restaurant.CreateRestaurantUseCase,
	createCategoryUC *menu.CreateMenuCategoryUseCase,
	createItemUC *menu.CreateMenuItemUseCase,
	getMenuUC *menu.GetMenuUseCase,
	uploadPhotoUC *menu.UploadPhotoUseCase,
	generateQRUC *restaurant.GenerateRestaurantQRUseCase,
) *AdminHandler {
	return &AdminHandler{
		createRestaurantUC: createRestaurantUC,
		createCategoryUC:   createCategoryUC,
		createItemUC:       createItemUC,
		getMenuUC:          getMenuUC,
		uploadPhotoUC:      uploadPhotoUC,
		generateQRUC:       generateQRUC,
	}
}

// Dashboard handles GET /admin/dashboard
func (h *AdminHandler) Dashboard(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP

	menuData, err := h.getMenuUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load dashboard: "+err.Error())
	}

	return admin.Dashboard(menuData).Render(c.Request().Context(), c.Response())
}

// GetMenu handles GET /dashboard/menu
func (h *AdminHandler) GetMenu(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP

	menuData, err := h.getMenuUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load menu: "+err.Error())
	}

	return admin.Dashboard(menuData).Render(c.Request().Context(), c.Response())
}

// CreateCategory handles POST /admin/category
func (h *AdminHandler) CreateCategory(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP
	name := c.FormValue("name")
	displayOrder, _ := strconv.Atoi(c.FormValue("displayOrder"))

	req := menu.CreateMenuCategoryRequest{
		RestaurantID: restaurantID,
		Name:         name,
		DisplayOrder: displayOrder,
	}

	if _, err := h.createCategoryUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}

// CreateMenuCategory handles POST /dashboard/menu/category
func (h *AdminHandler) CreateMenuCategory(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP
	name := c.FormValue("name")
	displayOrder, _ := strconv.Atoi(c.FormValue("displayOrder"))

	req := menu.CreateMenuCategoryRequest{
		RestaurantID: restaurantID,
		Name:         name,
		DisplayOrder: displayOrder,
	}

	if _, err := h.createCategoryUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return HTML fragment for Datastar update
	menuData, err := h.getMenuUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return admin.Dashboard(menuData).Render(c.Request().Context(), c.Response())
}

// CreateItem handles POST /admin/item
func (h *AdminHandler) CreateItem(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP
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

	if _, err := h.createItemUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}

// CreateMenuItem handles POST /dashboard/menu/item
func (h *AdminHandler) CreateMenuItem(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1") // Hardcoded for MVP
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

	if _, err := h.createItemUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return HTML fragment for Datastar update
	menuData, err := h.getMenuUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return admin.Dashboard(menuData).Render(c.Request().Context(), c.Response())
}

// UploadPhoto handles POST /admin/item/:id/photo
func (h *AdminHandler) UploadPhoto(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1")
	itemID := domain.ItemID(c.Param("id"))

	file, err := c.FormFile("photo")
	if err != nil {
		return c.String(http.StatusBadRequest, "Image file required")
	}
	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	req := menu.UploadPhotoRequest{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		File:         src,
		Filename:     file.Filename,
		ContentType:  file.Header.Get("Content-Type"),
	}

	if _, err := h.uploadPhotoUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}

// UploadMenuItemPhoto handles POST /dashboard/menu/item/:id/photo
func (h *AdminHandler) UploadMenuItemPhoto(c echo.Context) error {
	restaurantID := domain.RestaurantID("restaurant_1")
	itemID := domain.ItemID(c.Param("id"))

	file, err := c.FormFile("photo")
	if err != nil {
		return c.String(http.StatusBadRequest, "Image file required")
	}
	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	req := menu.UploadPhotoRequest{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		File:         src,
		Filename:     file.Filename,
		ContentType:  file.Header.Get("Content-Type"),
	}

	if _, err := h.uploadPhotoUC.Execute(c.Request().Context(), req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return HTML fragment for Datastar update
	menuData, err := h.getMenuUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return admin.Dashboard(menuData).Render(c.Request().Context(), c.Response())
}

// GenerateQR handles GET /admin/qr
func (h *AdminHandler) GenerateQR(c echo.Context) error {
	restaurantID := "restaurant_1"

	png, err := h.generateQRUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Blob(http.StatusOK, "image/png", png)
}

// GetQRCode handles GET /dashboard/qr-code
func (h *AdminHandler) GetQRCode(c echo.Context) error {
	restaurantID := "restaurant_1"

	png, err := h.generateQRUC.Execute(c.Request().Context(), restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return HTML page with QR code
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>QR Code - BitMerchant</title>
		</head>
		<body>
			<h1>Restaurant QR Code</h1>
			<img src="data:image/png;base64,%s" alt="QR Code" />
			<p>Scan this QR code to view the menu</p>
		</body>
		</html>
	`, base64.StdEncoding.EncodeToString(png)))
}
