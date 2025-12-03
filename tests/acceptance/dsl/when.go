package dsl

// WhenBuilder provides a fluent API for defining actions
type WhenBuilder struct {
	scenario *Scenario
}

// Customer creates a customer actor
func (w *WhenBuilder) Customer(sessionID string) *CustomerActor {
	return &CustomerActor{
		scenario:  w.scenario,
		sessionID: sessionID,
	}
}

// Kitchen creates a kitchen actor
func (w *WhenBuilder) Kitchen() *KitchenActor {
	return &KitchenActor{scenario: w.scenario}
}

// System creates a system actor (for event-driven testing)
func (w *WhenBuilder) System() *SystemActor {
	return &SystemActor{scenario: w.scenario}
}

// ConnectsToSSE connects to an SSE stream
func (w *WhenBuilder) ConnectsToSSE(path string) *SSEActor {
	return &SSEActor{
		scenario: w.scenario,
		path:     path,
	}
}

// SSEActor represents SSE connection actions
type SSEActor struct {
	scenario *Scenario
	path     string
}

// Stream connects to the SSE stream
func (s *SSEActor) Stream() *SSEActor {
	s.scenario.addStep(&SSEStep{
		Path: s.path,
	})
	return s
}

// CustomerActor represents customer actions
type CustomerActor struct {
	scenario  *Scenario
	sessionID string
}

// AddsToCart adds an item to the customer's cart
func (c *CustomerActor) AddsToCart(itemID string, quantity int) *CustomerActor {
	c.scenario.addStep(&AddToCartStep{
		SessionID: c.sessionID,
		ItemID:    itemID,
		Quantity:  quantity,
	})
	return c
}

// CreatesOrder creates an order from the customer's cart
func (c *CustomerActor) CreatesOrder() *CustomerActor {
	c.scenario.addStep(&CreateOrderStep{
		SessionID: c.sessionID,
	})
	return c
}

// ViewsOrder views an order by order number
func (c *CustomerActor) ViewsOrder(orderNumber string) *CustomerActor {
	c.scenario.addStep(&ViewOrderStep{
		SessionID:   c.sessionID,
		OrderNumber: orderNumber,
	})
	return c
}

// ViewsMenu views the menu
func (c *CustomerActor) ViewsMenu() *CustomerActor {
	c.scenario.addStep(&ViewMenuStep{
		SessionID: c.sessionID,
	})
	return c
}

// AddsMultipleItems adds multiple items to cart
func (c *CustomerActor) AddsMultipleItems(items map[string]int) *CustomerActor {
	itemsList := []struct {
		ItemID   string
		Quantity int
	}{}
	for itemID, quantity := range items {
		itemsList = append(itemsList, struct {
			ItemID   string
			Quantity int
		}{ItemID: itemID, Quantity: quantity})
	}
	c.scenario.addStep(&AddMultipleItemsStep{
		SessionID: c.sessionID,
		Items:     itemsList,
	})
	return c
}

// ViewsOrderHistory views order history
func (c *CustomerActor) ViewsOrderHistory() *CustomerActor {
	c.scenario.addStep(&ViewOrderHistoryStep{
		SessionID: c.sessionID,
	})
	return c
}

// KitchenActor represents kitchen actions
type KitchenActor struct {
	scenario *Scenario
}

// ViewsDashboard views the kitchen dashboard
func (k *KitchenActor) ViewsDashboard() *KitchenActor {
	k.scenario.addStep(&ViewKitchenDashboardStep{})
	return k
}

// MarksOrderPaid marks an order as paid (if orderID is empty, uses last created order)
func (k *KitchenActor) MarksOrderPaid(orderID string) *KitchenActor {
	k.scenario.addStep(&MarkOrderPaidStep{
		OrderID: orderID,
	})
	return k
}

// MarksOrderPreparing marks an order as preparing (if orderID is empty, uses last created order)
func (k *KitchenActor) MarksOrderPreparing(orderID string) *KitchenActor {
	k.scenario.addStep(&MarkOrderPreparingStep{
		OrderID: orderID,
	})
	return k
}

// MarksOrderReady marks an order as ready (if orderID is empty, uses last created order)
func (k *KitchenActor) MarksOrderReady(orderID string) *KitchenActor {
	k.scenario.addStep(&MarkOrderReadyStep{
		OrderID: orderID,
	})
	return k
}

// SystemActor represents system events
type SystemActor struct {
	scenario *Scenario
}

// PublishesEvent publishes a domain event (for testing event-driven behavior)
func (s *SystemActor) PublishesEvent(event interface{}) *SystemActor {
	s.scenario.addStep(&PublishEventStep{
		Event: event,
	})
	return s
}
