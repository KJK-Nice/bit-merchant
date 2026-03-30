package auth_test

import (
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

func TestSelectRestaurantUpdatesActiveContext(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	user, err := domain.NewUser("switch-user-1", "Switcher")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restA, err := domain.NewRestaurant("restaurant-a", "Restaurant A")
	require.NoError(t, err)
	restB, err := domain.NewRestaurant("restaurant-b", "Restaurant B")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(restA))
	require.NoError(t, restaurantRepo.Save(restB))

	memA, err := domain.NewMembership("mem-a", user.ID, restA.ID, domain.RoleOwner)
	require.NoError(t, err)
	memB, err := domain.NewMembership("mem-b", user.ID, restB.ID, domain.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(memA))
	require.NoError(t, membershipRepo.Save(memB))

	userID := user.ID
	require.NoError(t, sessionRepo.Save(&domain.Session{
		ID:           "switch-session",
		UserID:       &userID,
		RestaurantID: &restA.ID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}))

	e.Use(httpMiddleware.SessionMiddlewareWithReposAndOptions(
		sessionRepo,
		userRepo,
		httpMiddleware.SessionOptions{TTL: time.Hour},
	))
	authHandler := handler.NewAuthHandler(nil, userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, nil, httpMiddleware.SessionOptions{})
	group := e.Group("/auth")
	group.Use(httpMiddleware.RequireAuth())
	group.POST("/select-restaurant", authHandler.PostSelectRestaurant)

	form := url.Values{}
	form.Set("restaurantID", string(restB.ID))
	req := httptest.NewRequest(http.MethodPost, "/auth/select-restaurant", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.AddCookie(&http.Cookie{Name: httpMiddleware.SessionCookieName, Value: "switch-session"})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard", rec.Header().Get(echo.HeaderLocation))

	session, err := sessionRepo.Get("switch-session")
	require.NoError(t, err)
	require.NotNil(t, session.RestaurantID)
	assert.Equal(t, restB.ID, *session.RestaurantID)
}
