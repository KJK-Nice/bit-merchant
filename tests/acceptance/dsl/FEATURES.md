# DSL Framework Features

## ✅ Completed Features

### 1. SSE Client Helper for Real-Time Testing

**Files:**
- `sse_client.go` - SSE client implementation
- `test_sse_handler.go` - Test wrapper that captures broadcasts

**Features:**
- `SSEClient` - Captures SSE events broadcast by the application
- `ParseSSEMessage` - Parses SSE messages including Datastar format
- `WaitForEvent` - Waits for specific event types
- `WaitForDatastarPatch` - Waits for Datastar patch events with selectors
- `GetAllEvents` - Returns all captured events
- `GetEventsByType` - Filters events by type

**Usage:**
```go
When(func(w *dsl.WhenBuilder) {
    w.ConnectsToSSE("/kitchen/stream").Stream()
}).
Then(func(t *dsl.ThenBuilder) {
    t.SSEStreamShouldReceive("/kitchen/stream").
        Event("datastar-patch-elements").
        WithSelector("#orders-list").
        ContainsHTML("Order #0001")
})
```

### 2. Enhanced Assertion Types

**New Assertions:**
- `MenuAssertion` - Assert menu state
- `OrderHistoryAssertion` - Assert order history
- `CQRSAssertion` - Assert CQRS-specific behavior
- Enhanced `SSEAssertion` - Full SSE event verification

**Usage:**
```go
Then(func(t *dsl.ThenBuilder) {
    // Menu assertions
    t.MenuShouldShow().
        ItemCount(5).
        ContainsItem("Burger")
    
    // Order history assertions
    t.OrderHistoryShouldShow().
        OrderCount(3).
        ContainsOrder("0001")
    
    // CQRS assertions
    t.CQRSShould().
        HavePublishedEvent("OrderCreated")
    
    // SSE assertions (enhanced)
    t.SSEStreamShouldReceive("/kitchen/stream").
        Event("datastar-patch-elements").
        WithSelector("#orders-list")
})
```

### 3. Extended Step Types

**New Steps:**
- `ViewMenuStep` - View the menu
- `AddMultipleItemsStep` - Add multiple items to cart at once
- `ViewOrderHistoryStep` - View customer order history
- `ViewDashboardStep` - View admin dashboard
- `SSEStep` - Connect to SSE stream

**Usage:**
```go
When(func(w *dsl.WhenBuilder) {
    // Customer actions
    w.Customer("session-1").
        ViewsMenu().
        AddsMultipleItems(map[string]int{
            "item_1": 2,
            "item_2": 1,
        }).
        ViewsOrderHistory()
    
    // SSE connection
    w.ConnectsToSSE("/kitchen/stream").Stream()
    
    // Dashboard view
    w.System().ViewsDashboard()
})
```

### 4. CQRS-Specific Assertions

**Features:**
- Verify command execution
- Verify query execution
- Verify event publication
- Track SSE broadcasts as indirect event verification

**Usage:**
```go
Then(func(t *dsl.ThenBuilder) {
    t.CQRSShould().
        HaveExecutedCommand().
        HavePublishedEvent("OrderCreated")
})
```

## Implementation Details

### SSE Event Capture

The framework uses a `TestSSEHandler` wrapper that:
1. Embeds the real `SSEHandler` for compatibility
2. Captures all `Broadcast` calls
3. Stores broadcasts by topic
4. Allows SSE clients to retrieve captured events

### Event Flow

1. Domain event is published via `EventBus`
2. Event handler processes event and calls `SSEHandler.Broadcast`
3. `TestSSEHandler` captures the broadcast
4. SSE clients retrieve captured events
5. Assertions verify events were received

### Context Management

The `TestContext` now tracks:
- Created orders (ID and number)
- SSE clients by path
- Shared state between steps and assertions

## Example: Complete Test with All Features

```go
func TestCompleteWorkflow(t *testing.T) {
    dsl.NewScenario(t, "Complete order workflow with SSE").
        Given(func(g *dsl.GivenBuilder) {
            g.Restaurant("restaurant_1", "Test Cafe", true).
                MenuCategory("cat_1", "restaurant_1", "Mains", 1).
                MenuItem("item_1", "cat_1", "restaurant_1", "Burger", 10.00, true).
                MenuItem("item_2", "cat_1", "restaurant_1", "Pizza", 15.00, true).
                CustomerSession("session-1")
        }).
        When(func(w *dsl.WhenBuilder) {
            // Connect to SSE streams
            w.ConnectsToSSE("/kitchen/stream").Stream()
            
            // Customer actions
            w.Customer("session-1").
                ViewsMenu().
                AddsMultipleItems(map[string]int{
                    "item_1": 2,
                    "item_2": 1,
                }).
                CreatesOrder()
        }).
        Then(func(t *dsl.ThenBuilder) {
            // Menu assertions
            t.MenuShouldShow().
                ItemCount(2).
                ContainsItem("Burger")
            
            // Kitchen dashboard assertions
            t.KitchenDashboardShouldShow().
                OrderCount(1)
            
            // SSE assertions
            t.SSEStreamShouldReceive("/kitchen/stream").
                Event("datastar-patch-elements").
                WithSelector("#orders-list").
                ContainsHTML("Burger")
            
            // CQRS assertions
            t.CQRSShould().
                HavePublishedEvent("OrderCreated")
        }).
        When(func(w *dsl.WhenBuilder) {
            w.Kitchen().MarksOrderPaid("")
        }).
        Then(func(t *dsl.ThenBuilder) {
            t.OrderShouldBe("").
                WithPaymentStatus(domain.PaymentStatusPaid)
            
            // Verify SSE event for order status update
            t.SSEStreamShouldReceive("/order/0001/stream").
                Event("datastar-patch-elements").
                ContainsHTML("PAID")
        }).
        Run()
}
```

## Next Steps (Future Enhancements)

1. **Watermill CQRS Integration**: Add assertions for Watermill command/query buses
2. **Performance Testing**: Add timing assertions for SSE event delivery
3. **Concurrent Testing**: Support multiple SSE clients simultaneously
4. **Event Replay**: Replay captured events for testing event handlers
5. **Visual Diff**: HTML diff assertions for better failure messages

