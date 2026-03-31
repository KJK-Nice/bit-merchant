package middleware

import (
	"bitmerchant/internal/domain"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const SessionCookieName = "bitmerchant_session"

const defaultSessionTTL = 24 * time.Hour

// SessionOptions configures cookie and persistence behavior.
type SessionOptions struct {
	TTL          time.Duration
	SecureCookie bool
}

func (o SessionOptions) WithDefaults() SessionOptions {
	if o.TTL <= 0 {
		o.TTL = defaultSessionTTL
	}
	return o
}

// NewSessionID generates a random session identifier.
func NewSessionID() string {
	return uuid.New().String()
}

// NewSessionCookie builds the session cookie with safe defaults.
func NewSessionCookie(sessionID string, opts SessionOptions) *http.Cookie {
	opts = opts.WithDefaults()
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(opts.TTL),
		SameSite: http.SameSiteStrictMode,
		Secure:   opts.SecureCookie,
	}
}

// ShouldUseSecureCookies decides whether to enable Secure cookies.
func ShouldUseSecureCookies(baseURL string, forceSecure bool) bool {
	if forceSecure {
		return true
	}
	return strings.HasPrefix(strings.ToLower(baseURL), "https://")
}

// SessionMiddleware ensures a session ID cookie exists
func SessionMiddleware() echo.MiddlewareFunc {
	return sessionMiddleware(nil, nil, SessionOptions{})
}

// SessionMiddlewareWithRepos enables authenticated session loading from repositories.
func SessionMiddlewareWithRepos(sessionRepo domain.SessionRepository, userRepo domain.UserRepository) echo.MiddlewareFunc {
	return sessionMiddleware(sessionRepo, userRepo, SessionOptions{})
}

// SessionMiddlewareWithReposAndOptions enables authenticated session loading with explicit options.
func SessionMiddlewareWithReposAndOptions(sessionRepo domain.SessionRepository, userRepo domain.UserRepository, opts SessionOptions) echo.MiddlewareFunc {
	return sessionMiddleware(sessionRepo, userRepo, opts)
}

func sessionMiddleware(sessionRepo domain.SessionRepository, userRepo domain.UserRepository, opts SessionOptions) echo.MiddlewareFunc {
	opts = opts.WithDefaults()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionID := ensureSessionCookie(c, opts)
			c.Set("sessionID", sessionID)

			if sessionRepo == nil || userRepo == nil {
				return next(c)
			}

			session := loadOrCreateSession(sessionRepo, sessionID, opts)
			c.Set(ContextAuthSession, session)
			attachIdentityFromSession(c, session, userRepo)
			return next(c)
		}
	}
}

func ensureSessionCookie(c echo.Context, opts SessionOptions) string {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		sessionID := NewSessionID()
		c.SetCookie(NewSessionCookie(sessionID, opts))
		return sessionID
	}
	return cookie.Value
}

func loadOrCreateSession(sessionRepo domain.SessionRepository, sessionID string, opts SessionOptions) *domain.Session {
	session, err := sessionRepo.Get(sessionID)
	if err != nil || session == nil || session.IsExpired(time.Now()) {
		session = &domain.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(opts.TTL),
		}
		_ = sessionRepo.Save(session)
	}
	return session
}

func attachIdentityFromSession(c echo.Context, session *domain.Session, userRepo domain.UserRepository) {
	if session.UserID != nil {
		user, err := userRepo.FindByID(*session.UserID)
		if err == nil && user != nil {
			c.Set(ContextAuthUser, user)
		}
	}
	if session.RestaurantID != nil {
		c.Set(ContextRestaurantID, *session.RestaurantID)
	}
}
