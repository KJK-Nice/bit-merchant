package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuth_RedirectsWhenUnauthenticated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := httpMiddleware.RequireAuth()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/auth/login", rec.Header().Get(echo.HeaderLocation))
}

func TestRequireRole_AllowsWhenRoleMatches(t *testing.T) {
	e := echo.New()
	membershipRepo := memory.NewMemoryMembershipRepository()

	user, err := domain.NewUser("user-1", "Owner")
	require.NoError(t, err)
	restaurantID := domain.RestaurantID("restaurant-1")
	membership, err := domain.NewMembership("membership-1", user.ID, restaurantID, domain.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextRestaurantID, restaurantID)

	handler := httpMiddleware.RequireRole(membershipRepo, domain.RoleOwner)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_DeniesWhenRoleDoesNotMatch(t *testing.T) {
	e := echo.New()
	membershipRepo := memory.NewMemoryMembershipRepository()

	user, err := domain.NewUser("user-1", "Staff")
	require.NoError(t, err)
	restaurantID := domain.RestaurantID("restaurant-1")
	membership, err := domain.NewMembership("membership-1", user.ID, restaurantID, domain.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextRestaurantID, restaurantID)

	handler := httpMiddleware.RequireRole(membershipRepo, domain.RoleOwner)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "forbidden")
}

func TestRequireRole_DeniesWhenRestaurantContextMissing(t *testing.T) {
	e := echo.New()
	membershipRepo := memory.NewMemoryMembershipRepository()

	user, err := domain.NewUser("user-1", "Owner")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)

	handler := httpMiddleware.RequireRole(membershipRepo, domain.RoleOwner)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "restaurant context missing")
}
