package http

import (
	"context"
	"strings"
	"unicode/utf8"

	"bitmerchant/internal/domain"

	"github.com/labstack/echo/v4"
)

// LayoutUserStrings returns display values for the dashboard sidebar from the authenticated user.
func LayoutUserStrings(user *domain.User) (displayName, subtitle, initials string) {
	if user == nil {
		return "Guest", "", "?"
	}
	displayName = strings.TrimSpace(user.DisplayName)
	if displayName == "" {
		displayName = string(user.ID)
	}
	subtitle = "ID " + string(user.ID)
	initials = userInitials(displayName)
	return displayName, subtitle, initials
}

// LayoutUserStringsFromContext resolves layout user strings from the Echo context.
func LayoutUserStringsFromContext(c echo.Context) (displayName, subtitle, initials string) {
	u, _ := getAuthenticatedUser(c)
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
func ActiveRestaurantLabel(ctx context.Context, id domain.RestaurantID, repo domain.RestaurantRepository) string {
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
