package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetMenu(t *testing.T) {
	// Setup Dependencies
	catRepo := memory.NewMemoryMenuCategoryRepository()
	itemRepo := memory.NewMemoryMenuItemRepository()
	cartService := cart.NewCartService()

	// Add some data
	cat, _ := domain.NewMenuCategory("c1", "r1", "Starters", 1)
	catRepo.Save(cat)

	uc := menu.NewGetMenuUseCase(catRepo, itemRepo)
	h := handler.NewMenuHandler(uc, cartService)

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
