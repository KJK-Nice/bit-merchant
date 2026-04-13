package customer_test

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/app/cart"
	placesCmd "bitmerchant/internal/places/app/command"
	placesQuery "bitmerchant/internal/places/app/query"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMenuThenMyPlacesListsRestaurant(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	orderRepo := memory.NewMemoryOrderRepository()

	r, _ := restaurant.NewRestaurant("visit-test-r", "Stamp Diner")
	require.NoError(t, restRepo.Save(r))
	cat, _ := menu.NewMenuCategory("cat-v", "visit-test-r", "All", 0)
	require.NoError(t, catRepo.Save(cat))

	getMenuUC := menuQuery.NewGetMenuUseCase(catRepo, itemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{})
	cartSvc := cart.NewCartService()
	recordUC := placesCmd.NewRecordMenuVisitUseCase(restRepo, visitRepo)
	menuH := handler.NewMenuHandler(getMenuUC, cartSvc, recordUC)
	listUC := placesQuery.NewListVisitedRestaurantsUseCase(visitRepo, restRepo, orderRepo)
	placesH := handler.NewPlacesHandler(listUC)

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/menu?restaurantID=visit-test-r", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "sess-visit-1")
	require.NoError(t, menuH.GetMenu(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/my-places", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.Set("sessionID", "sess-visit-1")
	require.NoError(t, placesH.GetMyPlaces(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
	body := rec2.Body.String()
	assert.Contains(t, body, "Stamp Diner")
	assert.Contains(t, body, "visit-test-r")
	assert.Contains(t, body, "Open")
}

func TestMyPlacesEmptyStateLinksToHome(t *testing.T) {
	restRepo := memory.NewMemoryRestaurantRepository()
	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	orderRepo := memory.NewMemoryOrderRepository()
	listUC := placesQuery.NewListVisitedRestaurantsUseCase(visitRepo, restRepo, orderRepo)
	placesH := handler.NewPlacesHandler(listUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/my-places", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "sess-empty")

	require.NoError(t, placesH.GetMyPlaces(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "href=\"/\"")
}
