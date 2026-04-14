package middleware_test

import (
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	"bitmerchant/internal/infrastructure/repositories/memory"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionMiddlewareWithReposAndOptions_SetsSecureCookie(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := httpMiddleware.SessionMiddlewareWithReposAndOptions(
		memory.NewMemorySessionRepository(),
		memory.NewMemoryUserRepository(),
		httpMiddleware.SessionOptions{SecureCookie: true, TTL: time.Hour},
	)(func(c echo.Context) error {
		sessionID, ok := c.Get("sessionID").(string)
		require.True(t, ok)
		require.NotEmpty(t, sessionID)
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	cookieHeader := rec.Header().Get(echo.HeaderSetCookie)
	assert.Contains(t, cookieHeader, httpMiddleware.SessionCookieName+"=")
	assert.Contains(t, cookieHeader, "HttpOnly")
	assert.Contains(t, cookieHeader, "Secure")
	assert.Contains(t, cookieHeader, "SameSite=Strict")
}

func TestSessionMiddlewareWithReposAndOptions_ReplacesExpiredSession(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	expiredID := "expired-session-id"

	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:        expiredID,
		CreatedAt: time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: expiredID})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := httpMiddleware.SessionMiddlewareWithReposAndOptions(
		sessionRepo,
		userRepo,
		httpMiddleware.SessionOptions{TTL: time.Hour},
	)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))

	session, err := sessionRepo.Get(expiredID)
	require.NoError(t, err)
	assert.Nil(t, session.UserID)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestSessionMiddlewareWithReposAndOptions_LoadsAuthenticatedContext(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()

	authUser, err := user.NewUser("user-1", "Jane")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(authUser))

	restaurantID := common.RestaurantID("restaurant-1")
	userID := authUser.ID
	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:           "auth-session",
		UserID:       &userID,
		RestaurantID: &restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: "auth-session"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := httpMiddleware.SessionMiddlewareWithReposAndOptions(
		sessionRepo,
		userRepo,
		httpMiddleware.SessionOptions{TTL: time.Hour},
	)(func(c echo.Context) error {
		authUser, ok := c.Get(httpMiddleware.ContextAuthUser).(*user.User)
		require.True(t, ok)
		require.Equal(t, userID, authUser.ID)

		ctxRestaurantID, ok := c.Get(httpMiddleware.ContextRestaurantID).(common.RestaurantID)
		require.True(t, ok)
		require.Equal(t, restaurantID, ctxRestaurantID)
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
}

func TestShouldUseSecureCookies(t *testing.T) {
	assert.True(t, httpMiddleware.ShouldUseSecureCookies("https://example.com", false))
	assert.False(t, httpMiddleware.ShouldUseSecureCookies("http://localhost:8080", false))
	assert.True(t, httpMiddleware.ShouldUseSecureCookies("http://localhost:8080", true))
	assert.True(t, httpMiddleware.ShouldUseSecureCookies(strings.ToUpper("https://example.com"), false))
}
