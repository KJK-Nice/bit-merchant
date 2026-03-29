package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingInvitationRepo struct{}

func (f *failingInvitationRepo) Save(invitation *domain.Invitation) error {
	return assert.AnError
}

func TestAuthHandlerPostSelectRestaurant_UpdatesSessionAndRedirectsByRole(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	user, err := domain.NewUser("staff-1", "Staff")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restID := domain.RestaurantID("restaurant-kitchen")
	rest, err := domain.NewRestaurant(restID, "Kitchen Place")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(rest))

	membership, err := domain.NewMembership("mem-kitchen", user.ID, restID, domain.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	require.NoError(t, sessionRepo.Save(&domain.Session{
		ID:        "session-select",
		UserID:    &user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, nil, httpMiddleware.SessionOptions{})

	form := url.Values{}
	form.Set("restaurantID", string(restID))
	req := httptest.NewRequest(http.MethodPost, "/auth/select-restaurant", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-select")
	c.Set(httpMiddleware.ContextAuthUser, user)

	require.NoError(t, h.PostSelectRestaurant(c))
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/kitchen", rec.Header().Get(echo.HeaderLocation))

	session, err := sessionRepo.Get("session-select")
	require.NoError(t, err)
	require.NotNil(t, session.RestaurantID)
	assert.Equal(t, restID, *session.RestaurantID)
}

func TestAuthHandlerGetSelectRestaurant_RendersRestaurantOptions(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	user, err := domain.NewUser("owner-2", "Owner Two")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restID := domain.RestaurantID("restaurant-owner")
	rest, err := domain.NewRestaurant(restID, "Owner Place")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(rest))

	membership, err := domain.NewMembership("mem-owner", user.ID, restID, domain.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, nil, httpMiddleware.SessionOptions{})

	req := httptest.NewRequest(http.MethodGet, "/auth/select-restaurant", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextAuthSession, &domain.Session{
		ID:           "s",
		UserID:       &user.ID,
		RestaurantID: &restID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	})

	require.NoError(t, h.GetSelectRestaurant(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	body, readErr := io.ReadAll(rec.Result().Body)
	require.NoError(t, readErr)
	assert.Contains(t, string(body), "Owner Place")
	assert.Contains(t, string(body), "restaurantID")
}
func (f *failingInvitationRepo) FindByToken(token string) (*domain.Invitation, error) {
	return nil, assert.AnError
}
func (f *failingInvitationRepo) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Invitation, error) {
	return nil, assert.AnError
}
func (f *failingInvitationRepo) Update(invitation *domain.Invitation) error {
	return assert.AnError
}

func TestAuthHandlerLogout_DeletesSessionAndExpiresCookie(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()

	require.NoError(t, sessionRepo.Save(&domain.Session{
		ID:        "session-to-logout",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, nil, httpMiddleware.SessionOptions{})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-to-logout")

	require.NoError(t, h.Logout(c))
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/menu", rec.Header().Get(echo.HeaderLocation))

	_, err := sessionRepo.Get("session-to-logout")
	assert.Error(t, err)
	assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), "Max-Age=0")
}

func TestAuthHandlerCreateInvitation_SanitizesSaveErrors(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := &failingInvitationRepo{}

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, nil, httpMiddleware.SessionOptions{})

	user, err := domain.NewUser("owner-1", "Owner")
	require.NoError(t, err)
	restaurantID := domain.RestaurantID("restaurant-1")
	membership, err := domain.NewMembership("mem-1", user.ID, restaurantID, domain.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	req := httptest.NewRequest(http.MethodPost, "/dashboard/invite", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextRestaurantID, restaurantID)

	require.NoError(t, h.CreateInvitation(c))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "failed to save invitation", strings.TrimSpace(rec.Body.String()))
}
