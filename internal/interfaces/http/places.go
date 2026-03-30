package http

import (
	"net/http"

	"bitmerchant/internal/application/places"
	"bitmerchant/internal/interfaces/templates"

	"github.com/labstack/echo/v4"
)

// PlacesHandler serves customer "My places" (visited restaurants for this session).
type PlacesHandler struct {
	listVisitedUC *places.ListVisitedRestaurantsUseCase
}

// NewPlacesHandler constructs the handler.
func NewPlacesHandler(listVisitedUC *places.ListVisitedRestaurantsUseCase) *PlacesHandler {
	return &PlacesHandler{listVisitedUC: listVisitedUC}
}

// GetMyPlaces handles GET /my-places
func (h *PlacesHandler) GetMyPlaces(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	visited, err := h.listVisitedUC.Execute(c.Request().Context(), sessionID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load places: "+err.Error())
	}
	return templates.MyPlacesPage(visited).Render(c.Request().Context(), c.Response())
}
