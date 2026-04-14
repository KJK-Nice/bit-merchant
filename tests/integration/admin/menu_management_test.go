package admin_test

import (
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"

	// Regression: admin menu HTML must list unavailable items and empty categories (not only public menu).
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

func TestAdminMenuDashboard_ShowsUnavailableItemsAndEmptyCategory(t *testing.T) {
	e := echo.New()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()

	restID := common.RestaurantID("restaurant-admin-ui")
	rest, err := restaurant.NewRestaurant(restID, "Test Bistro")
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
		createRestUC,
		createCatUC,
		createItemUC,
		getMenuAdminUC,
		updateItemUC,
		updateCategoryUC,
		toggleAvailUC,
		nil,
		reorderCatUC,
		reorderItemUC,
		repoItem,
		updateTableUC,
		generateQRUC,
		membershipRepo,
		repoRest,
	)

	_, err = createCatUC.Execute(context.Background(), menuCmd.CreateMenuCategoryRequest{
		RestaurantID: restID,
		Name:         "Soon",
		DisplayOrder: 0,
	})
	require.NoError(t, err)

	cat2, err := createCatUC.Execute(context.Background(), menuCmd.CreateMenuCategoryRequest{
		RestaurantID: restID,
		Name:         "Today",
		DisplayOrder: 1,
	})
	require.NoError(t, err)

	_, err = createItemUC.Execute(context.Background(), menuCmd.CreateMenuItemRequest{
		RestaurantID: restID,
		CategoryID:   cat2.ID,
		Name:         "Special",
		Description:  "",
		Price:        4,
		Available:    false,
	})
	require.NoError(t, err)

	user, err := user.NewUser("owner-menu", "Owner")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(httpMiddleware.ContextRestaurantID, restID)
	ctx.Set(httpMiddleware.ContextAuthUser, user)

	require.NoError(t, adminHandler.Dashboard(ctx))
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "Soon"), "empty category should appear")
	assert.True(t, strings.Contains(body, "Special"))
	assert.True(t, strings.Contains(body, "Unavailable"))
}
