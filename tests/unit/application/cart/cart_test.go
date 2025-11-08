package cart_test

import (
	"testing"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/domain"
)

func TestCartStore_GetCart(t *testing.T) {
	store := cart.NewCartStore()
	sessionID := "session_001"

	// Get empty cart
	cart := store.GetCart(sessionID)
	if cart == nil {
		t.Fatal("GetCart() returned nil")
	}
	if len(cart.Items) != 0 {
		t.Errorf("GetCart() returned cart with %d items, want 0", len(cart.Items))
	}
	if cart.Total != 0 {
		t.Errorf("GetCart() returned cart with total %f, want 0", cart.Total)
	}
}

func TestCartStore_SaveCart(t *testing.T) {
	store := cart.NewCartStore()
	sessionID := "session_001"

	cart := &cart.Cart{
		Items: []cart.CartItem{
			{
				ItemID:    "item_001",
				Name:      "Test Item",
				Quantity:  2,
				UnitPrice: 10.99,
				Subtotal:  21.98,
			},
		},
		Total: 21.98,
	}

	store.SaveCart(sessionID, cart)

	retrieved := store.GetCart(sessionID)
	if len(retrieved.Items) != 1 {
		t.Errorf("GetCart() returned %d items, want 1", len(retrieved.Items))
	}
	if retrieved.Total != 21.98 {
		t.Errorf("GetCart() returned total %f, want 21.98", retrieved.Total)
	}
}

func TestAddToCartUseCase_Execute(t *testing.T) {
	store := cart.NewCartStore()
	itemRepo := &mockMenuItemRepository{
		items: map[domain.ItemID]*domain.MenuItem{
			"item_001": {
				ID:           "item_001",
				Name:         "Test Item",
				Price:        10.99,
				IsAvailable:  true,
				RestaurantID: "rest_001",
			},
		},
	}
	restaurantRepo := &mockRestaurantRepository{
		restaurants: map[domain.RestaurantID]*domain.Restaurant{
			"rest_001": {
				ID:     "rest_001",
				IsOpen: true,
			},
		},
	}

	useCase := cart.NewAddToCartUseCase(store, itemRepo, restaurantRepo)

	// Add item to cart
	cart, err := useCase.Execute("session_001", "item_001", 2)
	if err != nil {
		t.Fatalf("AddToCart() error = %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("AddToCart() returned cart with %d items, want 1", len(cart.Items))
	}
	if cart.Total != 21.98 {
		t.Errorf("AddToCart() returned total %f, want 21.98", cart.Total)
	}

	// Add same item again (should increase quantity)
	cart, err = useCase.Execute("session_001", "item_001", 1)
	if err != nil {
		t.Fatalf("AddToCart() error = %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("AddToCart() returned cart with %d items, want 1", len(cart.Items))
	}
	if cart.Items[0].Quantity != 3 {
		t.Errorf("AddToCart() returned quantity %d, want 3", cart.Items[0].Quantity)
	}
}

func TestAddToCartUseCase_InvalidQuantity(t *testing.T) {
	store := cart.NewCartStore()
	itemRepo := &mockMenuItemRepository{}
	restaurantRepo := &mockRestaurantRepository{}

	useCase := cart.NewAddToCartUseCase(store, itemRepo, restaurantRepo)

	_, err := useCase.Execute("session_001", "item_001", 0)
	if err == nil {
		t.Error("AddToCart() with quantity 0 should return error")
	}

	_, err = useCase.Execute("session_001", "item_001", -1)
	if err == nil {
		t.Error("AddToCart() with negative quantity should return error")
	}
}

func TestRemoveFromCartUseCase_Execute(t *testing.T) {
	store := cart.NewCartStore()
	store.SaveCart("session_001", &cart.Cart{
		Items: []cart.CartItem{
			{ItemID: "item_001", Name: "Item 1", Quantity: 2, UnitPrice: 10.99, Subtotal: 21.98},
			{ItemID: "item_002", Name: "Item 2", Quantity: 1, UnitPrice: 5.99, Subtotal: 5.99},
		},
		Total: 27.97,
	})

	useCase := cart.NewRemoveFromCartUseCase(store)

	// Remove item completely
	cart, err := useCase.Execute("session_001", "item_001", 0)
	if err != nil {
		t.Fatalf("RemoveFromCart() error = %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("RemoveFromCart() returned %d items, want 1", len(cart.Items))
	}
	if cart.Items[0].ItemID != "item_002" {
		t.Errorf("RemoveFromCart() kept wrong item")
	}

	// Update quantity
	cart, err = useCase.Execute("session_001", "item_002", 3)
	if err != nil {
		t.Fatalf("RemoveFromCart() error = %v", err)
	}
	if cart.Items[0].Quantity != 3 {
		t.Errorf("RemoveFromCart() returned quantity %d, want 3", cart.Items[0].Quantity)
	}
}

func TestGetCartUseCase_Execute(t *testing.T) {
	store := cart.NewCartStore()
	store.SaveCart("session_001", &cart.Cart{
		Items: []cart.CartItem{
			{ItemID: "item_001", Name: "Item 1", Quantity: 1, UnitPrice: 10.99, Subtotal: 10.99},
		},
		Total: 10.99,
	})

	useCase := cart.NewGetCartUseCase(store)
	cart := useCase.Execute("session_001")

	if len(cart.Items) != 1 {
		t.Errorf("GetCart() returned %d items, want 1", len(cart.Items))
	}
	if cart.Total != 10.99 {
		t.Errorf("GetCart() returned total %f, want 10.99", cart.Total)
	}
}

// Mock repositories for testing
type mockMenuItemRepository struct {
	items map[domain.ItemID]*domain.MenuItem
}

func (m *mockMenuItemRepository) FindByID(id domain.ItemID) (*domain.MenuItem, error) {
	item, exists := m.items[id]
	if !exists {
		return nil, &mockError{msg: "item not found"}
	}
	return item, nil
}

func (m *mockMenuItemRepository) Save(*domain.MenuItem) error { return nil }
func (m *mockMenuItemRepository) FindByCategoryID(domain.CategoryID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) FindByRestaurantID(domain.RestaurantID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) FindAvailableByRestaurantID(domain.RestaurantID) ([]*domain.MenuItem, error) {
	return nil, nil
}
func (m *mockMenuItemRepository) Update(*domain.MenuItem) error                        { return nil }
func (m *mockMenuItemRepository) Delete(domain.ItemID) error                           { return nil }
func (m *mockMenuItemRepository) CountByRestaurantID(domain.RestaurantID) (int, error) { return 0, nil }

type mockRestaurantRepository struct {
	restaurants map[domain.RestaurantID]*domain.Restaurant
}

func (m *mockRestaurantRepository) FindByID(id domain.RestaurantID) (*domain.Restaurant, error) {
	restaurant, exists := m.restaurants[id]
	if !exists {
		return nil, &mockError{msg: "restaurant not found"}
	}
	return restaurant, nil
}

func (m *mockRestaurantRepository) Save(*domain.Restaurant) error   { return nil }
func (m *mockRestaurantRepository) Update(*domain.Restaurant) error { return nil }

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
