package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitmerchant/internal/application/places"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMyPlaces(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	orders := memory.NewMemoryOrderRepository()

	r, _ := domain.NewRestaurant("pr-contract", "Contract Cafe")
	require.NoError(t, rest.Save(r))
	require.NoError(t, visits.Upsert(&domain.SessionRestaurantVisit{
		SessionID:    "sess-p",
		RestaurantID: "pr-contract",
	}))

	uc := places.NewListVisitedRestaurantsUseCase(visits, rest, orders)
	h := handler.NewPlacesHandler(uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/my-places", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "sess-p")

	require.NoError(t, h.GetMyPlaces(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contract Cafe")
}
