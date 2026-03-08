package http

import "github.com/labstack/echo/v4"

func getCSRFToken(c echo.Context) string {
	token, _ := c.Get("csrf").(string)
	return token
}
