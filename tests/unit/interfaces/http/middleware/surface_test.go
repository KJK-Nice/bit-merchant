package middleware_test

import (
	httpMiddleware "bitmerchant/internal/common/http/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSurfaceRoutingMiddleware_CanonicalRedirectsAndAccess(t *testing.T) {
	e := echo.New()
	e.Use(httpMiddleware.SurfaceRoutingMiddleware(httpMiddleware.SurfaceConfig{
		PublicBaseURL:   "https://bitmerchant.com",
		CustomerBaseURL: "https://order.bitmerchant.com",
		MerchantBaseURL: "https://merchant.bitmerchant.com",
	}))

	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "entry") })
	e.GET("/menu", func(c echo.Context) error { return c.String(http.StatusOK, "menu") })
	e.GET("/dashboard", func(c echo.Context) error { return c.String(http.StatusOK, "dashboard") })

	t.Run("customer route on merchant host redirects to customer host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/menu?restaurantID=r1", nil)
		req.Host = "merchant.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "https://order.bitmerchant.com/menu?restaurantID=r1", rec.Header().Get(echo.HeaderLocation))
	})

	t.Run("merchant route on customer host redirects to merchant host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Host = "order.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "https://merchant.bitmerchant.com/dashboard", rec.Header().Get(echo.HeaderLocation))
	})

	t.Run("customer route on public host redirects to customer host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/menu", nil)
		req.Host = "bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "https://order.bitmerchant.com/menu", rec.Header().Get(echo.HeaderLocation))
	})

	t.Run("customer host allows customer route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/menu", nil)
		req.Host = "order.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "menu", rec.Body.String())
	})

	t.Run("merchant host allows merchant route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Host = "merchant.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "dashboard", rec.Body.String())
	})

	t.Run("public host allows root entry route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "entry", rec.Body.String())
	})
}
