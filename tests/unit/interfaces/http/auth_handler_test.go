package http_test

import (
	"net/http"
	"net/http/httptest"
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

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, httpMiddleware.SessionOptions{})

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

	h := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, httpMiddleware.SessionOptions{})

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
