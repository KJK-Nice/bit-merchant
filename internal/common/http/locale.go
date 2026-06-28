package commonhttp

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// LocaleCookieName holds the customer's chosen menu language.
const LocaleCookieName = "lang"

// DefaultLocale is the base language; menu items fall back to it.
const DefaultLocale = "en"

func normalizeLocale(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	// Accept-Language entries look like "en-US,en;q=0.9" or "es-419"; take the
	// primary subtag of the first option.
	if i := strings.IndexAny(raw, ",;"); i >= 0 {
		raw = raw[:i]
	}
	if i := strings.IndexByte(raw, '-'); i >= 0 {
		raw = raw[:i]
	}
	if len(raw) < 2 || len(raw) > 8 {
		return ""
	}
	for _, r := range raw {
		if r < 'a' || r > 'z' {
			return ""
		}
	}
	return raw
}

// ResolveLocale determines the menu language for the request, in priority order:
// an explicit ?lang= query param (also persisted to a cookie), the lang cookie,
// the Accept-Language header, then DefaultLocale. When ?lang= is present it sets
// the cookie so the choice survives navigation.
func ResolveLocale(c echo.Context) string {
	if q := normalizeLocale(c.QueryParam("lang")); q != "" {
		c.SetCookie(&http.Cookie{
			Name:     LocaleCookieName,
			Value:    q,
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 365,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		})
		return q
	}
	if ck, err := c.Cookie(LocaleCookieName); err == nil {
		if v := normalizeLocale(ck.Value); v != "" {
			return v
		}
	}
	if h := normalizeLocale(c.Request().Header.Get("Accept-Language")); h != "" {
		return h
	}
	return DefaultLocale
}
