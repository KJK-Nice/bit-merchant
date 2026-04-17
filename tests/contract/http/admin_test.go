package http_test

import (
	"bitmerchant/internal/common"

	httpMiddleware "bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"

	// Mock Use Cases for Admin Handler
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminEndpoints(t *testing.T) {
	e := echo.New()

	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	createRestUC := restaurantCmd.NewCreateRestaurantHandler(repoRest, nil, nil)
	createCatUC := menuCmd.NewCreateMenuCategoryHandler(repoCat, nil, nil)
	createItemUC := menuCmd.NewCreateMenuItemHandler(repoItem, nil, nil)

	getMenuAdminUC := menuQuery.NewMenuForAdminHandler(repoCat, repoItem, repoRest, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	updateItemUC := menuCmd.NewUpdateMenuItemHandler(repoItem, repoCat, nil, nil)
	updateCategoryUC := menuCmd.NewUpdateMenuCategoryHandler(repoCat, nil, nil)
	toggleAvailUC := menuCmd.NewToggleMenuItemAvailabilityHandler(repoItem, nil, nil)
	reorderCatUC := menuCmd.NewReorderMenuCategoriesHandler(repoCat, nil, nil)
	reorderItemUC := menuCmd.NewReorderMenuItemsHandler(repoItem, repoCat, nil, nil)
	updateTableUC := restaurantCmd.NewUpdateRestaurantTableCountHandler(repoRest, nil, nil)
	generateQRUC := restaurantQuery.NewRestaurantTableQRImageHandler(qr.NewQRCodeService(), "http://localhost", repoRest, nil, nil)

	membershipRepo := memory.NewMemoryMembershipRepository()
	adminHandler := restauranthttp.NewAdminHandler(
		createRestUC,
		createCatUC,
		createItemUC,
		getMenuAdminUC,
		updateItemUC,
		updateCategoryUC,
		toggleAvailUC,
		nil, // uploadPhotoUC
		reorderCatUC,
		reorderItemUC,
		repoItem,
		updateTableUC,
		generateQRUC,
		membershipRepo,
		repoRest,
	)

	// Seed a restaurant for the dashboard context
	restID := common.RestaurantID("restaurant_1")
	rest, _ := restaurant.NewRestaurant(restID, "Test Rest")
	require.NoError(t, repoRest.Save(rest))

	t.Run("GET /admin/dashboard returns dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.Dashboard(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Menu Management")
	})

	t.Run("POST /admin/category creates category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/category", strings.NewReader("name=Starters&displayOrder=1"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.CreateCategory(c))
		// Expect redirect or HTML fragment depending on implementation (Datastar vs standard)
		// Assuming standard form post for now as per MVP
		assert.Equal(t, http.StatusFound, rec.Code)

		// Verify in repo
		cats, _ := repoCat.FindByRestaurantID(restID)
		assert.NotEmpty(t, cats)
		assert.Equal(t, "Starters", cats[0].Name)
	})

	t.Run("POST /admin/item creates item", func(t *testing.T) {
		// Need a category first
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID,
			Name:         "Mains",
			DisplayOrder: 1,
		})

		form := "name=Steak&price=25.00&description=Juicy&categoryID=" + string(cat.ID)
		req := httptest.NewRequest(http.MethodPost, "/admin/item", strings.NewReader(form))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.CreateItem(c))
		assert.Equal(t, http.StatusFound, rec.Code)

		// Verify in repo
		items, _ := repoItem.FindByCategoryID(cat.ID)
		assert.NotEmpty(t, items)
		assert.Equal(t, "Steak", items[0].Name)
	})
}
