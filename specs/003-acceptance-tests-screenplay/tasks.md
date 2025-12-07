# Acceptance Tests Tasks: Screenplay Pattern with go-rod

## Phase 1: Foundation Setup

### Task 1.1: Add go-rod dependency
- [ ] Run `go get github.com/go-rod/rod`
- [ ] Verify go.mod includes go-rod

### Task 1.2: Create directory structure
```
tests/acceptance/
├── screenplay/
│   ├── actors/
│   ├── abilities/
│   ├── tasks/
│   │   ├── customer/
│   │   ├── kitchen/
│   │   └── owner/
│   ├── interactions/
│   └── questions/
├── scenarios/
├── fixtures/
└── support/
```

### Task 1.3: Implement core screenplay interfaces
- [ ] `screenplay/actors/actor.go` - Base actor struct with abilities management
- [ ] `screenplay/tasks/task.go` - Task interface definition
- [ ] `screenplay/questions/question.go` - Question interface definition

### Task 1.4: Implement BrowseTheWeb ability
- [ ] `screenplay/abilities/browse_web.go` - go-rod wrapper
- [ ] Methods: OpenPage, NavigateTo, Page, BaseURL, Close

### Task 1.5: Implement base interactions
- [ ] `screenplay/interactions/click.go`
- [ ] `screenplay/interactions/fill.go`
- [ ] `screenplay/interactions/navigate.go`
- [ ] `screenplay/interactions/wait_for.go`
- [ ] `screenplay/interactions/datastar.go` - Datastar-specific helpers

### Task 1.6: Set up test server infrastructure
- [ ] `support/server.go` - Test server management (httptest.Server)
- [ ] `support/browser.go` - Browser factory with headless config
- [ ] `support/helpers.go` - Common test utilities

---

## Phase 2: Customer Journey (US1)

### Task 2.1: Implement Customer actor
- [ ] `screenplay/actors/customer.go` - Customer factory function
- [ ] Default abilities for browsing

### Task 2.2: Implement customer tasks
- [ ] `screenplay/tasks/customer/navigate_to_menu.go`
- [ ] `screenplay/tasks/customer/add_to_cart.go`
- [ ] `screenplay/tasks/customer/remove_from_cart.go`
- [ ] `screenplay/tasks/customer/view_cart.go`
- [ ] `screenplay/tasks/customer/proceed_to_checkout.go`
- [ ] `screenplay/tasks/customer/confirm_cash_payment.go`
- [ ] `screenplay/tasks/customer/view_order_status.go`
- [ ] `screenplay/tasks/customer/lookup_order.go`

### Task 2.3: Implement customer questions
- [ ] `screenplay/questions/cart_total.go`
- [ ] `screenplay/questions/cart_items.go`
- [ ] `screenplay/questions/cart_item_count.go`
- [ ] `screenplay/questions/order_number.go`
- [ ] `screenplay/questions/order_status.go`
- [ ] `screenplay/questions/menu_categories.go`
- [ ] `screenplay/questions/menu_items.go`

### Task 2.4: Add data-testid attributes to menu templates
- [ ] Menu item cards: `data-testid="menu-item-{id}"`
- [ ] Add to cart buttons: `data-testid="add-to-cart-{id}"`
- [ ] Cart summary: `data-testid="cart-summary"`
- [ ] Cart total: `data-testid="cart-total"`
- [ ] Cart item count: `data-testid="cart-item-count"`

### Task 2.5: Add data-testid attributes to order templates
- [ ] Order number: `data-testid="order-number"`
- [ ] Order status: `data-testid="order-status"`
- [ ] Checkout button: `data-testid="checkout-btn"`
- [ ] Confirm payment button: `data-testid="confirm-payment-btn"`

### Task 2.6: Write US1 acceptance tests
- [ ] `scenarios/customer_ordering_test.go`
  - [ ] Test: Customer can browse menu
  - [ ] Test: Customer can add items to cart
  - [ ] Test: Customer can remove items from cart
  - [ ] Test: Customer can view cart with correct total
  - [ ] Test: Customer can place order with cash payment
  - [ ] Test: Customer receives order confirmation with order number
  - [ ] Test: Customer sees order status as UNPAID initially
  - [ ] Test: Customer sees real-time status updates (SSE)

---

## Phase 3: Kitchen Operations (US2)

### Task 3.1: Implement Kitchen Staff actor
- [ ] `screenplay/actors/kitchen_staff.go` - Kitchen staff factory function

### Task 3.2: Implement kitchen tasks
- [ ] `screenplay/tasks/kitchen/navigate_to_display.go`
- [ ] `screenplay/tasks/kitchen/mark_order_paid.go`
- [ ] `screenplay/tasks/kitchen/mark_order_preparing.go`
- [ ] `screenplay/tasks/kitchen/mark_order_ready.go`

