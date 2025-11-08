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

// CartStore manages session-based carts
type CartStore struct {
	mu    sync.RWMutex
	carts map[string]*Cart
}

// NewCartStore creates a new cart store
func NewCartStore() *CartStore {
	return &CartStore{
		carts: make(map[string]*Cart),
	}
}

// GetCart retrieves cart for session
func (s *CartStore) GetCart(sessionID string) *Cart {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cart, exists := s.carts[sessionID]
	if !exists {
		return &Cart{Items: []CartItem{}, Total: 0}
	}
	return cart
}

// SaveCart saves cart for session
func (s *CartStore) SaveCart(sessionID string, cart *Cart) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.carts[sessionID] = cart
}

// AddToCartUseCase adds item to cart
type AddToCartUseCase struct {
	cartStore      *CartStore
	itemRepo       domain.MenuItemRepository
	restaurantRepo domain.RestaurantRepository
}

// NewAddToCartUseCase creates a new AddToCartUseCase
func NewAddToCartUseCase(
	cartStore *CartStore,
	itemRepo domain.MenuItemRepository,
	restaurantRepo domain.RestaurantRepository,
) *AddToCartUseCase {
	return &AddToCartUseCase{
		cartStore:      cartStore,
		itemRepo:       itemRepo,
		restaurantRepo: restaurantRepo,
	}
}

// Execute adds item to cart
func (uc *AddToCartUseCase) Execute(sessionID string, itemID domain.ItemID, quantity int) (*Cart, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	item, err := uc.itemRepo.FindByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	if !item.IsAvailable {
		return nil, errors.New("item is not available")
	}

	restaurant, err := uc.restaurantRepo.FindByID(item.RestaurantID)
	if err != nil {
		return nil, errors.New("restaurant not found")
	}

	if !restaurant.IsOpen {
		return nil, errors.New("restaurant is closed")
	}

	cart := uc.cartStore.GetCart(sessionID)

	// Find existing item in cart
	found := false
	for i := range cart.Items {
		if cart.Items[i].ItemID == itemID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = float64(cart.Items[i].Quantity) * cart.Items[i].UnitPrice
			found = true
			break
		}
	}

	// Add new item if not found
	if !found {
		cart.Items = append(cart.Items, CartItem{
			ItemID:    itemID,
			Name:      item.Name,
			Quantity:  quantity,
			UnitPrice: item.Price,
			Subtotal:  float64(quantity) * item.Price,
		})
	}

	// Recalculate total
	total := 0.0
	for _, item := range cart.Items {
		total += item.Subtotal
	}
	cart.Total = total

	uc.cartStore.SaveCart(sessionID, cart)
	return cart, nil
}
