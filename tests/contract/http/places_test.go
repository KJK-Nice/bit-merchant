package http_test

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	placesQuery "bitmerchant/internal/places/app/query"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMyPlaces(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	orders := memory.NewMemoryOrderRepository()

	r, _ := restaurant.NewRestaurant("pr-contract", "Contract Cafe")
	require.NoError(t, rest.Save(r))
	require.NoError(t, visits.Upsert(context.Background(), visit.NewSessionRestaurantVisit("sess-p", "pr-contract", time.Time{}, time.Time{})))

	uc := placesQuery.NewListVisitedRestaurantsUseCase(visits, rest, orders)
	h := handler.NewPlacesHandler(uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/my-places", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("sessionID", "sess-p")

	require.NoError(t, h.GetMyPlaces(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Contract Cafe")
	assert.Contains(t, body, "Open")
}