### Task 3.3: Implement kitchen questions
- [ ] `screenplay/questions/kitchen_orders.go` - List of orders on display
- [ ] `screenplay/questions/kitchen_order_status.go` - Status for specific order

### Task 3.4: Add data-testid attributes to kitchen templates
- [ ] Order cards: `data-testid="kitchen-order-{number}"`
- [ ] Order status badge: `data-testid="order-status-{number}"`
- [ ] Mark paid button: `data-testid="mark-paid-{id}"`
- [ ] Mark preparing button: `data-testid="mark-preparing-{id}"`
- [ ] Mark ready button: `data-testid="mark-ready-{id}"`

### Task 3.5: Write US2 acceptance tests
- [ ] `scenarios/kitchen_fulfillment_test.go`
  - [ ] Test: Kitchen staff sees new orders
  - [ ] Test: Kitchen staff can mark order as paid
  - [ ] Test: Kitchen staff can mark order as preparing
  - [ ] Test: Kitchen staff can mark order as ready
  - [ ] Test: Status updates propagate to customer view (SSE)

---

## Phase 4: Owner Management (US3 & US4)

### Task 4.1: Implement Restaurant Owner actor
- [ ] `screenplay/actors/owner.go` - Owner factory function

### Task 4.2: Implement owner tasks
- [ ] `screenplay/tasks/owner/navigate_to_dashboard.go`
- [ ] `screenplay/tasks/owner/create_category.go`
- [ ] `screenplay/tasks/owner/create_menu_item.go`
- [ ] `screenplay/tasks/owner/upload_photo.go`
- [ ] `screenplay/tasks/owner/toggle_restaurant_status.go`
- [ ] `screenplay/tasks/owner/view_analytics.go`

### Task 4.3: Implement owner questions
- [ ] `screenplay/questions/dashboard_stats.go` - Today's orders, sales, avg order
- [ ] `screenplay/questions/order_history.go`
- [ ] `screenplay/questions/top_selling_items.go`
- [ ] `screenplay/questions/restaurant_status.go` - Open/Closed

### Task 4.4: Add data-testid attributes to admin/dashboard templates
- [ ] Category form: `data-testid="category-form"`
- [ ] Item form: `data-testid="item-form"`
- [ ] Stats cards: `data-testid="stat-orders-today"`, `data-testid="stat-total-sales"`
- [ ] Toggle status button: `data-testid="toggle-status-btn"`

### Task 4.5: Write US3 acceptance tests
- [ ] `scenarios/menu_management_test.go`
  - [ ] Test: Owner can create menu category
  - [ ] Test: Owner can create menu item
  - [ ] Test: Menu changes appear in customer view
  - [ ] Test: Owner can toggle restaurant open/closed

### Task 4.6: Write US4 acceptance tests
- [ ] `scenarios/dashboard_analytics_test.go`
  - [ ] Test: Owner sees orders today count
  - [ ] Test: Owner sees total sales
  - [ ] Test: Owner sees top selling items
  - [ ] Test: Owner can view order history

---

## Phase 5: CI/CD & Polish

### Task 5.1: Configure CI pipeline
- [ ] Add acceptance test job to CI workflow
- [ ] Configure headless Chrome for CI
- [ ] Set appropriate timeouts

### Task 5.2: Add Taskfile commands
```yaml
test:acceptance:
  desc: Run acceptance tests
  cmds:
    - go test -v ./tests/acceptance/... -tags=acceptance

test:acceptance:debug:
  desc: Run acceptance tests with visible browser
  env:
    HEADLESS: "false"
  cmds:
    - go test -v ./tests/acceptance/... -tags=acceptance
```

### Task 5.3: Documentation
- [ ] Add README to tests/acceptance explaining structure
- [ ] Document how to run tests locally
- [ ] Document how to debug failing tests

### Task 5.4: Test utilities
- [ ] Screenshot on failure helper
- [ ] Test data seeding utilities
- [ ] Cleanup utilities

---

## Estimated Timeline

| Phase | Tasks | Est. Time |
|-------|-------|-----------|
| Phase 1 | Foundation | 1-2 days |
| Phase 2 | Customer Journey | 2-3 days |
| Phase 3 | Kitchen Operations | 1-2 days |
| Phase 4 | Owner Management | 1-2 days |
| Phase 5 | CI/CD & Polish | 1 day |
| **Total** | | **6-10 days** |

---

## Definition of Done

- [ ] All acceptance tests pass in CI (headless)
- [ ] Tests can be run locally in debug mode (visible browser)
- [ ] All 4 user stories have acceptance coverage
- [ ] Tests run in under 2 minutes total
- [ ] Code reviewed and merged
