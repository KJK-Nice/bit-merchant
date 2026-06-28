package auth_test

import (
	authapp "bitmerchant/internal/auth/app"
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSelectRestaurantUpdatesActiveContext(t *testing.T) {
	e := echo.New()
	sessionRepo := memory.NewMemorySessionRepository()
	userRepo := memory.NewMemoryUserRepository()
	membershipRepo := memory.NewMemoryMembershipRepository()
	invitationRepo := memory.NewMemoryInvitationRepository()
	restaurantRepo := memory.NewMemoryRestaurantRepository()

	user, err := user.NewUser("switch-user-1", "Switcher")
	require.NoError(t, err)
	require.NoError(t, userRepo.Save(user))

	restA, err := restaurant.NewRestaurant("restaurant-a", "Restaurant A")
	require.NoError(t, err)
	restB, err := restaurant.NewRestaurant("restaurant-b", "Restaurant B")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(restA))
	require.NoError(t, restaurantRepo.Save(restB))

	memA, err := membership.NewMembership("mem-a", user.ID, restA.ID, common.RoleOwner)
	require.NoError(t, err)
	memB, err := membership.NewMembership("mem-b", user.ID, restB.ID, common.RoleOwner)
	require.NoError(t, err)
	require.NoError(t, membershipRepo.Save(memA))
	require.NoError(t, membershipRepo.Save(memB))

	userID := user.ID
	require.NoError(t, sessionRepo.Save(&session.Session{
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
	createUC := restaurantCmd.NewCreateRestaurantHandler(restaurantRepo, nil, nil)
	authApp := authapp.NewApplication(userRepo, membershipRepo, invitationRepo, sessionRepo, restaurantRepo, nil, nil, "", createUC, nil, nil, nil)
	authHandler := authhttp.NewAuthHandler(nil, authApp, nil, httpMiddleware.SessionOptions{})
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
