# DSL Framework Features

## ✅ Completed Features

### 1. Rod as Universal Driver

**Files:**
- `rod_helpers.go` - Rod browser automation helpers
- `server.go` - HTTP test server for Rod
- `setup.go` - TestApplication with Rod browser integration

**Features:**
- **Real browser testing**: All tests run in a headless Chrome browser via Rod
- **Natural SSE testing**: SSE events handled by browser's EventSource API
- **DOM monitoring**: Automatic detection of DOM updates via MutationObserver
- **Consistent execution**: Single driver model for all acceptance tests

**Benefits:**
- Tests verify actual user experience (JavaScript, CSS, DOM)
- No complex SSE interception needed
- Better reliability and maintainability
- Clear separation: DSL = "what", Rod = "how"

**Usage:**
```go
// All steps automatically use Rod browser
When(func(w *dsl.WhenBuilder) {
    w.Customer("session-1").ViewsMenu()  // Navigates via browser
    w.Customer("session-1").CreatesOrder()  // HTTP POST via browser
})
```

### 2. SSE Testing with Rod (Real-Time Updates)

**Files:**
- `rod_helpers.go` - DOM monitoring helpers
- `assertions.go` - SSE assertions using DOM verification

**Features:**
- Automatic SSE connection via Datastar's `data-init` attributes
- DOM change detection via MutationObserver
- Verification of DOM updates rather than SSE message interception

**How it works:**
1. `ConnectsToSSE` navigates to the appropriate page
2. Datastar automatically connects to SSE stream via `data-init`
3. DOM observers detect changes made by SSE events
4. Assertions verify DOM was updated correctly

**Usage:**
```go
When(func(w *dsl.WhenBuilder) {
    w.Customer("session-1").CreatesOrder()
    // Navigate to order page - Datastar connects to SSE automatically
    w.ConnectsToSSE("/order/0001/stream").Stream()
}).
Then(func(t *dsl.ThenBuilder) {
    // Verify DOM was updated by SSE event
    t.SSEStreamShouldReceive("/order/0001/stream").
        Event("datastar-patch-elements").
        ContainsHTML("PAID")
})
```

**Note:** The old `SSEClient` approach is deprecated. New tests should use Rod DOM monitoring.

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

### Rod Browser Integration

The framework uses Rod to:
1. Launch a headless Chrome browser for each test
2. Start an HTTP server for the test application
3. Navigate to pages and interact with the DOM
4. Monitor DOM changes via MutationObserver
5. Verify SSE-driven updates by checking DOM state

### SSE Event Flow (Rod-Based)

1. Domain event is published via `EventBus`
2. Event handler processes event and calls `SSEHandler.Broadcast`
3. Browser's EventSource API receives SSE message
4. Datastar processes the SSE event and updates the DOM
5. MutationObserver detects DOM changes
6. Assertions verify DOM was updated correctly

This approach is more reliable because it tests the actual user experience.

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

## Rod Helper Methods

Available in `rod_helpers.go`:

- `NavigateTo(path)` - Navigate to a URL
- `GetPage()` - Get current browser page
- `WaitForDOMUpdate(selector, count, timeout)` - Wait for DOM changes
- `WaitForSSEEvent(timeout)` - Wait for SSE-driven DOM updates
- `SetupDOMObserver(selector)` - Set up MutationObserver
- `SetCookie(name, value)` - Set browser cookie
- `GetCurrentURL()` - Get current page URL
- `ElementExists(selector)` - Check if element exists
- `GetElementText(selector)` - Get element text content
- `GetElementCount(selector)` - Get child element count

## Next Steps (Future Enhancements)

1. **More Rod Helpers**: Add click, fill, screenshot helpers
2. **Watermill CQRS Integration**: Add assertions for Watermill command/query buses
3. **Performance Testing**: Enhanced timing assertions for SSE event delivery
4. **Concurrent Testing**: Support multiple browser tabs/windows
5. **Visual Diff**: HTML diff assertions for better failure messages
6. **Network Monitoring**: Track and verify network requests

