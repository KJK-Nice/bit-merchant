package cart

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/menu/domain/menu"
)

// CartItemModifier captures a single selected option from an option group.
type CartItemModifier struct {
	GroupID    string
	GroupName  string
	OptionID   string
	OptionName string
	PriceDelta float64
}

// CartItem represents an item in the cart.
type CartItem struct {
	ItemID              common.ItemID
	Name                string
	Quantity            int
	UnitPrice           float64 // base item price (without modifiers)
	ModifierPrice       float64 // sum of selected modifier PriceDeltas
	Subtotal            float64 // (UnitPrice + ModifierPrice) * Quantity
	Modifiers           []CartItemModifier
	SpecialInstructions string
}

// Cart represents a shopping cart. Currency is set from the first item added
// (all items in a cart share the restaurant's base currency).
type Cart struct {
	RestaurantID common.RestaurantID
	Items        []CartItem
	Total        float64
	Currency     money.Currency
}

// Money returns the cart total as money.Money.
func (c *Cart) Money() money.Money {
	cur := c.Currency
	if cur.IsZero() {
		cur = money.USD
	}
	return money.FromMajor(c.Total, cur)
}

// CartService manages session-based carts.
type CartService struct {
	mu    sync.RWMutex
	carts map[string]*Cart
}

func NewCartService() *CartService {
	return &CartService{
		carts: make(map[string]*Cart),
	}
}

func (s *CartService) GetCart(sessionID string) *Cart {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cart, exists := s.carts[sessionID]
	if !exists {
		return &Cart{Items: []CartItem{}, Total: 0, RestaurantID: ""}
	}
	return s.copyCart(cart)
}

// AddItem adds an item to the cart without modifiers.
func (s *CartService) AddItem(sessionID string, item *menu.MenuItem, quantity int) error {
	return s.AddItemWithModifiers(sessionID, item, quantity, nil, "")
}

// AddItemWithModifiers adds an item with selected modifier options and a special note.
// If the item is already in the cart (same ItemID), quantity increments; modifiers are
// kept from the first add (first-add-wins for modifier snapshot).
func (s *CartService) AddItemWithModifiers(sessionID string, item *menu.MenuItem, quantity int, modifiers []CartItemModifier, specialInstructions string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cart := s.getOrCreateCart(sessionID)
	s.clearIfRestaurantChanged(cart, item.RestaurantID)

	modifierTotal := sumModifierPrices(modifiers)
	if !s.incrementExisting(cart, item.ID, quantity) {
		cart.Items = append(cart.Items, newCartItem(item, quantity, modifiers, modifierTotal, specialInstructions))
	}

	cart.RestaurantID = item.RestaurantID
	if !item.Currency.IsZero() {
		cart.Currency = item.Currency
	}
	s.recalculateTotal(cart)
	return nil
}

func (s *CartService) getOrCreateCart(sessionID string) *Cart {
	cart, exists := s.carts[sessionID]
	if !exists {
		cart = &Cart{Items: []CartItem{}, Total: 0, RestaurantID: ""}
		s.carts[sessionID] = cart
	}
	return cart
}

func (s *CartService) clearIfRestaurantChanged(cart *Cart, restaurantID common.RestaurantID) {
	if len(cart.Items) > 0 && cart.RestaurantID != "" && cart.RestaurantID != restaurantID {
		cart.Items = nil
		cart.Total = 0
		cart.RestaurantID = ""
		cart.Currency = money.Currency{}
	}
}

func (s *CartService) incrementExisting(cart *Cart, itemID common.ItemID, quantity int) bool {
	for i := range cart.Items {
		if cart.Items[i].ItemID == itemID {
			cart.Items[i].Quantity += quantity
			effPrice := cart.Items[i].UnitPrice + cart.Items[i].ModifierPrice
			cart.Items[i].Subtotal = float64(cart.Items[i].Quantity) * effPrice
			return true
		}
	}
	return false
}

func sumModifierPrices(modifiers []CartItemModifier) float64 {
	total := 0.0
	for _, m := range modifiers {
		total += m.PriceDelta
	}
	return total
}

func newCartItem(item *menu.MenuItem, quantity int, modifiers []CartItemModifier, modifierTotal float64, specialInstructions string) CartItem {
	effPrice := item.Price + modifierTotal
	return CartItem{
		ItemID:              item.ID,
		Name:                item.Name,
		Quantity:            quantity,
		UnitPrice:           item.Price,
		ModifierPrice:       modifierTotal,
		Subtotal:            float64(quantity) * effPrice,
		Modifiers:           modifiers,
		SpecialInstructions: specialInstructions,
	}
}

func (s *CartService) RemoveItem(sessionID string, itemID common.ItemID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[sessionID]
	if !exists {
		return nil
	}

	newItems := []CartItem{}
	for _, item := range cart.Items {
		if item.ItemID != itemID {
			newItems = append(newItems, item)
		}
	}
	cart.Items = newItems
	if len(cart.Items) == 0 {
		cart.RestaurantID = ""
		cart.Currency = money.Currency{}
	}
	s.recalculateTotal(cart)
	return nil
}

// DecrementItem reduces an item's quantity by 1. If the quantity reaches 0, the item is removed.
func (s *CartService) DecrementItem(sessionID string, itemID common.ItemID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[sessionID]
	if !exists {
		return nil
	}

	newItems := []CartItem{}
	for _, item := range cart.Items {
		if item.ItemID == itemID {
			item.Quantity--
			if item.Quantity <= 0 {
				continue
			}
			item.Subtotal = float64(item.Quantity) * (item.UnitPrice + item.ModifierPrice)
		}
		newItems = append(newItems, item)
	}
	cart.Items = newItems
	if len(cart.Items) == 0 {
		cart.RestaurantID = ""
		cart.Currency = money.Currency{}
	}
	s.recalculateTotal(cart)
	return nil
}

func (s *CartService) ClearCart(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.carts, sessionID)
}

func (s *CartService) recalculateTotal(cart *Cart) {
	total := 0.0
	for _, item := range cart.Items {
		total += item.Subtotal
	}
	cart.Total = total
}

func (s *CartService) copyCart(cart *Cart) *Cart {
	newCart := &Cart{
		RestaurantID: cart.RestaurantID,
		Total:        cart.Total,
		Currency:     cart.Currency,
		Items:        make([]CartItem, len(cart.Items)),
	}
	copy(newCart.Items, cart.Items)
	return newCart
}
