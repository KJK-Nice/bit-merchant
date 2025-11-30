package middleware

import (
	"net/http"

	"bitmerchant/internal/interfaces/templates/errors"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// ErrorHandler handles HTTP errors
func ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		} else if msgErr, ok := he.Message.(error); ok {
			message = msgErr.Error()
		} else {
			message = http.StatusText(code)
		}
	}

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err := c.NoContent(code)
			if err != nil {
				c.Logger().Error(err)
			}
		} else {
			c.Response().WriteHeader(code)

			var component templ.Component
			switch code {
			case http.StatusNotFound:
				component = errors.NotFound()
			case http.StatusBadRequest:
				component = errors.BadRequest(message)
			case http.StatusServiceUnavailable:
				component = errors.ServiceUnavailable(message)
			default:
				// For 500 and other errors, use the ServiceUnavailable template as a generic error page
				// or we could create a GenericError template.
				// Given the task list only specified 404, 400, 503, let's use 503 style for generic errors
				// but maybe with a generic message if it's 500?
				// The implementation above passes 'message' which is "Internal Server Error" for 500.
				component = errors.ServiceUnavailable(message)
			}

			err := component.Render(c.Request().Context(), c.Response())
			if err != nil {
				c.Logger().Error(err)
			}
		}
	}
}
