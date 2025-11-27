package http

import (
	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/interfaces/templates"
	"net/http"

	"github.com/labstack/echo/v4"
)

type OwnerHandler struct {
	createRestaurantUC *restaurant.CreateRestaurantUseCase
}

func NewOwnerHandler(createRestaurantUC *restaurant.CreateRestaurantUseCase) *OwnerHandler {
	return &OwnerHandler{
		createRestaurantUC: createRestaurantUC,
	}
}

// GetSignup handles GET /owner/signup
func (h *OwnerHandler) GetSignup(c echo.Context) error {
	return templates.OwnerSignup().Render(c.Request().Context(), c.Response())
}

// PostSignup handles POST /owner/signup
func (h *OwnerHandler) PostSignup(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "Restaurant name is required")
	}

	req := restaurant.CreateRestaurantRequest{
		Name: name,
	}

	rest, err := h.createRestaurantUC.Execute(c.Request().Context(), req)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create restaurant: "+err.Error())
	}

	// Redirect to dashboard menu after signup
	return c.Redirect(http.StatusFound, "/dashboard/menu?restaurant_id="+string(rest.ID))
}

