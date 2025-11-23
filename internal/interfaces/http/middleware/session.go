package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const SessionCookieName = "bitmerchant_session"

// SessionMiddleware ensures a session ID cookie exists
func SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				// Create new session
				sessionID := uuid.New().String()
				cookie = &http.Cookie{
					Name:     SessionCookieName,
					Value:    sessionID,
					Path:     "/",
					HttpOnly: true,
					Expires:  time.Now().Add(24 * time.Hour),
					SameSite: http.SameSiteStrictMode,
				}
				c.SetCookie(cookie)
				c.Set("sessionID", sessionID)
			} else {
				c.Set("sessionID", cookie.Value)
			}
			return next(c)
		}
	}
}

