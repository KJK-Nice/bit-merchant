package http

import (
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	httpMiddleware "bitmerchant/internal/interfaces/http/middleware"
	"errors"

	"github.com/labstack/echo/v4"
)

func setAuthenticatedContext(c echo.Context, user *user.User, session *session.Session) {
	c.Set(httpMiddleware.ContextAuthUser, user)
	c.Set(httpMiddleware.ContextAuthSession, session)
	if session != nil && session.RestaurantID != nil {
		c.Set(httpMiddleware.ContextRestaurantID, *session.RestaurantID)
	}
}

func getAuthenticatedUser(c echo.Context) (*user.User, bool) {
	user, ok := c.Get(httpMiddleware.ContextAuthUser).(*user.User)
	return user, ok
}

func getSession(c echo.Context) (*session.Session, bool) {
	session, ok := c.Get(httpMiddleware.ContextAuthSession).(*session.Session)
	return session, ok
}

func getRestaurantIDFromContext(c echo.Context) (common.RestaurantID, error) {
	if restaurantID, ok := c.Get(httpMiddleware.ContextRestaurantID).(common.RestaurantID); ok && restaurantID != "" {
		return restaurantID, nil
	}
	if session, ok := getSession(c); ok && session.RestaurantID != nil {
		return *session.RestaurantID, nil
	}
	return "", errors.New("restaurant context not available")
}
