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
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: admin menu HTML must list unavailable items and empty categories (not only public menu).
func TestAdminMenuDashboard_ShowsUnavailableItemsAndEmptyCategory(t *testing.T) {
	e := echo.New()
	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()

	restID := domain.RestaurantID("restaurant-admin-ui")
	rest, err := domain.NewRestaurant(restID, "Test Bistro")
	require.NoError(t, err)
	require.NoError(t, repoRest.Save(rest))

	createRestUC := restaurant.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menu.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menu.NewCreateMenuItemUseCase(repoItem)
	getMenuAdminUC := menu.NewGetMenuForAdminUseCase(repoCat, repoItem, repoRest)
	updateItemUC := menu.NewUpdateMenuItemUseCase(repoItem, repoCat)
	updateCategoryUC := menu.NewUpdateMenuCategoryUseCase(repoCat)
	toggleAvailUC := menu.NewToggleMenuItemAvailabilityUseCase(repoItem)

	adminHandler := handler.NewAdminHandler(
		createRestUC,
		createCatUC,
		createItemUC,
		getMenuAdminUC,
		updateItemUC,
		updateCategoryUC,
		toggleAvailUC,
		nil,
		nil,
		membershipRepo,
		repoRest,
	)

	_, err = createCatUC.Execute(context.Background(), menu.CreateMenuCategoryRequest{
		RestaurantID: restID,
		Name:         "Soon",
		DisplayOrder: 0,
	})
	require.NoError(t, err)

	cat2, err := createCatUC.Execute(context.Background(), menu.CreateMenuCategoryRequest{
		RestaurantID: restID,
		Name:         "Today",
		DisplayOrder: 1,
	})
	require.NoError(t, err)

	_, err = createItemUC.Execute(context.Background(), menu.CreateMenuItemRequest{
		RestaurantID: restID,
		CategoryID:   cat2.ID,
		Name:         "Special",
		Description:  "",
		Price:        4,
		Available:    false,
	})
	require.NoError(t, err)

	user, err := domain.NewUser("owner-menu", "Owner")
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
