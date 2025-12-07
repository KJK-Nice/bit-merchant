# Acceptance Tests Plan: Screenplay Pattern with go-rod

**Feature Branch**: `003-acceptance-tests-screenplay`  
**Created**: 2025-12-07  
**Status**: Draft

## Overview

This plan outlines the implementation of acceptance tests for BitMerchant using the **Screenplay Pattern** with [go-rod](https://github.com/go-rod/rod) as the browser automation driver.

### Why Screenplay Pattern?

The Screenplay pattern provides several benefits over traditional Page Object Models:

1. **User-Centric**: Tests read like user stories ("Customer adds item to cart")
2. **Composable**: Tasks can be composed from smaller interactions
3. **Maintainable**: Changes in UI only affect low-level interactions
4. **Reusable**: Actors, abilities, and tasks are highly reusable
5. **Readable**: Business stakeholders can understand test scenarios

### Why go-rod?

- **Native Go**: Perfect integration with Go test framework
- **Chromium DevTools Protocol**: Direct browser control without WebDriver
- **Headless by default**: Fast CI/CD execution
- **Auto-wait**: Built-in smart waiting for elements
- **Simple API**: Easy to learn and use

---

## Screenplay Pattern Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Test Scenario                            │
│  "Given customer has items in cart, when they place order..."   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                          ACTORS                                  │
│  Customer, KitchenStaff, RestaurantOwner                        │
│  (Who is performing the action?)                                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         ABILITIES                                │
│  BrowseTheWeb (wraps go-rod Page)                               │
│  (What can the actor do?)                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                          TASKS                                   │
│  AddItemToCart, PlaceOrder, MarkOrderPaid, CreateMenuItem       │
│  (High-level business actions composed of interactions)         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      INTERACTIONS                                │
│  Click, Fill, Navigate, WaitFor, Select                         │
│  (Low-level browser actions)                                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        QUESTIONS                                 │
│  CartTotal, OrderStatus, MenuItemCount                          │
│  (Query system state for assertions)                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
tests/
├── acceptance/
│   ├── screenplay/
│   │   ├── actors/
│   │   │   ├── actor.go           # Base actor with abilities
│   │   │   ├── customer.go        # Customer actor factory
│   │   │   ├── kitchen_staff.go   # Kitchen staff actor factory
│   │   │   └── owner.go           # Restaurant owner actor factory
│   │   │
│   │   ├── abilities/
│   │   │   └── browse_web.go      # go-rod browser ability
│   │   │
│   │   ├── tasks/
│   │   │   ├── customer/
│   │   │   │   ├── browse_menu.go
│   │   │   │   ├── add_to_cart.go
│   │   │   │   ├── place_order.go
│   │   │   │   └── view_order_status.go
│   │   │   ├── kitchen/
│   │   │   │   ├── view_orders.go
│   │   │   │   ├── mark_paid.go
│   │   │   │   ├── mark_preparing.go
│   │   │   │   └── mark_ready.go
│   │   │   └── owner/
│   │   │       ├── create_category.go
│   │   │       ├── create_item.go
│   │   │       └── view_dashboard.go
│   │   │
│   │   ├── interactions/
│   │   │   ├── click.go
│   │   │   ├── fill.go
│   │   │   ├── navigate.go
│   │   │   ├── wait_for.go
│   │   │   └── datastar.go        # Datastar-specific interactions
│   │   │
│   │   └── questions/
│   │       ├── cart_total.go
│   │       ├── cart_items.go
│   │       ├── order_status.go
│   │       ├── menu_items.go
│   │       └── kitchen_orders.go
│   │
│   ├── scenarios/
│   │   ├── customer_ordering_test.go      # US1: Customer orders food
│   │   ├── kitchen_fulfillment_test.go    # US2: Kitchen staff fulfills orders
│   │   ├── menu_management_test.go        # US3: Owner sets up menu
│   │   └── dashboard_analytics_test.go    # US4: Owner views dashboard
│   │
│   ├── fixtures/
│   │   └── test_data.go           # Test data setup
│   │
│   └── support/
│       ├── server.go              # Test server management
│       ├── browser.go             # go-rod browser setup
│       └── helpers.go             # Common test utilities
```

---

## Core Components Implementation

### 1. Actor (Base)

```go
// tests/acceptance/screenplay/actors/actor.go
package actors

import "bitmerchant/tests/acceptance/screenplay/abilities"

type Actor struct {
    name      string
    abilities map[string]interface{}
}

func NewActor(name string) *Actor {
    return &Actor{
        name:      name,
        abilities: make(map[string]interface{}),
    }
}

func (a *Actor) Can(ability interface{}) *Actor {
    switch ab := ability.(type) {
    case *abilities.BrowseTheWeb:
        a.abilities["BrowseTheWeb"] = ab
    }
    return a
}

func (a *Actor) BrowseTheWeb() *abilities.BrowseTheWeb {
    return a.abilities["BrowseTheWeb"].(*abilities.BrowseTheWeb)
}

func (a *Actor) AttemptsTo(tasks ...Task) error {
    for _, task := range tasks {
        if err := task.PerformAs(a); err != nil {
            return fmt.Errorf("%s failed to %s: %w", a.name, task.Name(), err)
        }
    }
    return nil
}

func (a *Actor) AsksFor(question Question) (interface{}, error) {
    return question.AnsweredBy(a)
}
```

### 2. Browse The Web Ability (go-rod wrapper)

```go
// tests/acceptance/screenplay/abilities/browse_web.go
package abilities

import (
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

type BrowseTheWeb struct {
    browser *rod.Browser
    page    *rod.Page
    baseURL string
}

func BrowseTheWebUsing(browser *rod.Browser, baseURL string) *BrowseTheWeb {
    return &BrowseTheWeb{
        browser: browser,
        baseURL: baseURL,
    }
}

func (b *BrowseTheWeb) OpenPage() *rod.Page {
    if b.page == nil {
        b.page = b.browser.MustPage()
    }
    return b.page
}

func (b *BrowseTheWeb) NavigateTo(path string) error {
    return b.OpenPage().Navigate(b.baseURL + path)
}

func (b *BrowseTheWeb) Page() *rod.Page {
    return b.page
}

func (b *BrowseTheWeb) BaseURL() string {
    return b.baseURL
}

func (b *BrowseTheWeb) Close() {
    if b.page != nil {
        b.page.Close()
    }
}
```

### 3. Task Interface

```go
// tests/acceptance/screenplay/tasks/task.go
package tasks

import "bitmerchant/tests/acceptance/screenplay/actors"

type Task interface {
    Name() string
    PerformAs(actor *actors.Actor) error
}
```

### 4. Question Interface

```go
// tests/acceptance/screenplay/questions/question.go
package questions

import "bitmerchant/tests/acceptance/screenplay/actors"

type Question interface {
    AnsweredBy(actor *actors.Actor) (interface{}, error)
}
```

### 5. Example Task: Add Item to Cart

```go
// tests/acceptance/screenplay/tasks/customer/add_to_cart.go
package customer

import (
    "fmt"
    "bitmerchant/tests/acceptance/screenplay/actors"
)

type AddItemToCart struct {
    itemName string
    quantity int
}

func AddToCart(itemName string) *AddItemToCart {
    return &AddItemToCart{itemName: itemName, quantity: 1}
}

func (t *AddItemToCart) WithQuantity(qty int) *AddItemToCart {
    t.quantity = qty
    return t
}

func (t *AddItemToCart) Name() string {
    return fmt.Sprintf("add %d x %s to cart", t.quantity, t.itemName)
}

func (t *AddItemToCart) PerformAs(actor *actors.Actor) error {
    page := actor.BrowseTheWeb().Page()
    
    // Find the menu item card by name
    itemCard := page.MustElement(fmt.Sprintf(`[data-testid="menu-item-%s"]`, t.itemName))
    
    // Click "Add to Cart" button
    addButton := itemCard.MustElement(`button:has-text("Add to Cart")`)
    
    for i := 0; i < t.quantity; i++ {
        addButton.MustClick()
        // Wait for Datastar to update the cart
        page.MustWaitIdle()
    }
    
    return nil
}
```

### 6. Example Question: Cart Total

```go
// tests/acceptance/screenplay/questions/cart_total.go
package questions

import (
    "strconv"
    "strings"
    "bitmerchant/tests/acceptance/screenplay/actors"
)

type CartTotalQuestion struct{}

func TheCartTotal() *CartTotalQuestion {
    return &CartTotalQuestion{}
}

func (q *CartTotalQuestion) AnsweredBy(actor *actors.Actor) (interface{}, error) {
    page := actor.BrowseTheWeb().Page()
    
    totalText := page.MustElement(`[data-testid="cart-total"]`).MustText()
    // Parse "$10.00" to 10.00
    totalStr := strings.TrimPrefix(totalText, "$")
    return strconv.ParseFloat(totalStr, 64)
}
```

---

## Test Scenarios

### Scenario 1: Customer Orders Food with Cash Payment (US1)

```go
// tests/acceptance/scenarios/customer_ordering_test.go
package scenarios

import (
    "testing"
    "bitmerchant/tests/acceptance/screenplay/actors"
    "bitmerchant/tests/acceptance/screenplay/tasks/customer"
    "bitmerchant/tests/acceptance/screenplay/questions"
    "github.com/stretchr/testify/assert"
)

func TestCustomerOrdersFood(t *testing.T) {
    // Setup
    server := support.StartTestServer(t)
    defer server.Stop()
    
    browser := support.NewBrowser(t)
    defer browser.Close()
    
    // Create actor
    sarah := actors.NewCustomer("Sarah").
        WithAbility(abilities.BrowseTheWebUsing(browser, server.URL()))
    
    t.Run("customer can browse menu and add items to cart", func(t *testing.T) {
        // Given Sarah is viewing the menu
        err := sarah.AttemptsTo(
            customer.NavigateToMenu(),
        )
        assert.NoError(t, err)
        
        // When she adds items to cart
        err = sarah.AttemptsTo(
            customer.AddToCart("Bitcoin Burger"),
            customer.AddToCart("Satoshi Soda"),
        )
        assert.NoError(t, err)
        
        // Then the cart total should be correct
        total, _ := sarah.AsksFor(questions.TheCartTotal())
        assert.Equal(t, 18.00, total)
        
        itemCount, _ := sarah.AsksFor(questions.TheCartItemCount())
        assert.Equal(t, 2, itemCount)
    })
    
    t.Run("customer can place order with cash payment", func(t *testing.T) {
        // Given Sarah has items in cart
        err := sarah.AttemptsTo(
            customer.NavigateToMenu(),
            customer.AddToCart("Bitcoin Burger"),
        )
        assert.NoError(t, err)
        
        // When she places the order
        err = sarah.AttemptsTo(
            customer.ViewCart(),
            customer.ProceedToCheckout(),
            customer.ConfirmCashPayment(),
        )
        assert.NoError(t, err)
        
        // Then she should see order confirmation
        orderNumber, _ := sarah.AsksFor(questions.TheOrderNumber())
        assert.NotEmpty(t, orderNumber)
        
        orderStatus, _ := sarah.AsksFor(questions.TheOrderStatus())
        assert.Equal(t, "UNPAID", orderStatus)
    })
    
    t.Run("customer sees real-time order status updates", func(t *testing.T) {
        // Given Sarah has placed an order
        // ... (place order first)
        
        // When kitchen marks order as paid
        // (This requires a separate kitchen actor)
        marcus := actors.NewKitchenStaff("Marcus").
            WithAbility(abilities.BrowseTheWebUsing(browser, server.URL()))
        
        err := marcus.AttemptsTo(
            kitchen.NavigateToKitchenDisplay(),
            kitchen.MarkOrderAsPaid(orderNumber),
        )
        assert.NoError(t, err)
        
        // Then Sarah should see the status update (via SSE)
        // Wait for SSE update
        time.Sleep(2 * time.Second)
        
        orderStatus, _ := sarah.AsksFor(questions.TheOrderStatus())
        assert.Equal(t, "PAID", orderStatus)
    })
}
```

### Scenario 2: Kitchen Staff Fulfills Orders (US2)

```go
// tests/acceptance/scenarios/kitchen_fulfillment_test.go
package scenarios

func TestKitchenFulfillsOrders(t *testing.T) {
    server := support.StartTestServer(t)
    defer server.Stop()
    
    browser := support.NewBrowser(t)
    defer browser.Close()
    
    // Setup: Customer places order
    customer := actors.NewCustomer("Sarah")
    customer.Can(abilities.BrowseTheWebUsing(browser.MustPage(), server.URL()))
    
    orderNumber := placeTestOrder(t, customer, "Bitcoin Burger")
    
    // Kitchen staff actor
    marcus := actors.NewKitchenStaff("Marcus")
    marcus.Can(abilities.BrowseTheWebUsing(browser.MustPage(), server.URL()))
    
    t.Run("kitchen staff sees new orders", func(t *testing.T) {
        err := marcus.AttemptsTo(
            kitchen.NavigateToKitchenDisplay(),
        )
        assert.NoError(t, err)
        
        orders, _ := marcus.AsksFor(questions.TheKitchenOrders())
        assert.Contains(t, orders, orderNumber)
    })
    
    t.Run("kitchen staff marks order as paid", func(t *testing.T) {
        err := marcus.AttemptsTo(
            kitchen.MarkOrderAsPaid(orderNumber),
        )
        assert.NoError(t, err)
        
        status, _ := marcus.AsksFor(questions.TheOrderStatusFor(orderNumber))
        assert.Equal(t, "PAID", status)
    })
    
    t.Run("kitchen staff marks order as preparing", func(t *testing.T) {
        err := marcus.AttemptsTo(
            kitchen.MarkOrderAsPreparing(orderNumber),
        )
        assert.NoError(t, err)
        
        status, _ := marcus.AsksFor(questions.TheOrderStatusFor(orderNumber))
        assert.Equal(t, "PREPARING", status)
    })
    
    t.Run("kitchen staff marks order as ready", func(t *testing.T) {
        err := marcus.AttemptsTo(
            kitchen.MarkOrderAsReady(orderNumber),
        )
        assert.NoError(t, err)
        
        status, _ := marcus.AsksFor(questions.TheOrderStatusFor(orderNumber))
        assert.Equal(t, "READY", status)
    })
}
```

### Scenario 3: Owner Sets Up Menu (US3)

```go
// tests/acceptance/scenarios/menu_management_test.go
package scenarios

func TestOwnerSetsUpMenu(t *testing.T) {
    server := support.StartTestServer(t)
    defer server.Stop()
    
    browser := support.NewBrowser(t)
    defer browser.Close()
    
    linda := actors.NewRestaurantOwner("Linda")
    linda.Can(abilities.BrowseTheWebUsing(browser.MustPage(), server.URL()))
    
    t.Run("owner creates menu category", func(t *testing.T) {
        err := linda.AttemptsTo(
            owner.NavigateToAdminDashboard(),
            owner.CreateCategory("Desserts", 4),
        )
        assert.NoError(t, err)
        
        categories, _ := linda.AsksFor(questions.TheMenuCategories())
        assert.Contains(t, categories, "Desserts")
    })
    
    t.Run("owner creates menu item", func(t *testing.T) {
        err := linda.AttemptsTo(
            owner.CreateMenuItem(owner.MenuItem{
                Name:        "Chocolate Cake",
                Description: "Rich chocolate layer cake",
                Price:       8.50,
                Category:    "Desserts",
            }),
        )
        assert.NoError(t, err)
        
        items, _ := linda.AsksFor(questions.TheMenuItemsInCategory("Desserts"))
        assert.Contains(t, items, "Chocolate Cake")
    })
}
```

---

## Test Support Infrastructure

### Test Server Management

```go
// tests/acceptance/support/server.go
package support

import (
    "net/http/httptest"
    "testing"
)

type TestServer struct {
    server *httptest.Server
    // Add repositories, event bus, etc. for test setup
}

func StartTestServer(t *testing.T) *TestServer {
    t.Helper()
    
    // Initialize all dependencies (similar to main.go but for tests)
    // Use in-memory repositories
    // Start Echo server on random port
    
    e := setupTestEchoServer()
    server := httptest.NewServer(e)
    
    return &TestServer{server: server}
}

func (s *TestServer) URL() string {
    return s.server.URL
}

func (s *TestServer) Stop() {
    s.server.Close()
}
```

### Browser Setup (go-rod)

```go
// tests/acceptance/support/browser.go
package support

import (
    "testing"
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

func NewBrowser(t *testing.T) *rod.Browser {
    t.Helper()
    
    // Use launcher for more control
    url := launcher.New().
        Headless(true).  // Set to false for debugging
        MustLaunch()
    
    browser := rod.New().
        ControlURL(url).
        MustConnect()
    
    t.Cleanup(func() {
        browser.MustClose()
    })
    
    return browser
}

func NewBrowserWithDevTools(t *testing.T) *rod.Browser {
    t.Helper()
    
    // For debugging - opens browser with DevTools
    url := launcher.New().
        Headless(false).
        Devtools(true).
        MustLaunch()
    
    return rod.New().
        ControlURL(url).
        MustConnect()
}
```

---

## Datastar-Specific Considerations

Since BitMerchant uses Datastar for hypermedia interactions, we need special handling:

### Datastar Interactions

```go
// tests/acceptance/screenplay/interactions/datastar.go
package interactions

import (
    "time"
    "github.com/go-rod/rod"
)

// WaitForDatastarUpdate waits for Datastar to complete its update cycle
func WaitForDatastarUpdate(page *rod.Page) error {
    // Datastar uses data-* attributes for reactivity
    // Wait for any pending network requests
    page.MustWaitIdle()
    
    // Additional wait for SSE updates if needed
    time.Sleep(100 * time.Millisecond)
    
    return nil
}

// ClickDatastarAction clicks an element with data-on:click attribute
func ClickDatastarAction(page *rod.Page, selector string) error {
    el := page.MustElement(selector)
    el.MustClick()
    
    // Wait for Datastar to process the action
    return WaitForDatastarUpdate(page)
}

// WaitForSSEUpdate waits for Server-Sent Events to update the DOM
func WaitForSSEUpdate(page *rod.Page, selector string, expectedContent string, timeout time.Duration) error {
    return page.Timeout(timeout).MustElement(selector).MustContains(expectedContent)
}
```

---

## Implementation Phases

### Phase 1: Foundation (1-2 days)
- [ ] Set up go-rod dependency
- [ ] Create base actor, ability, task, question interfaces
- [ ] Implement BrowseTheWeb ability
- [ ] Set up test server infrastructure
- [ ] Add basic interactions (click, fill, navigate)

### Phase 2: Customer Journey (2-3 days)
- [ ] Implement customer actor and tasks
- [ ] Create menu browsing tasks
- [ ] Create cart management tasks
- [ ] Create order placement tasks
- [ ] Add cart and order questions
- [ ] Write US1 acceptance tests

### Phase 3: Kitchen Operations (1-2 days)
- [ ] Implement kitchen staff actor and tasks
- [ ] Create order management tasks (mark paid/preparing/ready)
- [ ] Add kitchen order questions
- [ ] Write US2 acceptance tests

### Phase 4: Owner Management (1-2 days)
- [ ] Implement owner actor and tasks
- [ ] Create menu management tasks
- [ ] Add dashboard questions
- [ ] Write US3 and US4 acceptance tests

### Phase 5: Polish & CI (1 day)
- [ ] Add data-testid attributes to templates
- [ ] Configure CI pipeline for headless testing
- [ ] Add test utilities and helpers
- [ ] Documentation

---

## Required Changes to Codebase

### 1. Add test IDs to templates

To make tests reliable, add `data-testid` attributes to key elements:

```html
<!-- Menu item -->
<div data-testid="menu-item-{{ .ID }}">

<!-- Cart total -->
<span data-testid="cart-total">$10.00</span>

<!-- Order status -->
<span data-testid="order-status">UNPAID</span>

<!-- Kitchen order card -->
<div data-testid="kitchen-order-{{ .OrderNumber }}">
```

### 2. Add go-rod dependency

```bash
go get github.com/go-rod/rod
```

### 3. Taskfile.yml additions

```yaml
tasks:
  test:acceptance:
    desc: Run acceptance tests with browser
    cmds:
      - go test -v ./tests/acceptance/... -tags=acceptance

  test:acceptance:debug:
    desc: Run acceptance tests with visible browser
    env:
      HEADLESS: "false"
    cmds:
      - go test -v ./tests/acceptance/... -tags=acceptance
```

---

## Success Criteria

- [ ] All 4 user stories have acceptance test coverage
- [ ] Tests run in CI (headless) in under 2 minutes
- [ ] Tests are readable by non-technical stakeholders
- [ ] Test failures provide clear diagnostic information
- [ ] SSE real-time updates are properly tested

---

## References

- [go-rod Documentation](https://go-rod.github.io/)
- [Screenplay Pattern](https://serenity-js.org/handbook/design/screenplay-pattern.html)
- [BitMerchant Spec](../002-cash-payment-hypermedia/spec.md)
