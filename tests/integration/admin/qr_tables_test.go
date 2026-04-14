package admin_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminQR_TableCountAndPrint(t *testing.T) {
	e := echo.New()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()

	restID := common.RestaurantID("restaurant-qr")
	rest, err := restaurant.NewRestaurant(restID, "QR Bistro")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	createRestUC := restaurantCmd.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menuCmd.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menuCmd.NewCreateMenuItemUseCase(repoItem)
	getMenuAdminUC := menuQuery.NewGetMenuForAdminUseCase(repoCat, repoItem, repoRest, nil, menuQuery.PhotoSignerConfig{})
	updateItemUC := menuCmd.NewUpdateMenuItemUseCase(repoItem, repoCat)
	updateCategoryUC := menuCmd.NewUpdateMenuCategoryUseCase(repoCat)
	toggleAvailUC := menuCmd.NewToggleMenuItemAvailabilityUseCase(repoItem)
	reorderCatUC := menuCmd.NewReorderMenuCategoriesUseCase(repoCat)
	reorderItemUC := menuCmd.NewReorderMenuItemsUseCase(repoItem, repoCat)
	updateTableUC := restaurantCmd.NewUpdateRestaurantTableCountUseCase(repoRest)
	generateQRUC := restaurantQuery.NewGenerateRestaurantQRUseCase(qr.NewQRCodeService(), "http://localhost", repoRest)

	adminHandler := handler.NewAdminHandler(
		createRestUC, createCatUC, createItemUC, getMenuAdminUC,
		updateItemUC, updateCategoryUC, toggleAvailUC, nil,
		reorderCatUC, reorderItemUC,
		repoItem,
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
