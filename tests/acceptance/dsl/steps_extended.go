package dsl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// ViewMenuStep represents viewing the menu
type ViewMenuStep struct {
	SessionID string
}

func (s *ViewMenuStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	if s.SessionID != "" {
		req.AddCookie(&http.Cookie{Name: "bitmerchant_session", Value: s.SessionID})
	}

	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	if s.SessionID != "" {
		c.Set("sessionID", s.SessionID)
	}

	err := app.menuHandler.GetMenu(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

// AddMultipleItemsStep represents adding multiple items to cart
type AddMultipleItemsStep struct {
	SessionID string
	Items     []struct {
		ItemID   string
		Quantity int
	}
}

func (s *AddMultipleItemsStep) Execute(t *testing.T, app *TestApplication) {
	for _, item := range s.Items {
		step := &AddToCartStep{
			SessionID: s.SessionID,
			ItemID:    item.ItemID,
			Quantity:  item.Quantity,
		}
		step.Execute(t, app)
	}
}

// ViewOrderHistoryStep represents viewing customer order history
type ViewOrderHistoryStep struct {
	SessionID string
}

func (s *ViewOrderHistoryStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/order/history", nil)
	req.AddCookie(&http.Cookie{Name: "bitmerchant_session", Value: s.SessionID})

	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)
	c.Set("sessionID", s.SessionID)

	err := app.orderHandler.GetLookup(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

// ViewDashboardStep represents viewing the dashboard
type ViewDashboardStep struct{}

func (s *ViewDashboardStep) Execute(t *testing.T, app *TestApplication) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := app.echo.NewContext(req, rec)

	err := app.dashboardHandler.Dashboard(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}
