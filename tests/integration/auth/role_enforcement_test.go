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

func TestRequireRoleRouteProtection(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	restaurantID := domain.RestaurantID("restaurant-1")

	ownerUser, err := domain.NewUser("owner-1", "Owner")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(ownerUser))
	ownerUserID := ownerUser.ID
	require.NoError(t, sessionRepo.Save(&domain.Session{
		ID:           "owner-session",
		UserID:       &ownerUserID,
		RestaurantID: &restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))
	ownerMembership, err := domain.NewMembership("owner-membership", ownerUser.ID, restaurantID, domain.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(ownerMembership))

	staffUser, err := domain.NewUser("staff-1", "Staff")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(staffUser))
	staffUserID := staffUser.ID
	require.NoError(t, sessionRepo.Save(&domain.Session{
		ID:           "staff-session",
		UserID:       &staffUserID,
		RestaurantID: &restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))
	staffMembership, err := domain.NewMembership("staff-membership", staffUser.ID, restaurantID, domain.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(staffMembership))

	e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
		sessionRepo,
		userRepo,
		httpMiddleware.SessionOptions{TTL: time.Hour},
	))
	protected := e.Group("/owner-only")
	protected.Use(httpMiddleware.RequireAuth(), httpMiddleware.RequireRole(membershipRepo, domain.RoleOwner))
	protected.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "owner-ok")
	})

	t.Run("owner is allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner-only", nil)
		req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: "owner-session"})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "owner-ok", rec.Body.String())
	})

	t.Run("kitchen staff is forbidden on owner-only route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/owner-only", nil)
		req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: "staff-session"})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "forbidden")
	})
}
