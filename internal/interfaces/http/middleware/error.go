package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler handles HTTP errors
func ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err := c.NoContent(code)
			if err != nil {
				c.Logger().Error(err)
			}
		} else {
			err := c.JSON(code, map[string]interface{}{
				"error": message,
			})
			if err != nil {
				c.Logger().Error(err)
			}
		}
	}
}
