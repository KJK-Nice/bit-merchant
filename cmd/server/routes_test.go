package main

import (
	"bitmerchant/internal/infrastructure/repositories/memory"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootRouteRendersEntryPage(t *testing.T) {
	e := echo.New()
	registerRoutes(e, routeHandlers{}, memory.NewMemoryMembershipRepository())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Scan a table QR")
}
