package http_test

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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Use Cases for Admin Handler
type MockCreateRestaurantUseCase struct {
	mock.Mock
}

func (m *MockCreateRestaurantUseCase) Execute(ctx context.Context, req restaurant.CreateRestaurantRequest) (*domain.Restaurant, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

type MockCreateMenuCategoryUseCase struct {
	mock.Mock
}

func (m *MockCreateMenuCategoryUseCase) Execute(ctx context.Context, req menu.CreateMenuCategoryRequest) (*domain.MenuCategory, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuCategory), args.Error(1)
}

type MockCreateMenuItemUseCase struct {
	mock.Mock
}

func (m *MockCreateMenuItemUseCase) Execute(ctx context.Context, req menu.CreateMenuItemRequest) (*domain.MenuItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
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

	createRestUC := restaurant.NewCreateRestaurantUseCase(repoRest)
	createCatUC := menu.NewCreateMenuCategoryUseCase(repoCat)
	createItemUC := menu.NewCreateMenuItemUseCase(repoItem)

	// We also need GetMenuUseCase to render the dashboard
	getMenuUC := menu.NewGetMenuUseCase(repoCat, repoItem)

	// Initialize Handler (Does not exist yet)
	adminHandler := handler.NewAdminHandler(
		createRestUC,
		createCatUC,
		createItemUC,
		getMenuUC,
		nil, // uploadPhotoUC
		nil, // generateQRUC
	)

	// Seed a restaurant for the dashboard context
	restID := domain.RestaurantID("restaurant_1")
	_, _ = createRestUC.Execute(context.Background(), restaurant.CreateRestaurantRequest{Name: "Test Rest"})

	t.Run("GET /admin/dashboard returns dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, adminHandler.Dashboard(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Dashboard")
	})

	t.Run("POST /admin/category creates category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/category", strings.NewReader("name=Starters&displayOrder=1"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		cat, _ := createCatUC.Execute(context.Background(), menu.CreateMenuCategoryRequest{
			RestaurantID: restID,
			Name:         "Mains",
			DisplayOrder: 1,
		})

		form := "name=Steak&price=25.00&description=Juicy&categoryID=" + string(cat.ID)
		req := httptest.NewRequest(http.MethodPost, "/admin/item", strings.NewReader(form))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, adminHandler.CreateItem(c))
		assert.Equal(t, http.StatusFound, rec.Code)

		// Verify in repo
		items, _ := repoItem.FindByCategoryID(cat.ID)
		assert.NotEmpty(t, items)
		assert.Equal(t, "Steak", items[0].Name)
	})
}
