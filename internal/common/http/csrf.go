package commonhttp

import "github.com/labstack/echo/v4"

// CSRFToken returns the CSRF token from Echo context (set by middleware).
func CSRFToken(c echo.Context) string {
	token, _ := c.Get("csrf").(string)
	return token
}
