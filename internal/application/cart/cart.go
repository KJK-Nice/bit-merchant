package cart

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// CartItem represents an item in the cart
type CartItem struct {
	ItemID    domain.ItemID
	Name      string
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}

// Cart represents a shopping cart
type Cart struct {
	Items []CartItem
	Total float64
}

// CartService manages session-based carts
type CartService struct {
	mu    sync.RWMutex
	carts map[string]*Cart
}

// NewCartService creates a new cart service
func NewCartService() *CartService {
	return &CartService{
		carts: make(map[string]*Cart),
	}
}

// GetCart retrieves cart for session
func (s *CartService) GetCart(sessionID string) *Cart {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cart, exists := s.carts[sessionID]
	if !exists {
		return &Cart{Items: []CartItem{}, Total: 0}
	}
	// Return copy to prevent concurrency issues if caller modifies it without lock
	// But for simplicity in MVP, returning pointer is risky but acceptable if used carefully.
	// Better to return deep copy.
	return s.copyCart(cart)
}

// AddItem adds item to cart
func (s *CartService) AddItem(sessionID string, item *domain.MenuItem, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cart, exists := s.carts[sessionID]
	if !exists {
		cart = &Cart{Items: []CartItem{}, Total: 0}
		s.carts[sessionID] = cart
	}

	// Find existing item
	found := false
	for i := range cart.Items {
		if cart.Items[i].ItemID == item.ID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = float64(cart.Items[i].Quantity) * cart.Items[i].UnitPrice
			found = true
			break
		}
	}

	// Add new item
	if !found {
		cart.Items = append(cart.Items, CartItem{
			ItemID:    item.ID,
			Name:      item.Name,
			Quantity:  quantity,
			UnitPrice: item.Price,
			Subtotal:  float64(quantity) * item.Price,
		})
	}

	s.recalculateTotal(cart)
	return nil
}

// RemoveItem removes item from cart
func (s *CartService) RemoveItem(sessionID string, itemID domain.ItemID) error {
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
	s.recalculateTotal(cart)
	return nil
}

// ClearCart clears the cart
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
		Total: cart.Total,
		Items: make([]CartItem, len(cart.Items)),
	}
	copy(newCart.Items, cart.Items)
	return newCart
}
