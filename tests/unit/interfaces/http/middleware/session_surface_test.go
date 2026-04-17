package middleware_test

import (
	httpMiddleware "bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/repositories/memory"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionMiddleware_UsesDifferentCookiesPerSurface(t *testing.T) {
	e := echo.New()
	e.Use(httpMiddleware.SurfaceRoutingMiddleware(httpMiddleware.SurfaceConfig{
		PublicBaseURL:   "https://bitmerchant.com",
		CustomerBaseURL: "https://order.bitmerchant.com",
		MerchantBaseURL: "https://merchant.bitmerchant.com",
	}))
	e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
		memory.NewMemorySessionRepository(),
		memory.NewMemoryUserRepository(),
		httpMiddleware.SessionOptions{
			TTL:                time.Hour,
			CookieName:         httpMiddleware.MerchantSessionCookieName,
			MerchantCookieName: httpMiddleware.MerchantSessionCookieName,
			CustomerCookieName: httpMiddleware.CustomerSessionCookieName,
			LegacyCookieName:   httpMiddleware.SessionCookieName,
		},
	))

	e.GET("/menu", func(c echo.Context) error { return c.String(http.StatusOK, "menu") })
	e.GET("/dashboard", func(c echo.Context) error { return c.String(http.StatusOK, "dashboard") })

	t.Run("customer route issues customer cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/menu", nil)
		req.Host = "order.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), httpMiddleware.CustomerSessionCookieName+"=")
	})

	t.Run("merchant route issues merchant cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.Host = "merchant.bitmerchant.com"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), httpMiddleware.MerchantSessionCookieName+"=")
	})
}
