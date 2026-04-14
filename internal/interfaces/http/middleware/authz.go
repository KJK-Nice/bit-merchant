package middleware

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	ContextAuthUser     = "authUser"
	ContextAuthSession  = "authSession"
	ContextRestaurantID = "activeRestaurantID"
)

// RequireAuth ensures the request has an authenticated user.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get(ContextAuthUser) == nil {
				return c.Redirect(http.StatusFound, "/auth/login")
			}
			return next(c)
		}
	}
}

// RequireRole ensures the authenticated user has one of the required roles
// in the active restaurant context.
func RequireRole(membershipRepo membership.Repository, roles ...common.MemberRole) echo.MiddlewareFunc {
	allowed := make(map[common.MemberRole]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get(ContextAuthUser).(*user.User)
			if !ok || user == nil {
				return c.Redirect(http.StatusFound, "/auth/login")
			}

			restaurantID, ok := c.Get(ContextRestaurantID).(common.RestaurantID)
			if !ok || restaurantID == "" {
				return c.String(http.StatusForbidden, "restaurant context missing")
			}

			membership, err := membershipRepo.FindByUserAndRestaurant(user.ID, restaurantID)
			if err != nil {
				return c.String(http.StatusForbidden, "membership not found")
			}

			if _, ok := allowed[membership.Role]; !ok {
				return c.String(http.StatusForbidden, "forbidden")
			}

			return next(c)
		}
	}
}
