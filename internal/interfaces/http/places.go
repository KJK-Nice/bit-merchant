package http

import (
	"bitmerchant/internal/interfaces/templates"
	placesQuery "bitmerchant/internal/places/app/query"

	"github.com/labstack/echo/v4"
	"net/http"
)

// PlacesHandler serves customer "My places" (visited restaurants for this session).
type PlacesHandler struct {
	listVisitedUC *placesQuery.ListVisitedRestaurantsUseCase
}

// NewPlacesHandler constructs the handler.
func NewPlacesHandler(listVisitedUC *placesQuery.ListVisitedRestaurantsUseCase) *PlacesHandler {
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

// GetScanQR handles GET /scan.
func (h *PlacesHandler) GetScanQR(c echo.Context) error {
	return templates.ScanQRPage().Render(c.Request().Context(), c.Response())
}
