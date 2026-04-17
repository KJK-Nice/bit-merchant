package commonhttp

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
	mw "bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/labstack/echo/v4"
)

// SetAuthenticatedContext stores user and session on the Echo context.
func SetAuthenticatedContext(c echo.Context, u *user.User, s *session.Session) {
	c.Set(mw.ContextAuthUser, u)
	c.Set(mw.ContextAuthSession, s)
	if s != nil && s.RestaurantID != nil {
		c.Set(mw.ContextRestaurantID, *s.RestaurantID)
	}
}

// GetAuthenticatedUser returns the authenticated user from context, if any.
func GetAuthenticatedUser(c echo.Context) (*user.User, bool) {
	u, ok := c.Get(mw.ContextAuthUser).(*user.User)
	return u, ok
}

// GetSession returns the session from context, if any.
func GetSession(c echo.Context) (*session.Session, bool) {
	s, ok := c.Get(mw.ContextAuthSession).(*session.Session)
	return s, ok
}

// RestaurantIDFromContext resolves the active restaurant ID from Echo context.
func RestaurantIDFromContext(c echo.Context) (common.RestaurantID, error) {
	if restaurantID, ok := c.Get(mw.ContextRestaurantID).(common.RestaurantID); ok && restaurantID != "" {
		return restaurantID, nil
	}
	if s, ok := GetSession(c); ok && s.RestaurantID != nil {
		return *s.RestaurantID, nil
	}
	return "", errors.New("restaurant context not available")
}

// LayoutUserStrings returns display values for the dashboard sidebar from the authenticated user.
func LayoutUserStrings(u *user.User) (displayName, subtitle, initials string) {
	if u == nil {
		return "Guest", "", "?"
	}
	displayName = strings.TrimSpace(u.DisplayName)
	if displayName == "" {
		displayName = string(u.ID)
	}
	subtitle = "ID " + string(u.ID)
	initials = userInitials(displayName)
	return displayName, subtitle, initials
}

// LayoutUserStringsFromContext resolves layout user strings from the Echo context.
func LayoutUserStringsFromContext(c echo.Context) (displayName, subtitle, initials string) {
	u, _ := GetAuthenticatedUser(c)
	return LayoutUserStrings(u)
}

func userInitials(displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return "?"
	}
	fields := strings.Fields(displayName)
	if len(fields) >= 2 {
		r1, _ := utf8.DecodeRuneInString(fields[0])
		r2, _ := utf8.DecodeRuneInString(fields[1])
		return strings.ToUpper(string(r1) + string(r2))
	}
	word := fields[0]
	r1, w := utf8.DecodeRuneInString(word)
	if len(word) > w {
		r2, _ := utf8.DecodeRuneInString(word[w:])
		return strings.ToUpper(string(r1) + string(r2))
	}
	return strings.ToUpper(string(r1))
}

// ActiveRestaurantLabel returns the restaurant name when available, otherwise the raw ID.
func ActiveRestaurantLabel(ctx context.Context, id common.RestaurantID, repo restaurant.Repository) string {
	if id == "" {
		return ""
	}
	if repo != nil {
		rest, err := repo.FindByID(id)
		if err == nil && rest != nil && rest.Name != "" {
			return rest.Name
		}
	}
	return string(id)
}
