package http_test

import (
	authapp "bitmerchant/internal/auth/app"
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	authhttp "bitmerchant/internal/auth/ports/http"
	httpMiddleware "bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/repositories/memory"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newTestAuthHandler(
	userRepo user.Repository,
	membershipRepo membership.Repository,
	invitationRepo invitation.Repository,
	sessionRepo session.Repository,
	restaurantRepo restaurant.Repository,
	createUC restaurantCmd.CreateRestaurantHandler,
	sessionOpts httpMiddleware.SessionOptions,
) *authhttp.AuthHandler {
	if restaurantRepo == nil {
		restaurantRepo = memory.NewMemoryRestaurantRepository()
	}
	if createUC == nil {
		createUC = restaurantCmd.NewCreateRestaurantHandler(restaurantRepo, nil, nil)
	}
	app := authapp.NewApplication(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, createUC, nil, nil)
	return authhttp.NewAuthHandler(nil, app, nil, sessionOpts)
}

type failingInvitationRepo struct{}

func (f *failingInvitationRepo) Save(invitation *invitation.Invitation) error {
	return assert.AnError
}

func TestAuthHandlerPostSelectRestaurant_UpdatesSessionAndRedirectsByRole(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	user, err := user.NewUser("staff-1", "Staff")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restID := common.RestaurantID("restaurant-kitchen")
	rest, err := restaurant.NewRestaurant(restID, "Kitchen Place")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(rest))

	membership, err := membership.NewMembership("mem-kitchen", user.ID, restID, common.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:        "session-select",
		UserID:    &user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, httpMiddleware.SessionOptions{})

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

	user, err := user.NewUser("owner-2", "Owner Two")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restID := common.RestaurantID("restaurant-owner")
	rest, err := restaurant.NewRestaurant(restID, "Owner Place")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(rest))

	membership, err := membership.NewMembership("mem-owner", user.ID, restID, common.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, httpMiddleware.SessionOptions{})

	req := httptest.NewRequest(http.MethodGet, "/auth/select-restaurant", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextAuthSession, &session.Session{
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

func TestAuthHandlerPostNewRestaurant_CreatesMembershipAndSwitchesSession(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()
	createUC := restaurantCmd.NewCreateRestaurantHandler(restaurantRepo, nil, nil)

	user, err := user.NewUser("owner-newr", "Owner NewR")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	existingID := common.RestaurantID("restaurant-existing")
	existingRest, err := restaurant.NewRestaurant(existingID, "Existing")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(existingRest))

	membership, err := membership.NewMembership("mem-existing", user.ID, existingID, common.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:           "session-new-rest",
		UserID:       &user.ID,
		RestaurantID: &existingID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, createUC, httpMiddleware.SessionOptions{TTL: time.Hour})

	form := url.Values{}
	form.Set("name", "Brand New Cafe")
	req := httptest.NewRequest(http.MethodPost, "/auth/restaurants", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-new-rest")
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextRestaurantID, existingID)

	require.NoError(t, h.PostNewRestaurant(c))
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard", rec.Header().Get(echo.HeaderLocation))

	allMems, err := membershipRepo.FindByUserID(user.ID)
	require.NoError(t, err)
	require.Len(t, allMems, 2)

	session, err := sessionRepo.Get("session-new-rest")
	require.NoError(t, err)
	require.NotNil(t, session.RestaurantID)
	assert.NotEqual(t, existingID, *session.RestaurantID)

	found, err := restaurantRepo.FindByID(*session.RestaurantID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Brand New Cafe", found.Name)
}

func TestAuthHandlerPostNewRestaurant_ForbiddenForKitchenStaff(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()
	createUC := restaurantCmd.NewCreateRestaurantHandler(restaurantRepo, nil, nil)

	user, err := user.NewUser("kitchen-u", "Kitchen")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restID := common.RestaurantID("restaurant-k")
	rest, err := restaurant.NewRestaurant(restID, "Kitchen Rest")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(rest))

	membership, err := membership.NewMembership("mem-k", user.ID, restID, common.RoleKitchenStaff)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(membership))

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, createUC, httpMiddleware.SessionOptions{})

	form := url.Values{}
	form.Set("name", "Sneaky Place")
	req := httptest.NewRequest(http.MethodPost, "/auth/restaurants", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "s")
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextRestaurantID, restID)

	require.NoError(t, h.PostNewRestaurant(c))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func (f *failingInvitationRepo) FindByToken(token string) (*invitation.Invitation, error) {
	return nil, assert.AnError
}
func (f *failingInvitationRepo) FindByRestaurantID(restaurantID common.RestaurantID) ([]*invitation.Invitation, error) {
	return nil, assert.AnError
}
func (f *failingInvitationRepo) Update(invitation *invitation.Invitation) error {
	return assert.AnError
}

func TestAuthHandlerLogout_DeletesSessionAndExpiresCookie(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()

	require.NoError(t, sessionRepo.Save(&session.Session{
		ID:        "session-to-logout",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, httpMiddleware.SessionOptions{})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "session-to-logout")

	require.NoError(t, h.Logout(c))
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/", rec.Header().Get(echo.HeaderLocation))

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

	h := newTestAuthHandler(userRepo, membershipRepo, invitationRepo, sessionRepo, nil, nil, httpMiddleware.SessionOptions{})

	user, err := user.NewUser("owner-1", "Owner")
	require.NoError(t, err)
	restaurantID := common.RestaurantID("restaurant-1")
	membership, err := membership.NewMembership("mem-1", user.ID, restaurantID, common.RoleOwner)
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
