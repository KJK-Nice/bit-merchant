package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuthRouteProtection(t *testing.T) {
	t.Run("unauthenticated request redirects to login", func(t *testing.T) {
		e := echo.New()
		e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
			memory.NewMemorySessionRepository(),
			memory.NewMemoryUserRepository(),
			httpMiddleware.SessionOptions{TTL: time.Hour},
		))
		protected := e.Group("/protected")
		protected.Use(httpMiddleware.RequireAuth())
		protected.GET("", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/auth/login", rec.Header().Get(echo.HeaderLocation))
	})

	t.Run("authenticated request passes", func(t *testing.T) {
		e := echo.New()
		sessionRepo := memory.NewMemorySessionRepository()
		userRepo := memory.NewMemoryUserRepository()

		user, err := domain.NewUser("user-1", "Owner")
		require.NoError(t, err)
		require.NoError(t, userRepo.Save(user))

		userID := user.ID
		require.NoError(t, sessionRepo.Save(&domain.Session{
			ID:        "authenticated-session",
			UserID:    &userID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}))

		e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
			sessionRepo,
			userRepo,
			httpMiddleware.SessionOptions{TTL: time.Hour},
		))
		protected := e.Group("/protected")
		protected.Use(httpMiddleware.RequireAuth())
		protected.GET("", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: "authenticated-session"})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})
}
