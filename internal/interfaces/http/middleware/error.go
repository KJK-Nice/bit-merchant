package middleware

import (
	"net/http"

	"bitmerchant/internal/interfaces/templates/errors"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// ErrorHandler handles HTTP errors
func ErrorHandler(err error, c echo.Context) {
	code, message := resolveHTTPError(err)
	if c.Response().Committed {
		return
	}
	if c.Request().Method == http.MethodHead {
		if writeErr := c.NoContent(code); writeErr != nil {
			c.Logger().Error(writeErr)
		}
		return
	}
	c.Response().WriteHeader(code)
	if writeErr := selectErrorComponent(code, message).Render(c.Request().Context(), c.Response()); writeErr != nil {
		c.Logger().Error(writeErr)
	}
}

func resolveHTTPError(err error) (int, string) {
	he, ok := err.(*echo.HTTPError)
	if !ok {
		return http.StatusInternalServerError, "Internal Server Error"
	}
	switch msg := he.Message.(type) {
	case string:
		return he.Code, msg
	case error:
		return he.Code, msg.Error()
	default:
		return he.Code, http.StatusText(he.Code)
	}
}

func selectErrorComponent(code int, message string) templ.Component {
	switch code {
	case http.StatusNotFound:
		return errors.NotFound()
	case http.StatusBadRequest:
		return errors.BadRequest(message)
	default:
		return errors.ServiceUnavailable(message)
	}
}
