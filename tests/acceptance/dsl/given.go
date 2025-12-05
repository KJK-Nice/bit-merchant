package dsl

import (
	"bitmerchant/internal/domain"
)

// GivenBuilder provides a fluent API for setting up test data
type GivenBuilder struct {
	setup *TestSetup
}

// Restaurant sets up a restaurant
func (g *GivenBuilder) Restaurant(id domain.RestaurantID, name string, isOpen bool) *GivenBuilder {
	g.setup.restaurants = append(g.setup.restaurants, &domain.Restaurant{
		ID:     id,
		Name:   name,
		IsOpen: isOpen,
	})
	return g
}

// MenuCategory sets up a menu category
func (g *GivenBuilder) MenuCategory(id, restaurantID, name string, order int) *GivenBuilder {
	cat, _ := domain.NewMenuCategory(domain.CategoryID(id), domain.RestaurantID(restaurantID), name, order)
	g.setup.categories = append(g.setup.categories, cat)
	return g
}

// MenuItem sets up a menu item
func (g *GivenBuilder) MenuItem(id, categoryID, restaurantID, name string, price float64, available bool) *GivenBuilder {
	item, _ := domain.NewMenuItem(domain.ItemID(id), domain.CategoryID(categoryID), domain.RestaurantID(restaurantID), name, price)
	item.IsAvailable = available
	g.setup.items = append(g.setup.items, item)
	return g
}

// Order sets up an existing order
func (g *GivenBuilder) Order(id, orderNumber, restaurantID, sessionID string, status domain.PaymentStatus) *GivenBuilder {
	// Create order with default items
	items := []domain.OrderItem{}
	order, _ := domain.NewOrder(
		domain.OrderID(id),
		domain.OrderNumber(orderNumber),
		domain.RestaurantID(restaurantID),
		sessionID,
		items,
		1000,
		domain.PaymentMethodTypeCash,
	)
	order.PaymentStatus = status
	g.setup.orders = append(g.setup.orders, order)
	return g
}

// OrderWithItems sets up an order with specific items
func (g *GivenBuilder) OrderWithItems(id, orderNumber, restaurantID, sessionID string, status domain.PaymentStatus, items []domain.OrderItem) *GivenBuilder {
	order, _ := domain.NewOrder(
		domain.OrderID(id),
		domain.OrderNumber(orderNumber),
		domain.RestaurantID(restaurantID),
		sessionID,
		items,
		1000,
		domain.PaymentMethodTypeCash,
	)
	order.PaymentStatus = status
	g.setup.orders = append(g.setup.orders, order)
	return g
}

// CustomerSession sets up a customer session
func (g *GivenBuilder) CustomerSession(sessionID string) *GivenBuilder {
	g.setup.sessions = append(g.setup.sessions, sessionID)
	return g
}

// CartWithItems sets up a cart with items
func (g *GivenBuilder) CartWithItems(sessionID string, itemID string, quantity int) *GivenBuilder {
	g.setup.cartItems = append(g.setup.cartItems, CartItem{
		SessionID: sessionID,
		ItemID:    itemID,
		Quantity:  quantity,
	})
	return g
}

