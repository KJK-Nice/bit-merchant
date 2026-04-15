package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	ContextHostSurface  = "hostSurface"
	ContextRouteSurface = "routeSurface"
)

type AppSurface string

const (
	AppSurfacePublic   AppSurface = "public"
	AppSurfaceCustomer AppSurface = "customer"
	AppSurfaceMerchant AppSurface = "merchant"
)

type SurfaceConfig struct {
	PublicBaseURL   string
	CustomerBaseURL string
	MerchantBaseURL string
}

type surfaceRoutingConfig struct {
	publicURL   *url.URL
	customerURL *url.URL
	merchantURL *url.URL
}

func SurfaceRoutingMiddleware(cfg SurfaceConfig) echo.MiddlewareFunc {
	parsed := surfaceRoutingConfig{
		publicURL:   mustParseURL(cfg.PublicBaseURL),
		customerURL: mustParseURL(cfg.CustomerBaseURL),
		merchantURL: mustParseURL(cfg.MerchantBaseURL),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			routeSurface := classifyRouteSurface(path)
			hostSurface := parsed.classifyHostSurface(c.Request().Host, routeSurface)

			c.Set(ContextRouteSurface, routeSurface)
			c.Set(ContextHostSurface, hostSurface)

			if routeSurface == AppSurfacePublic || routeSurface == hostSurface {
				return next(c)
			}

			targetBase := parsed.baseURLForSurface(routeSurface)
			if targetBase == "" {
				return c.NoContent(http.StatusNotFound)
			}

			target := BuildCanonicalURL(targetBase, c.Request().URL.Path, c.Request().URL.RawQuery)
			if target == "" || isSameRequestTarget(c.Request(), target) {
				return next(c)
			}
			return c.Redirect(http.StatusFound, target)
		}
	}
}

func classifyRouteSurface(path string) AppSurface {
	if hasAnyPathPrefix(path, "/menu", "/my-places", "/scan", "/cart", "/order") {
		return AppSurfaceCustomer
	}
	if hasAnyPathPrefix(path, "/dashboard", "/admin", "/kitchen", "/auth", "/owner") {
		return AppSurfaceMerchant
	}
	return AppSurfacePublic
}

func hasAnyPathPrefix(path string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if hasPathPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func hasPathPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}
	return u
}

func normalizeHost(host string) string {
	return strings.ToLower(strings.TrimSpace(host))
}

func hostMatches(reqHost string, target *url.URL) bool {
	if target == nil {
		return false
	}
	req := normalizeHost(reqHost)
	cfg := normalizeHost(target.Host)
	if req == cfg {
		return true
	}
	reqName := req
	if h, _, err := splitHostPort(req); err == nil {
		reqName = h
	}
	cfgName := cfg
	if h, _, err := splitHostPort(cfg); err == nil {
		cfgName = h
	}
	return reqName == cfgName
}

func splitHostPort(host string) (string, string, error) {
	if strings.Count(host, ":") == 0 {
		return "", "", net.InvalidAddrError("missing port in address")
	}
	return net.SplitHostPort(host)
}

func (cfg surfaceRoutingConfig) classifyHostSurface(reqHost string, routeSurface AppSurface) AppSurface {
	matchesCustomer := hostMatches(reqHost, cfg.customerURL)
	matchesMerchant := hostMatches(reqHost, cfg.merchantURL)
	matchesPublic := hostMatches(reqHost, cfg.publicURL)

	switch {
	case matchesCustomer && matchesMerchant:
		if routeSurface == AppSurfaceMerchant || routeSurface == AppSurfaceCustomer {
			return routeSurface
		}
		if matchesPublic {
			return AppSurfacePublic
		}
		return AppSurfaceCustomer
	case matchesCustomer:
		return AppSurfaceCustomer
	case matchesMerchant:
		return AppSurfaceMerchant
	case matchesPublic:
		return AppSurfacePublic
	default:
		return AppSurfacePublic
	}
}

func (cfg surfaceRoutingConfig) baseURLForSurface(surface AppSurface) string {
	switch surface {
	case AppSurfaceCustomer:
		if cfg.customerURL != nil {
			return cfg.customerURL.String()
		}
	case AppSurfaceMerchant:
		if cfg.merchantURL != nil {
			return cfg.merchantURL.String()
		}
	default:
		if cfg.publicURL != nil {
			return cfg.publicURL.String()
		}
	}
	return ""
}

func BuildCanonicalURL(baseURL, path, rawQuery string) string {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return ""
	}
	u := &url.URL{
		Scheme:   base.Scheme,
		Host:     base.Host,
		Path:     path,
		RawQuery: rawQuery,
	}
	return u.String()
}

func isSameRequestTarget(r *http.Request, target string) bool {
	tu, err := url.Parse(target)
	if err != nil {
		return false
	}
	return normalizeHost(r.Host) == normalizeHost(tu.Host) && r.URL.Path == tu.Path && r.URL.RawQuery == tu.RawQuery
}
