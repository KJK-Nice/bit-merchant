package auth_test

import (
	"bitmerchant/internal/auth/domain/membership"
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
	"testing"
	"time"
)

func TestRequireRoleRouteProtection(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	restaurantID := common.RestaurantID("restaurant-1")

	ownerUser, err := user.NewUser("owner-1", "Owner")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(ownerUser))
	ownerUserID := ownerUser.ID
	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:           "owner-session",
		UserID:       &ownerUserID,
		RestaurantID: &restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))
	ownerMembership, err := membership.NewMembership("owner-membership", ownerUser.ID, restaurantID, common.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(ownerMembership))

	staffUser, err := user.NewUser("staff-1", "Staff")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(staffUser))
	staffUserID := staffUser.ID
	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:           "staff-session",
		UserID:       &staffUserID,
		RestaurantID: &restaurantID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))
	staffMembership, err := membership.NewMembership("staff-membership", staffUser.ID, restaurantID, common.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(staffMembership))

	e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
		sessionRepo,
		userRepo,
		httpMiddleware.SessionOptions{TTL: time.Hour},
	))
	protected := e.Group("/owner-only")
	protected.Use(httpMiddleware.RequireAuth(), httpMiddleware.RequireRole(membershipRepo, common.RoleOwner))
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
