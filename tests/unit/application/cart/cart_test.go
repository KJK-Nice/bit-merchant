package cart_test

import (
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCartService(t *testing.T) {
	s := cart.NewCartService()
	sessionID := "session_1"

	// Setup mocks/repos if needed but CartService is mostly self-contained for basic ops
	// except when adding items we pass full item struct, not ID lookup inside service?
	// Wait, checking implementation of AddItem:
	// func (s *CartService) AddItem(sessionID string, item *domain.MenuItem, quantity int) error

	item, _ := domain.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)

	t.Run("AddItem", func(t *testing.T) {
		err := s.AddItem(sessionID, item, 2)
		assert.NoError(t, err)

		c := s.GetCart(sessionID)
		assert.Len(t, c.Items, 1)
		assert.Equal(t, 2, c.Items[0].Quantity)
		assert.Equal(t, 20.0, c.Items[0].Subtotal)
		assert.Equal(t, 20.0, c.Total)
	})

	t.Run("AddItem existing", func(t *testing.T) {
		err := s.AddItem(sessionID, item, 1)
		assert.NoError(t, err)

		c := s.GetCart(sessionID)
		assert.Len(t, c.Items, 1)
		assert.Equal(t, 3, c.Items[0].Quantity) // 2 + 1
		assert.Equal(t, 30.0, c.Total)
	})

	t.Run("RemoveItem", func(t *testing.T) {
		err := s.RemoveItem(sessionID, "i1")
		assert.NoError(t, err)

		c := s.GetCart(sessionID)
		assert.Empty(t, c.Items)
		assert.Equal(t, 0.0, c.Total)
	})

	t.Run("ClearCart", func(t *testing.T) {
		s.AddItem(sessionID, item, 1)
		s.ClearCart(sessionID)

		c := s.GetCart(sessionID)
		assert.Empty(t, c.Items)
	})
}

// Need to test infrastructure/payment/cash too?
// T048 [P] [US1] Unit tests for CashPaymentMethod
// I'll add that one too.
