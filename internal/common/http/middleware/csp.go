package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// CSPMiddleware sets a per-request Content-Security-Policy header with a random nonce
// and injects that nonce into the templ rendering context via templ.WithNonce.
//
// extraImageURLs is an optional list of base URLs whose origins are added to
// img-src. Use this to allow images served from external object storage
// (e.g. S3_PUBLIC_BASE_URL = "https://t3.storageapi.dev/bucket-name").
// Only the scheme+host is extracted; paths and query strings are ignored.
//
// Ship as report-only first: swap the header name to
// "Content-Security-Policy-Report-Only" during an initial staging window.
func CSPMiddleware(extraImageURLs ...string) echo.MiddlewareFunc {
	imgOrigins := parseOrigins(extraImageURLs)
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

			csp := buildCSP(nonce, imgOrigins)
			c.Response().Header().Set("Content-Security-Policy", csp)
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			return next(c)
		}
	}
}

// parseOrigins extracts scheme+host from each URL string, skipping blank or
// unparseable entries. This produces safe CSP tokens regardless of whether
// the caller passes a full path or a bare origin.
func parseOrigins(rawURLs []string) []string {
	origins := make([]string, 0, len(rawURLs))
	for _, raw := range rawURLs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			continue
		}
		origins = append(origins, u.Scheme+"://"+u.Host)
	}
	return origins
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func buildCSP(nonce string, extraImageOrigins []string) string {
	// img-src: always allow self, data URIs, and blobs; append any external
	// storage origins (e.g. "https://t3.storageapi.dev") passed at startup.
	imgSrc := strings.Join(
		append([]string{"'self'", "data:", "blob:"}, extraImageOrigins...),
		" ",
	)

	// script-src: same-origin scripts + per-request nonce for inline scripts.
	//   'unsafe-eval' is required by Datastar v1 which uses new Function() to
	//   evaluate reactive expressions (data-show, data-bind, etc.).
	//   See: assets/js/datastar.js — Function("el","$","__action","evt",...)
	// style-src: unsafe-inline retained — Tailwind utility classes are static-file
	//   but some templUI components inject inline style strings.
	// connect-src 'self': covers Datastar SSE streams.
	// worker-src 'self': allows /sw.js to be registered as a service worker.
	return strings.Join([]string{
		"default-src 'self'",
		fmt.Sprintf("script-src 'self' 'unsafe-eval' 'nonce-%s'", nonce),
		"style-src 'self' 'unsafe-inline'",
		"img-src " + imgSrc,
		"connect-src 'self'",
		"font-src 'self' data:",
		"manifest-src 'self'",
		"worker-src 'self'",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}, "; ")
}
