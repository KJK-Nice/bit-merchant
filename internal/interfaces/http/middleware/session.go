package middleware

import (
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "bitmerchant_session" // legacy/default cookie name
const MerchantSessionCookieName = "bitmerchant_merchant_session"
const CustomerSessionCookieName = "bitmerchant_customer_session"

const defaultSessionTTL = 24 * time.Hour

// SessionOptions configures cookie and persistence behavior.
type SessionOptions struct {
	TTL                time.Duration
	SecureCookie       bool
	CookieName         string
	MerchantCookieName string
	CustomerCookieName string
	LegacyCookieName   string
}

func (o SessionOptions) WithDefaults() SessionOptions {
	if o.TTL <= 0 {
		o.TTL = defaultSessionTTL
	}
	if o.CookieName == "" {
		o.CookieName = SessionCookieName
	}
	if o.MerchantCookieName == "" {
		o.MerchantCookieName = MerchantSessionCookieName
	}
	if o.CustomerCookieName == "" {
		o.CustomerCookieName = CustomerSessionCookieName
	}
	if o.LegacyCookieName == "" {
		o.LegacyCookieName = SessionCookieName
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
	return NewSessionCookieWithName(opts.CookieName, sessionID, opts)
}

func NewSessionCookieWithName(cookieName, sessionID string, opts SessionOptions) *http.Cookie {
	opts = opts.WithDefaults()
	return &http.Cookie{
		Name:     cookieName,
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
func SessionMiddlewareWithRepos(sessionRepo session.Repository, userRepo user.Repository) echo.MiddlewareFunc {
	return sessionMiddleware(sessionRepo, userRepo, SessionOptions{})
}

// SessionMiddlewareWithReposAndOptions enables authenticated session loading with explicit options.
func SessionMiddlewareWithReposAndOptions(sessionRepo session.Repository, userRepo user.Repository, opts SessionOptions) echo.MiddlewareFunc {
	return sessionMiddleware(sessionRepo, userRepo, opts)
}

func sessionMiddleware(sessionRepo session.Repository, userRepo user.Repository, opts SessionOptions) echo.MiddlewareFunc {
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
	cookieName := resolveSessionCookieName(c, opts)
	cookie, err := c.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		if opts.LegacyCookieName != "" && opts.LegacyCookieName != cookieName {
			legacyCookie, legacyErr := c.Cookie(opts.LegacyCookieName)
			if legacyErr == nil && legacyCookie != nil && legacyCookie.Value != "" {
				c.SetCookie(NewSessionCookieWithName(cookieName, legacyCookie.Value, opts))
				return legacyCookie.Value
			}
		}
		sessionID := NewSessionID()
		c.SetCookie(NewSessionCookieWithName(cookieName, sessionID, opts))
		return sessionID
	}
	return cookie.Value
}

func resolveSessionCookieName(c echo.Context, opts SessionOptions) string {
	opts = opts.WithDefaults()

	if routeSurface, ok := c.Get(ContextRouteSurface).(AppSurface); ok {
		switch routeSurface {
		case AppSurfaceMerchant:
			return opts.MerchantCookieName
		case AppSurfaceCustomer:
			return opts.CustomerCookieName
		}
	}

	if hostSurface, ok := c.Get(ContextHostSurface).(AppSurface); ok {
		switch hostSurface {
		case AppSurfaceMerchant:
			return opts.MerchantCookieName
		case AppSurfaceCustomer:
			return opts.CustomerCookieName
		}
	}

	return opts.CookieName
}

func loadOrCreateSession(sessionRepo session.Repository, sessionID string, opts SessionOptions) *session.Session {
	currentSession, err := sessionRepo.Get(sessionID)
	if err != nil || currentSession == nil || currentSession.IsExpired(time.Now()) {
		currentSession = &session.Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(opts.TTL),
		}
		_ = sessionRepo.Save(currentSession)
	}
	return currentSession
}

func attachIdentityFromSession(c echo.Context, session *session.Session, userRepo user.Repository) {
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
