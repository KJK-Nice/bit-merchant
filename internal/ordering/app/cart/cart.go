package cart

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

// CartItem represents an item in the cart.
type CartItem struct {
	ItemID    common.ItemID
	Name      string
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}

// Cart represents a shopping cart.
type Cart struct {
	RestaurantID common.RestaurantID
	Items        []CartItem
	Total        float64
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

func (s *CartService) AddItem(sessionID string, item *menu.MenuItem, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[sessionID]
	if !exists {
		cart = &Cart{Items: []CartItem{}, Total: 0, RestaurantID: ""}
		s.carts[sessionID] = cart
	}

	if len(cart.Items) > 0 && cart.RestaurantID != "" && cart.RestaurantID != item.RestaurantID {
		cart.Items = nil
		cart.Total = 0
		cart.RestaurantID = ""
	}

	found := false
	for i := range cart.Items {
		if cart.Items[i].ItemID == item.ID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = float64(cart.Items[i].Quantity) * cart.Items[i].UnitPrice
			found = true
			break
		}
	}

	if !found {
		cart.Items = append(cart.Items, CartItem{
			ItemID:    item.ID,
			Name:      item.Name,
			Quantity:  quantity,
			UnitPrice: item.Price,
			Subtotal:  float64(quantity) * item.Price,
		})
	}

	cart.RestaurantID = item.RestaurantID
	s.recalculateTotal(cart)
	return nil
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
			item.Subtotal = float64(item.Quantity) * item.UnitPrice
		}
		newItems = append(newItems, item)
	}
	cart.Items = newItems
	if len(cart.Items) == 0 {
		cart.RestaurantID = ""
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
		Items:        make([]CartItem, len(cart.Items)),
	}
	copy(newCart.Items, cart.Items)
	return newCart
}
