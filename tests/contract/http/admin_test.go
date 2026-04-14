package http_test

import (
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"

	// Mock Use Cases for Admin Handler
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockCreateRestaurantUseCase struct {
	mock.Mock
}

func (m *MockCreateRestaurantUseCase) Execute(ctx context.Context, req restaurantCmd.CreateRestaurantRequest) (*restaurant.Restaurant, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*restaurant.Restaurant), args.Error(1)
}

type MockCreateMenuCategoryUseCase struct {
	mock.Mock
}

func (m *MockCreateMenuCategoryUseCase) Execute(ctx context.Context, req menuCmd.CreateMenuCategoryRequest) (*menu.MenuCategory, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menu.MenuCategory), args.Error(1)
}

type MockCreateMenuItemUseCase struct {
	mock.Mock
}

func (m *MockCreateMenuItemUseCase) Execute(ctx context.Context, req menuCmd.CreateMenuItemRequest) (*menu.MenuItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menu.MenuItem), args.Error(1)
}

func TestAdminEndpoints(t *testing.T) {
	e := echo.New()

	// Setup Mocks
	// Note: Ideally we test with real use cases + memory repos for "component" tests,
	// or mocks for pure "unit" tests of handler.
	// Given I have real use cases and memory repos, I can use them to be more robust.

	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

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

	membershipRepo := memory.NewMemoryMembershipRepository()
	adminHandler := handler.NewAdminHandler(
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
		cat, _ := createCatUC.Execute(context.Background(), menuCmd.CreateMenuCategoryRequest{
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
