package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// CSPMiddleware sets a per-request Content-Security-Policy header with a random nonce
// and injects that nonce into the templ rendering context via templ.WithNonce.
//
// Ship as report-only first: swap the header name to
// "Content-Security-Policy-Report-Only" during an initial staging window.
func CSPMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nonce, err := generateNonce()
			if err != nil {
				return fmt.Errorf("csp: generate nonce: %w", err)
			}

			// Inject nonce into the templ rendering context so that
			// templ.GetNonce(ctx) and utils.ComponentScript() pick it up automatically.
			ctx := templ.WithNonce(c.Request().Context(), nonce)
			c.SetRequest(c.Request().WithContext(ctx))

			csp := buildCSP(nonce)
			c.Response().Header().Set("Content-Security-Policy", csp)

			return next(c)
		}
	}
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func buildCSP(nonce string) string {
	return fmt.Sprintf(
		// script-src: same-origin scripts + per-request nonce for inline scripts.
		//   'unsafe-eval' is required by Datastar v1 which uses new Function() to
		//   evaluate reactive expressions (data-show, data-bind, etc.).
		//   See: assets/js/datastar.js — Function("el","$","__action","evt",...)
		// style-src: unsafe-inline retained — Tailwind utility classes are static-file
		//   but some templUI components inject inline style strings.
		// connect-src 'self': covers Datastar SSE streams.
		// worker-src 'self': allows /sw.js to be registered as a service worker.
		"default-src 'self'; "+
			"script-src 'self' 'unsafe-eval' 'nonce-%s'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: blob:; "+
			"connect-src 'self'; "+
			"font-src 'self' data:; "+
			"manifest-src 'self'; "+
			"worker-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'",
		nonce,
	)
}
