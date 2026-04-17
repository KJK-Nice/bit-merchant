package http

import (
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/interfaces/templates"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"

	"github.com/labstack/echo/v4"
	"net/http"
)

type OwnerHandler struct {
	createRestaurantUC restaurantCmd.CreateRestaurantHandler
}

func NewOwnerHandler(createRestaurantUC restaurantCmd.CreateRestaurantHandler) *OwnerHandler {
	return &OwnerHandler{
		createRestaurantUC: createRestaurantUC,
	}
}

// GetSignup handles GET /owner/signup
func (h *OwnerHandler) GetSignup(c echo.Context) error {
	return templates.OwnerSignup(commonhttp.CSRFToken(c)).Render(c.Request().Context(), c.Response())
}

// PostSignup handles POST /owner/signup
func (h *OwnerHandler) PostSignup(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "Restaurant name is required")
	}

	_, err := h.createRestaurantUC.Handle(c.Request().Context(), restaurantCmd.CreateRestaurant{
		Name: name,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create restaurant: "+err.Error())
	}

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}
