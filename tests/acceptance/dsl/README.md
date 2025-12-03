# Acceptance Test DSL

A fluent Domain-Specific Language (DSL) for writing readable acceptance tests in a Given-When-Then style.

## Overview

This DSL provides a clean, readable way to write acceptance tests that verify end-to-end behavior of the application. Tests are written in a natural language style that makes them easy to understand and maintain.

## Basic Structure

```go
dsl.NewScenario(t, "Test description").
    Given(func(g *dsl.GivenBuilder) {
        // Setup initial state
    }).
    When(func(w *dsl.WhenBuilder) {
        // Perform actions
    }).
    Then(func(t *dsl.ThenBuilder) {
        // Verify outcomes
    }).
    Run()
```

## Given - Setting Up Initial State

Use `Given` to set up the initial state of your test:

```go
Given(func(g *dsl.GivenBuilder) {
    // Restaurant setup
    g.Restaurant("restaurant_1", "Test Cafe", true)
    
    // Menu setup
    g.MenuCategory("cat_1", "restaurant_1", "Mains", 1)
    g.MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true)
    
    // Customer setup
    g.CustomerSession("session-1")
    g.CartWithItems("session-1", "item_1", 1)
    
    // Pre-existing orders
    g.Order("ord_1", "0001", "restaurant_1", "session-1", domain.PaymentStatusPaid)
})
```

### Available Given Methods

- `Restaurant(id, name, isOpen)` - Create a restaurant
- `MenuCategory(id, restaurantID, name, order)` - Create a menu category
- `MenuItem(id, categoryID, restaurantID, name, price, available)` - Create a menu item
- `CustomerSession(sessionID)` - Create a customer session
- `CartWithItems(sessionID, itemID, quantity)` - Add items to cart
- `Order(...)` - Create a pre-existing order
- `OrderWithItems(...)` - Create an order with specific items

## When - Performing Actions

Use `When` to define actions performed by different actors:

```go
When(func(w *dsl.WhenBuilder) {
    // Customer actions
    w.Customer("session-1").
        AddsToCart("item_1", 2).
        CreatesOrder()
    
    // Kitchen actions
    w.Kitchen().
        ViewsDashboard().
        MarksOrderPaid("").
        MarksOrderPreparing("").
        MarksOrderReady("")
    
    // System events
    w.System().PublishesEvent(domain.OrderCreated{...})
})
```

### Customer Actor Methods

- `AddsToCart(itemID, quantity)` - Add item to cart
- `CreatesOrder()` - Create order from cart
- `ViewsOrder(orderNumber)` - View order details

### Kitchen Actor Methods

- `ViewsDashboard()` - View kitchen dashboard
- `MarksOrderPaid(orderID)` - Mark order as paid (empty string = last created order)
- `MarksOrderPreparing(orderID)` - Mark order as preparing
- `MarksOrderReady(orderID)` - Mark order as ready

### System Actor Methods

- `PublishesEvent(event)` - Publish a domain event

## Then - Verifying Outcomes

Use `Then` to assert expected outcomes:

```go
Then(func(t *dsl.ThenBuilder) {
    // Order assertions
    t.OrderShouldBe("0001").
        WithPaymentStatus(domain.PaymentStatusPaid).
        WithFulfillmentStatus(domain.FulfillmentStatusPreparing).
        ContainsHTML("Order #0001")
    
    // Kitchen dashboard assertions
    t.KitchenDashboardShouldShow().
        OrderCount(1).
        ContainsOrder("0001").
        OrderWithStatus("0001", "UNPAID")
    
    // SSE assertions
    t.SSEStreamShouldReceive("/kitchen/stream").
        Event("order-updated").
        WithSelector("#order-0001").
        ContainsHTML("PAID")
})
```

### Order Assertions

- `OrderShouldBe(orderNumber)` - Assert order state (empty string = last created order)
  - `WithPaymentStatus(status)` - Assert payment status
  - `WithFulfillmentStatus(status)` - Assert fulfillment status
  - `ContainsHTML(html)` - Assert HTML content

### Kitchen Dashboard Assertions

- `KitchenDashboardShouldShow()` - Assert dashboard state
  - `OrderCount(count)` - Assert number of orders
  - `ContainsOrder(orderNumber)` - Assert order exists
  - `OrderWithStatus(orderNumber, status)` - Assert order status

### SSE Assertions

- `SSEStreamShouldReceive(stream)` - Assert SSE events
  - `Event(eventType)` - Assert event type
  - `WithSelector(selector)` - Assert CSS selector
  - `ContainsHTML(html)` - Assert HTML content

## Context and Order Tracking

The DSL automatically tracks created orders in a test context. When you create an order, you can reference it later using an empty string:

```go
When(func(w *dsl.WhenBuilder) {
    w.Customer("session-1").CreatesOrder()
}).
Then(func(t *dsl.ThenBuilder) {
    // Empty string uses the last created order
    t.OrderShouldBe("").WithPaymentStatus(domain.PaymentStatusPaid)
}).
When(func(w *dsl.WhenBuilder) {
    // Empty string uses the last created order ID
    w.Kitchen().MarksOrderPaid("")
})
```

## Execution Order

Steps and assertions are executed **in sequence** (interleaved), not batched. This means:

```go
When(...).Then(...).When(...).Then(...)
```

Will execute: Step → Assertion → Step → Assertion

This ensures assertions verify state immediately after actions.

## Examples

See `tests/acceptance/kitchen_workflow_test.go` for complete examples.

## Extending the DSL

To add new steps:

1. Create a step struct in `steps.go` implementing the `Step` interface
2. Add a method to the appropriate actor in `when.go`
3. Call `scenario.addStep()` in the actor method

To add new assertions:

1. Create an assertion struct in `then.go` implementing the `Assertion` interface
2. Add a builder method in `then.go`
3. Implement `Verify()` in `assertions.go`
4. Call `scenario.addAssertion()` in the builder method

