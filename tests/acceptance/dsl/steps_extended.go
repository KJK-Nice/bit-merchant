package dsl

import (
	"testing"
)

// ViewMenuStep represents viewing the menu
type ViewMenuStep struct {
	SessionID string
}

func (s *ViewMenuStep) Execute(t *testing.T, app *TestApplication) {
	// Navigate to menu page via browser
	app.NavigateTo("/menu")

	// Set session cookie if provided
	if s.SessionID != "" {
		app.SetCookie("bitmerchant_session", s.SessionID)
		app.ReloadPage()
	}
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
	// Navigate to order history page
	app.NavigateTo("/order/lookup")

	// Set session cookie
	app.SetCookie("bitmerchant_session", s.SessionID)
	app.ReloadPage()
}

// ViewDashboardStep represents viewing the dashboard
type ViewDashboardStep struct{}

func (s *ViewDashboardStep) Execute(t *testing.T, app *TestApplication) {
	// Navigate to dashboard page
	app.NavigateTo("/dashboard")
}

// ViewOrderConfirmationStep represents viewing the order confirmation page
type ViewOrderConfirmationStep struct {
	SessionID string
}

func (s *ViewOrderConfirmationStep) Execute(t *testing.T, app *TestApplication) {
	// Navigate to order confirmation page
	app.NavigateTo("/order/confirm")

	// Set session cookie
	app.SetCookie("bitmerchant_session", s.SessionID)
	app.ReloadPage()
}
