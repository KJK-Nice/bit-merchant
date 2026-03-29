package http

import (
	"bitmerchant/internal/domain"
	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
)

func setAuthenticatedContext(c echo.Context, user *domain.User, session *domain.Session) {
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextAuthSession, session)
	if session != nil && session.RestaurantID != nil {
		c.Set(httpMiddleware.ContextRestaurantID, *session.RestaurantID)
	}
}

func getAuthenticatedUser(c echo.Context) (*domain.User, bool) {
	user, ok := c.Get(httpMiddleware.ContextAuthUser).(*domain.User)
	return user, ok
}

func getSession(c echo.Context) (*domain.Session, bool) {
	session, ok := c.Get(httpMiddleware.ContextAuthSession).(*domain.Session)
	return session, ok
}

func getRestaurantIDFromContext(c echo.Context) (domain.RestaurantID, error) {
	if restaurantID, ok := c.Get(httpMiddleware.ContextRestaurantID).(domain.RestaurantID); ok && restaurantID != "" {
		return restaurantID, nil
	}
	if session, ok := getSession(c); ok && session.RestaurantID != nil {
		return *session.RestaurantID, nil
	}
	// Backward-compatible fallback for legacy handlers/tests.
	return domain.RestaurantID("restaurant_1"), nil
}
