package cart

import (
	"errors"

	"bitmerchant/internal/domain"
)

// RemoveFromCartUseCase removes item from cart or adjusts quantity
type RemoveFromCartUseCase struct {
	cartStore *CartStore
}

// NewRemoveFromCartUseCase creates a new RemoveFromCartUseCase
func NewRemoveFromCartUseCase(cartStore *CartStore) *RemoveFromCartUseCase {
	return &RemoveFromCartUseCase{
		cartStore: cartStore,
	}
}

// Execute removes item from cart or adjusts quantity
func (uc *RemoveFromCartUseCase) Execute(sessionID string, itemID domain.ItemID, quantity int) (*Cart, error) {
	if quantity < 0 {
		return nil, errors.New("quantity cannot be negative")
	}

	cart := uc.cartStore.GetCart(sessionID)

	// Find and update/remove item
	newItems := make([]CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item.ItemID == itemID {
			if quantity == 0 {
				// Remove item completely
				continue
			}
			// Update quantity
			item.Quantity = quantity
			item.Subtotal = float64(quantity) * item.UnitPrice
		}
		newItems = append(newItems, item)
	}

	cart.Items = newItems

	// Recalculate total
	total := 0.0
	for _, item := range cart.Items {
		total += item.Subtotal
	}
	cart.Total = total

	uc.cartStore.SaveCart(sessionID, cart)
	return cart, nil
}
