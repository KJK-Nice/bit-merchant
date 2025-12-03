package dsl

import (
	"bitmerchant/internal/domain"
)

// TestContext holds shared state between steps and assertions
type TestContext struct {
	createdOrderID     domain.OrderID
	createdOrderNumber domain.OrderNumber
	restaurantID       domain.RestaurantID
	sseClients         map[string]*SSEClient // path -> client
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		restaurantID: domain.RestaurantID("restaurant_1"),
		sseClients:   make(map[string]*SSEClient),
	}
}

// SetCreatedOrder stores the created order information
func (ctx *TestContext) SetCreatedOrder(orderID domain.OrderID, orderNumber domain.OrderNumber) {
	ctx.createdOrderID = orderID
	ctx.createdOrderNumber = orderNumber
}

// GetCreatedOrderID returns the created order ID
func (ctx *TestContext) GetCreatedOrderID() domain.OrderID {
	return ctx.createdOrderID
}

// GetCreatedOrderNumber returns the created order number
func (ctx *TestContext) GetCreatedOrderNumber() domain.OrderNumber {
	return ctx.createdOrderNumber
}

// SetSSEClient stores an SSE client
func (ctx *TestContext) SetSSEClient(path string, client *SSEClient) {
	ctx.sseClients[path] = client
}

// GetSSEClient retrieves an SSE client
func (ctx *TestContext) GetSSEClient(path string) *SSEClient {
	return ctx.sseClients[path]
}
