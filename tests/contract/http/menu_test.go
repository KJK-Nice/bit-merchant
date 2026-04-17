package http_test

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	menuhttp "bitmerchant/internal/menu/ports/http"
	"bitmerchant/internal/ordering/app/cart"
	placesCmd "bitmerchant/internal/places/app/command"

	// Setup Dependencies
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMenu(t *testing.T) {

	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	cartService := cart.NewCartService()

	// Add some data
	rest, _ := restaurant.NewRestaurant("r1", "Test Restaurant")
	require.NoError(t, restRepo.Save(rest))
	cat, _ := menu.NewMenuCategory("c1", "r1", "Starters", 1)
	require.NoError(t, catRepo.Save(cat))

	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	recordVisitUC := placesCmd.NewRecordMenuVisitHandler(restRepo, visitRepo, nil, nil)
	uc := menuQuery.NewMenuForCustomerHandler(catRepo, itemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	h := menuhttp.NewMenuHandler(uc, cartService, recordVisitUC)

	// Setup Echo
	e := echo.New()
	e.GET("/menu", h.GetMenu)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/menu?restaurantID=r1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-1") // Mock session

	// Execute
	err := h.GetMenu(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Menu")
}

func TestGetMenu_MissingRestaurantIDRedirectsToEntry(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	cartService := cart.NewCartService()

	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	recordVisitUC := placesCmd.NewRecordMenuVisitHandler(restRepo, visitRepo, nil, nil)
	uc := menuQuery.NewMenuForCustomerHandler(catRepo, itemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	h := menuhttp.NewMenuHandler(uc, cartService, recordVisitUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-missing")

	err := h.GetMenu(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/?reason=restaurant_required", rec.Header().Get(echo.HeaderLocation))
}

func TestGetMenu_UnknownRestaurantIDRedirectsToEntry(t *testing.T) {
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	restRepo := memory.NewMemoryRestaurantRepository()
	cartService := cart.NewCartService()

	visitRepo := memory.NewMemorySessionRestaurantVisitRepository()
	recordVisitUC := placesCmd.NewRecordMenuVisitHandler(restRepo, visitRepo, nil, nil)
	uc := menuQuery.NewMenuForCustomerHandler(catRepo, itemRepo, restRepo, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	h := menuhttp.NewMenuHandler(uc, cartService, recordVisitUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/menu?restaurantID=unknown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-unknown")

	err := h.GetMenu(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/?reason=restaurant_not_found", rec.Header().Get(echo.HeaderLocation))
}
