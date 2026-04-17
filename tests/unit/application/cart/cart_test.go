package cart_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/app/cart"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCartService(t *testing.T) {
	s := cart.NewCartService()
	sessionID := "session_1"

	// Setup mocks/repos if needed but CartService is mostly self-contained for basic ops
	// except when adding items we pass full item struct, not ID lookup inside service?
	// Wait, checking implementation of AddItem:
	// func (s *CartService) AddItem(sessionID string, item *menu.MenuItem, quantity int) error

	item, _ := menu.NewMenuItem("i1", "c1", "r1", "Burger", 10.0)

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
		require.NoError(t, s.AddItem(sessionID, item, 1))
		s.ClearCart(sessionID)

		c := s.GetCart(sessionID)
		assert.Empty(t, c.Items)
	})

	t.Run("DecrementItem reduces qty by 1", func(t *testing.T) {
		s3 := cart.NewCartService()
		sid := "session_dec"
		item3, _ := menu.NewMenuItem("id3", "c1", "r1", "Pizza", 12.0)
		require.NoError(t, s3.AddItem(sid, item3, 3))

		require.NoError(t, s3.DecrementItem(sid, "id3"))
		c := s3.GetCart(sid)
		assert.Len(t, c.Items, 1)
		assert.Equal(t, 2, c.Items[0].Quantity)
		assert.InDelta(t, 24.0, c.Total, 0.001)
	})

	t.Run("DecrementItem removes item at zero", func(t *testing.T) {
		s4 := cart.NewCartService()
		sid := "session_dec2"
		item4, _ := menu.NewMenuItem("id4", "c1", "r1", "Salad", 8.0)
		require.NoError(t, s4.AddItem(sid, item4, 1))

		require.NoError(t, s4.DecrementItem(sid, "id4"))
		c := s4.GetCart(sid)
		assert.Empty(t, c.Items)
		assert.Equal(t, 0.0, c.Total)
	})

	t.Run("switch restaurant clears cart", func(t *testing.T) {
		s2 := cart.NewCartService()
		sid := "session_switch"
		a, _ := menu.NewMenuItem("ia", "c1", "ra", "A", 5)
		b, _ := menu.NewMenuItem("ib", "c2", "rb", "B", 7)
		require.NoError(t, s2.AddItem(sid, a, 1))
		c := s2.GetCart(sid)
		assert.Equal(t, common.RestaurantID("ra"), c.RestaurantID)
		require.NoError(t, s2.AddItem(sid, b, 2))
		c = s2.GetCart(sid)
		assert.Len(t, c.Items, 1)
		assert.Equal(t, common.RestaurantID("rb"), c.RestaurantID)
		assert.Equal(t, "B", c.Items[0].Name)
		assert.Equal(t, 2, c.Items[0].Quantity)
	})
}

// Need to test infrastructure/payment/cash too?
// T048 [P] [US1] Unit tests for CashPaymentMethod
// I'll add that one too.
