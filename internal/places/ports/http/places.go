package http

import (
	"net/http"

	"bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/interfaces/templates"
	placesQuery "bitmerchant/internal/places/app/query"

	"github.com/labstack/echo/v4"
)

// PlacesHandler serves customer "My places" (visited restaurants for this session).
type PlacesHandler struct {
	listVisitedUC placesQuery.SessionVisitedPlacesHandler
}

// NewPlacesHandler constructs the handler.
func NewPlacesHandler(listVisitedUC placesQuery.SessionVisitedPlacesHandler) *PlacesHandler {
	return &PlacesHandler{listVisitedUC: listVisitedUC}
}

// GetEntry handles GET / — renders the entry page with the user's visited places.
// The home route is AppSurfacePublic, but visited places are keyed to the customer session,
// so we prefer the customer session cookie over the public session when it exists.
func (h *PlacesHandler) GetEntry(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	if cookie, err := c.Cookie(middleware.CustomerSessionCookieName); err == nil && cookie.Value != "" {
		sessionID = cookie.Value
	}
	visited, err := h.listVisitedUC.Handle(c.Request().Context(), placesQuery.SessionVisitedPlaces{SessionID: sessionID})
	if err != nil {
		visited = nil
	}
	return templates.EntryPage(c.QueryParam("reason"), visited).Render(c.Request().Context(), c.Response())
}

// GetMyPlaces handles GET /my-places
func (h *PlacesHandler) GetMyPlaces(c echo.Context) error {
	sessionID, _ := c.Get("sessionID").(string)
	visited, err := h.listVisitedUC.Handle(c.Request().Context(), placesQuery.SessionVisitedPlaces{SessionID: sessionID})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load places: "+err.Error())
	}
	return templates.MyPlacesPage(visited).Render(c.Request().Context(), c.Response())
}

// GetScanQR handles GET /scan.
func (h *PlacesHandler) GetScanQR(c echo.Context) error {
	return templates.ScanQRPage().Render(c.Request().Context(), c.Response())
}
