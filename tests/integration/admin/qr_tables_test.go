package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminQR_TableCountAndPrint(t *testing.T) {
	e := echo.New()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()

	restID := domain.RestaurantID("restaurant-qr")
	rest, err := domain.NewRestaurant(restID, "QR Bistro")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	createRestUC := restaurant.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menu.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menu.NewCreateMenuItemUseCase(repoItem)
	getMenuAdminUC := menu.NewGetMenuForAdminUseCase(repoCat, repoItem, repoRest)
	updateItemUC := menu.NewUpdateMenuItemUseCase(repoItem, repoCat)
	updateCategoryUC := menu.NewUpdateMenuCategoryUseCase(repoCat)
	toggleAvailUC := menu.NewToggleMenuItemAvailabilityUseCase(repoItem)
	updateTableUC := restaurant.NewUpdateRestaurantTableCountUseCase(repoRest)
	generateQRUC := restaurant.NewGenerateRestaurantQRUseCase(qr.NewQRCodeService(), "http://localhost", repoRest)

	adminHandler := handler.NewAdminHandler(
		createRestUC, createCatUC, createItemUC, getMenuAdminUC,
		updateItemUC, updateCategoryUC, toggleAvailUC, nil,
		updateTableUC, generateQRUC, membershipRepo, repoRest,
	)

	require.NoError(t, updateTableUC.Execute(context.Background(), restID, 2))

	t.Run("GET admin qr shows management copy", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/qr", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)
		require.NoError(t, adminHandler.GetQRPage(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Table QR codes")
		assert.Contains(t, body, `/admin/qr/table/1`)
		assert.Contains(t, body, `/admin/qr/table/2`)
	})

	t.Run("POST settings updates count and print has three images", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/qr/settings", strings.NewReader("tableCount=3&csrf=x"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)
		require.NoError(t, adminHandler.PostQRSettings(c))
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Contains(t, rec.Header().Get(echo.HeaderLocation), "/admin/qr?saved=1")

		reqP := httptest.NewRequest(http.MethodGet, "/admin/qr/print", nil)
		recP := httptest.NewRecorder()
		cp := e.NewContext(reqP, recP)
		cp.Set(httpMiddleware.ContextRestaurantID, restID)
		require.NoError(t, adminHandler.GetQRPrint(cp))
		assert.Equal(t, http.StatusOK, recP.Code)
		b := recP.Body.String()
		assert.Contains(t, b, `/admin/qr/table/1`)
		assert.Contains(t, b, `/admin/qr/table/2`)
		assert.Contains(t, b, `/admin/qr/table/3`)
	})

	t.Run("PNG out of range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/qr/table/99", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("table")
		c.SetParamValues("99")
		c.Set(httpMiddleware.ContextRestaurantID, restID)
		err := adminHandler.GetQRTablePNG(c)
		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)
	})
}
