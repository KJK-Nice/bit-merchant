package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCartEndpoints(t *testing.T) {
	// Setup
	cartService := cart.NewCartService()
	itemRepo := memory.NewMemoryMenuItemRepository()
	item, _ := domain.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)
	itemRepo.Save(item)

	h := handler.NewCartHandler(cartService, itemRepo)
	e := echo.New()

	t.Run("Add Item", func(t *testing.T) {
		f := make(url.Values)
		f.Set("itemID", "i1")
		f.Set("quantity", "2")
		req := httptest.NewRequest(http.MethodPost, "/cart/add", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("sessionID", "sess_1")

		err := h.AddToCart(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify cart state
		cart := cartService.GetCart("sess_1")
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, 20.0, cart.Total)
	})

	t.Run("Add Item via Query Params", func(t *testing.T) {
		// Test the fallback mechanism for Datastar empty body requests
		req := httptest.NewRequest(http.MethodPost, "/cart/add?itemID=i1&quantity=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON) // Simulate Datastar default content type
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("sessionID", "sess_qp")

		err := h.AddToCart(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		cart := cartService.GetCart("sess_qp")
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, 10.0, cart.Total)
	})

	t.Run("Remove Item", func(t *testing.T) {
		// Pre-populate
		cartService.AddItem("sess_2", item, 1)

		f := make(url.Values)
		f.Set("itemID", "i1")
		req := httptest.NewRequest(http.MethodPost, "/cart/remove", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("sessionID", "sess_2")

		err := h.RemoveFromCart(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		cart := cartService.GetCart("sess_2")
		assert.Len(t, cart.Items, 0)
	})
}
