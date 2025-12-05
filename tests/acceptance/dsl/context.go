package dsl

import (
	"time"

	"bitmerchant/internal/domain"
)

// TestContext holds shared state between steps and assertions
type TestContext struct {
	createdOrderID     domain.OrderID
	createdOrderNumber domain.OrderNumber
	restaurantID       domain.RestaurantID
	lastStepDuration   time.Duration
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		restaurantID: domain.RestaurantID("restaurant_1"),
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

// SetLastStepDuration stores the duration of the last executed step
func (ctx *TestContext) SetLastStepDuration(d time.Duration) {
	ctx.lastStepDuration = d
}

// GetLastStepDuration returns the duration of the last executed step
func (ctx *TestContext) GetLastStepDuration() time.Duration {
	return ctx.lastStepDuration
}
