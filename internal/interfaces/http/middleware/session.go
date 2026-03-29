package middleware

import (
	"bitmerchant/internal/domain"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const SessionCookieName = "bitmerchant_session"

// SessionMiddleware ensures a session ID cookie exists
func SessionMiddleware() echo.MiddlewareFunc {
	return sessionMiddleware(nil, nil)
}

// SessionMiddlewareWithRepos enables authenticated session loading from repositories.
func SessionMiddlewareWithRepos(sessionRepo domain.SessionRepository, userRepo domain.UserRepository) echo.MiddlewareFunc {
	return sessionMiddleware(sessionRepo, userRepo)
}

func sessionMiddleware(sessionRepo domain.SessionRepository, userRepo domain.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(SessionCookieName)
			var sessionID string
			if err != nil || cookie.Value == "" {
				// Create new session
				sessionID = uuid.New().String()
				cookie = &http.Cookie{
					Name:     SessionCookieName,
					Value:    sessionID,
					Path:     "/",
					HttpOnly: true,
					Expires:  time.Now().Add(24 * time.Hour),
					SameSite: http.SameSiteStrictMode,
				}
				c.SetCookie(cookie)
			} else {
				sessionID = cookie.Value
			}

			c.Set("sessionID", sessionID)

			// Preserve legacy anonymous-only behavior when repositories are not configured.
			if sessionRepo == nil || userRepo == nil {
				return next(c)
			}

			session, err := sessionRepo.Get(sessionID)
			if err != nil || session == nil || session.IsExpired(time.Now()) {
				session = &domain.Session{
					ID:        sessionID,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				_ = sessionRepo.Save(session)
			}

			c.Set(ContextAuthSession, session)

			if session.UserID != nil {
				user, userErr := userRepo.FindByID(*session.UserID)
				if userErr == nil && user != nil {
					c.Set(ContextAuthUser, user)
				}
			}
			if session.RestaurantID != nil {
				c.Set(ContextRestaurantID, *session.RestaurantID)
			}
			return next(c)
		}
	}
}
