# Tasks: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

**Input**: Design documents from `/specs/001-lightning-restaurant-platform/`  
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Testing Standards), all features MUST include comprehensive tests. Test tasks are MANDATORY and MUST be written before implementation (TDD). Minimum 80% coverage required (95% for critical paths: payment processing, order management, kitchen display).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Single project with Go:
- `internal/` - Application code (domain, application, infrastructure, interfaces)
- `cmd/server/` - Entry point
- `tests/` - All tests (unit, integration, contract)
- `static/` - Static assets (CSS, JS, PWA)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project directory structure per implementation plan
- [x] T002 Initialize Go module with `go mod init bitmerchant`
- [x] T003 [P] Install core dependencies: Echo, Templ, Watermill, Datastar, testify
- [x] T004 [P] Configure golangci-lint with strict rules in `.golangci.yml`
- [x] T005 [P] Configure gocyclo for complexity checks (max 10)
- [x] T006 [P] Create `.env.example` with required environment variables
- [x] T007 [P] Create PWA manifest in `static/pwa/manifest.json`
- [x] T008 [P] Create service worker for offline support in `static/pwa/sw.js`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T009 Create domain entities: Restaurant in `internal/domain/restaurant.go`
- [x] T010 [P] Create domain entities: MenuCategory in `internal/domain/menu.go`
- [x] T011 [P] Create domain entities: MenuItem in `internal/domain/menu.go`
- [x] T012 [P] Create domain entities: Order, OrderItem in `internal/domain/order.go`
- [x] T013 [P] Create domain entities: Payment in `internal/domain/payment.go`
- [x] T014 [P] Create domain events: OrderPaid, OrderStatusChanged, PaymentFailed in `internal/domain/events.go`
- [x] T015 Create repository interfaces in `internal/domain/repositories.go`
- [x] T016 Implement in-memory RestaurantRepository with sync.RWMutex in `internal/infrastructure/repositories/memory/restaurant.go`
- [x] T017 [P] Implement in-memory MenuCategoryRepository in `internal/infrastructure/repositories/memory/menu_category.go`
- [x] T018 [P] Implement in-memory MenuItemRepository in `internal/infrastructure/repositories/memory/menu_item.go`
- [x] T019 [P] Implement in-memory OrderRepository in `internal/infrastructure/repositories/memory/order.go`
- [x] T020 [P] Implement in-memory PaymentRepository in `internal/infrastructure/repositories/memory/payment.go`
- [x] T021 Setup Watermill in-memory event bus in `internal/infrastructure/events/bus.go`
- [x] T022 [P] Create base Templ layout template in `internal/interfaces/templates/layout.go` including Datastar script tag (`<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.6/bundles/datastar.js"></script>`)
- [x] T023 [P] Create Echo server setup with HTTP/2 in `cmd/server/main.go`
- [x] T024 [P] Configure error handling middleware in `internal/interfaces/http/middleware/error.go`
- [x] T025 [P] Configure logging middleware in `internal/interfaces/http/middleware/logging.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Customer Orders Food with Lightning Payment (Priority: P1) 🎯 MVP

**Goal**: Enable customers to scan QR code, browse menu, add items to cart, pay with Lightning, and track order status in real-time

**Independent Test**: Scan restaurant QR code → browse menu → add items to cart → pay with Lightning → receive order confirmation with real-time status updates

### Tests for User Story 1 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Write tests → Get approval → Tests fail → Implement → Tests pass. Minimum 95% coverage required (critical path: payment processing, order management).**

- [x] T026 [P] [US1] Unit test for MenuItem display logic in `tests/unit/domain/menu_test.go` (80% coverage)
- [x] T027 [P] [US1] Unit test for Cart operations in `tests/unit/application/cart_test.go` (80% coverage)
- [x] T028 [P] [US1] Unit test for Payment domain entity in `tests/unit/domain/payment_test.go` (95% coverage - critical path)
- [x] T029 [P] [US1] Unit test for Order creation use case in `tests/unit/application/order_create_test.go` (95% coverage - critical path)
- [ ] T030 [P] [US1] Contract test for GET /menu endpoint in `tests/contract/menu_test.go`
- [ ] T031 [P] [US1] Contract test for POST /cart/add endpoint in `tests/contract/cart_test.go`
- [ ] T032 [P] [US1] Contract test for POST /payment/create-invoice endpoint in `tests/contract/payment_test.go`
- [ ] T033 [P] [US1] Contract test for GET /payment/status/{invoiceId} endpoint in `tests/contract/payment_test.go`
- [ ] T034 [P] [US1] Contract test for GET /order/{orderNumber} endpoint in `tests/contract/order_test.go`
- [ ] T035 [P] [US1] Contract test for GET /order/{orderNumber}/stream SSE endpoint in `tests/contract/order_sse_test.go`
- [ ] T036 [P] [US1] Integration test for Strike API invoice creation in `tests/integration/strike_test.go`
- [ ] T037 [P] [US1] Integration test for Strike API payment status polling in `tests/integration/strike_test.go`
- [ ] T038 [P] [US1] Integration test for Watermill event bus OrderPaid event in `tests/integration/events_test.go`
- [ ] T039 [P] [US1] Performance test for menu page load (<2s) in `tests/integration/performance_test.go`
- [ ] T040 [P] [US1] Performance test for payment flow (<10s end-to-end) in `tests/integration/performance_test.go`

### Implementation for User Story 1

- [x] T041 [P] [US1] Create Strike API client struct in `internal/infrastructure/strike/client.go`
- [x] T042 [P] [US1] Implement CreateInvoice method in Strike client in `internal/infrastructure/strike/invoice.go`
- [x] T043 [P] [US1] Implement GetInvoiceStatus method in Strike client in `internal/infrastructure/strike/invoice.go`
- [x] T044 [P] [US1] Implement GetExchangeRate method in Strike client in `internal/infrastructure/strike/exchange.go`
- [x] T045 [US1] Create GetMenu use case in `internal/application/menu/get_menu.go`
- [x] T046 [US1] Create AddToCart use case (session-based) in `internal/application/cart/add_to_cart.go`
- [x] T047 [US1] Create RemoveFromCart use case in `internal/application/cart/remove_from_cart.go`
- [x] T048 [US1] Create GetCart use case in `internal/application/cart/get_cart.go`
- [x] T049 [US1] Create CreatePaymentInvoice use case in `internal/application/payment/create_invoice.go`
- [x] T050 [US1] Create CheckPaymentStatus use case in `internal/application/payment/check_status.go`
- [x] T051 [US1] Create CreateOrder use case (triggered by OrderPaid event) in `internal/application/order/create_order.go`
- [x] T052 [US1] Create GetOrderByNumber use case in `internal/application/order/get_order.go`
- [x] T053 [US1] Implement GET /menu HTTP handler in `internal/interfaces/http/menu.go`
- [x] T054 [P] [US1] Implement POST /cart/add HTTP handler in `internal/interfaces/http/cart.go`
- [x] T055 [P] [US1] Implement POST /cart/remove HTTP handler in `internal/interfaces/http/cart.go`
- [x] T056 [P] [US1] Implement GET /cart HTTP handler in `internal/interfaces/http/cart.go`
- [x] T057 [US1] Implement POST /payment/create-invoice HTTP handler in `internal/interfaces/http/payment.go`
- [x] T058 [US1] Implement GET /payment/status/{invoiceId} HTTP handler in `internal/interfaces/http/payment.go`
- [x] T059 [US1] Implement GET /order/{orderNumber} HTTP handler in `internal/interfaces/http/order.go`
- [x] T060 [US1] Implement SSE handler for order status updates in `internal/interfaces/http/sse.go`
- [x] T061 [US1] Setup Watermill handler for OrderPaid event to create Order in `internal/infrastructure/events/handlers.go`
- [x] T062 [US1] Setup Watermill handler for OrderStatusChanged to stream SSE in `internal/infrastructure/events/handlers.go`
- [ ] T063 [P] [US1] Create menu display Templ template in `internal/interfaces/templates/menu.templ`
- [ ] T064 [P] [US1] Create cart display Templ template in `internal/interfaces/templates/cart.templ`
- [ ] T065 [P] [US1] Create payment invoice QR code Templ template in `internal/interfaces/templates/payment.templ`
- [ ] T066 [P] [US1] Create order status display Templ template in `internal/interfaces/templates/order_status.templ`
- [ ] T067 [US1] Add error handling for payment failures (FR-030) in `internal/application/payment/create_invoice.go`
- [ ] T068 [US1] Add order lookup by order number (FR-037) validation in `internal/application/order/get_order.go`
- [ ] T069 [US1] Verify menu load time <2 seconds on 3G (SC-003)
- [ ] T070 [US1] Verify payment completion <10 seconds (SC-002)
- [ ] T071 [US1] Verify code quality: functions <50 lines, classes <300 lines, cyclomatic complexity <10

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Customers can order food and pay with Lightning.

---

## Phase 4: User Story 2 - Kitchen Staff Fulfills Orders (Priority: P2)

**Goal**: Enable kitchen staff to view incoming orders, mark orders as preparing/ready, with real-time updates to customers

**Independent Test**: Kitchen staff views kitchen display → sees new paid orders → marks orders as preparing → marks orders as ready → customer receives notification

### Tests for User Story 2 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 95% coverage required (critical path: order management, kitchen display).**

- [ ] T072 [P] [US2] Unit test for UpdateOrderStatus use case in `tests/unit/application/order_update_test.go` (95% coverage - critical path)
- [ ] T073 [P] [US2] Unit test for GetActiveOrders use case in `tests/unit/application/order_query_test.go` (95% coverage - critical path)
- [ ] T074 [P] [US2] Contract test for GET /kitchen/{restaurantId} endpoint in `tests/contract/kitchen_test.go`
- [ ] T075 [P] [US2] Contract test for GET /kitchen/{restaurantId}/stream SSE endpoint in `tests/contract/kitchen_sse_test.go`
- [ ] T076 [P] [US2] Contract test for POST /kitchen/order/{orderId}/status endpoint in `tests/contract/kitchen_test.go`
- [ ] T077 [P] [US2] Integration test for SSE event streaming new orders in `tests/integration/sse_test.go`
- [ ] T078 [P] [US2] Integration test for SSE event streaming order status changes in `tests/integration/sse_test.go`
- [ ] T079 [P] [US2] Performance test for order status update propagation (<5s) in `tests/integration/performance_test.go`

### Implementation for User Story 2

- [ ] T080 [P] [US2] Create GetActiveOrders use case in `internal/application/order/get_active_orders.go`
- [ ] T081 [P] [US2] Create UpdateOrderStatus use case in `internal/application/order/update_status.go`
- [ ] T082 [US2] Create ArchiveCompletedOrders use case (after 1 hour) in `internal/application/order/archive_orders.go`
- [ ] T083 [US2] Implement GET /kitchen/{restaurantId} HTTP handler in `internal/interfaces/http/kitchen.go`
- [ ] T084 [US2] Implement GET /kitchen/{restaurantId}/stream SSE handler in `internal/interfaces/http/kitchen.go`
- [ ] T085 [US2] Implement POST /kitchen/order/{orderId}/status HTTP handler in `internal/interfaces/http/kitchen.go`
- [ ] T086 [US2] Setup Watermill handler for new OrderPaid event to stream to kitchen display in `internal/infrastructure/events/handlers.go`
- [ ] T087 [US2] Setup Watermill handler for OrderStatusChanged to notify customers in `internal/infrastructure/events/handlers.go`
- [ ] T088 [P] [US2] Create kitchen display Templ template in `internal/interfaces/templates/kitchen.templ`
- [ ] T089 [P] [US2] Add audible alert functionality for new orders (FR-012) in kitchen template
- [ ] T090 [US2] Add visual indicator for offline status (FR-016) in kitchen template
- [ ] T091 [US2] Implement order sorting by timestamp (oldest first) in GetActiveOrders use case
- [ ] T092 [US2] Add validation for order status transitions (paid→preparing→ready) in UpdateOrderStatus use case
- [ ] T093 [US2] Verify new orders appear within 5 seconds (SC-005)
- [ ] T094 [US2] Verify order status updates appear within 5 seconds (SC-004)
- [ ] T095 [US2] Verify code quality: functions <50 lines, classes <300 lines, cyclomatic complexity <10

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. End-to-end order flow is complete: customer orders → kitchen fulfills → customer picks up.

---

## Phase 5: User Story 3 - Owner Sets Up Restaurant Menu (Priority: P3)

**Goal**: Enable restaurant owners to create account, set up menu categories and items, upload photos, and generate customer QR code

**Independent Test**: Owner signs up → creates menu categories → adds items with photos → generates QR code → menu visible to customers

### Tests for User Story 3 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 80% coverage required.**

- [ ] T096 [P] [US3] Unit test for CreateRestaurant use case in `tests/unit/application/restaurant_test.go` (80% coverage)
- [ ] T097 [P] [US3] Unit test for CreateMenuCategory use case in `tests/unit/application/menu_category_test.go` (80% coverage)
- [ ] T098 [P] [US3] Unit test for CreateMenuItem use case in `tests/unit/application/menu_item_test.go` (80% coverage)
- [ ] T099 [P] [US3] Unit test for UpdateMenuItem use case in `tests/unit/application/menu_item_test.go` (80% coverage)
- [ ] T100 [P] [US3] Unit test for photo storage limit enforcement (100 photos) in `tests/unit/infrastructure/storage_test.go`
- [ ] T101 [P] [US3] Contract test for POST /dashboard/restaurant endpoint in `tests/contract/dashboard_test.go`
- [ ] T102 [P] [US3] Contract test for POST /dashboard/restaurant/{id}/status endpoint in `tests/contract/dashboard_test.go`
- [ ] T103 [P] [US3] Contract test for POST /dashboard/menu/category endpoint in `tests/contract/dashboard_test.go`
- [ ] T104 [P] [US3] Contract test for POST /dashboard/menu/item endpoint in `tests/contract/dashboard_test.go`
- [ ] T105 [P] [US3] Contract test for PUT /dashboard/menu/item/{id} endpoint in `tests/contract/dashboard_test.go`
- [ ] T106 [P] [US3] Contract test for GET /dashboard/restaurant/{id}/qr endpoint in `tests/contract/dashboard_test.go`
- [ ] T107 [P] [US3] Integration test for photo upload and optimization (2MB→300KB) in `tests/integration/storage_test.go`
- [ ] T108 [P] [US3] Integration test for 100 photo limit enforcement in `tests/integration/storage_test.go`

### Implementation for User Story 3

- [ ] T109 [P] [US3] Create photo storage interface in `internal/domain/storage.go`
- [ ] T110 [P] [US3] Implement local file storage in `internal/infrastructure/storage/local.go`
- [ ] T111 [P] [US3] Implement photo optimization (2MB→300KB) in `internal/infrastructure/storage/optimizer.go`
- [ ] T112 [US3] Create CreateRestaurant use case in `internal/application/restaurant/create.go`
- [ ] T113 [US3] Create UpdateRestaurantStatus use case (open/closed) in `internal/application/restaurant/update_status.go`
- [ ] T114 [US3] Create CreateMenuCategory use case in `internal/application/menu/create_category.go`
- [ ] T115 [US3] Create CreateMenuItem use case in `internal/application/menu/create_item.go`
- [ ] T116 [US3] Create UpdateMenuItem use case in `internal/application/menu/update_item.go`
- [ ] T117 [US3] Create UploadPhoto use case (validate 2MB limit, enforce 100 photo limit) in `internal/application/menu/upload_photo.go`
- [ ] T118 [US3] Create GenerateQRCode use case in `internal/application/restaurant/generate_qr.go`
- [ ] T119 [US3] Implement POST /dashboard/restaurant HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T120 [P] [US3] Implement POST /dashboard/restaurant/{id}/status HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T121 [P] [US3] Implement POST /dashboard/menu/category HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T122 [P] [US3] Implement POST /dashboard/menu/item HTTP handler (multipart/form-data) in `internal/interfaces/http/dashboard.go`
- [ ] T123 [P] [US3] Implement PUT /dashboard/menu/item/{id} HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T124 [P] [US3] Implement GET /dashboard/restaurant/{id}/qr HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T125 [P] [US3] Create dashboard menu management Templ template in `internal/interfaces/templates/dashboard_menu.templ`
- [ ] T126 [P] [US3] Create restaurant status toggle Templ template in `internal/interfaces/templates/dashboard_status.templ`
- [ ] T127 [P] [US3] Create QR code display Templ template in `internal/interfaces/templates/qr_code.templ`
- [ ] T128 [US3] Add validation for Lightning address format in CreateRestaurant use case
- [ ] T129 [US3] Add validation for photo count limit (100) in UploadPhoto use case
- [ ] T130 [US3] Add validation for photo size limit (2MB) in UploadPhoto use case
- [ ] T131 [US3] Update menu display to show "Currently Closed" banner when restaurant closed (FR-040, FR-041)
- [ ] T132 [US3] Verify menu setup completes in <10 minutes for 20 items (SC-006)
- [ ] T133 [US3] Verify menu changes reflect in customer view within 5 seconds (FR-023)
- [ ] T134 [US3] Verify code quality: functions <50 lines, classes <300 lines, cyclomatic complexity <10

**Checkpoint**: All user stories 1, 2, AND 3 should now work independently. Complete restaurant system operational: owner sets up menu → customers order → kitchen fulfills.

---

## Phase 6: User Story 4 - Owner Views Sales Dashboard (Priority: P4)

**Goal**: Enable restaurant owners to view daily sales, order history, top-selling items, and settlement status

**Independent Test**: Owner logs into dashboard → views sales metrics → sees order history → views top items → checks settlement status

### Tests for User Story 4 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 80% coverage required.**

- [ ] T135 [P] [US4] Unit test for GetSalesAnalytics use case in `tests/unit/application/analytics_test.go` (80% coverage)
- [ ] T136 [P] [US4] Unit test for GetOrderHistory use case in `tests/unit/application/analytics_test.go` (80% coverage)
- [ ] T137 [P] [US4] Unit test for GetTopItems use case in `tests/unit/application/analytics_test.go` (80% coverage)
- [ ] T138 [P] [US4] Unit test for GetSettlementStatus use case in `tests/unit/application/analytics_test.go` (80% coverage)
- [ ] T139 [P] [US4] Contract test for GET /dashboard/restaurant/{id}/analytics endpoint in `tests/contract/dashboard_analytics_test.go`

### Implementation for User Story 4

- [ ] T140 [P] [US4] Create GetSalesAnalytics use case (orders today, total sales, average order value) in `internal/application/analytics/get_sales.go`
- [ ] T141 [P] [US4] Create GetOrderHistory use case in `internal/application/analytics/get_orders.go`
- [ ] T142 [P] [US4] Create GetTopItems use case (ranked by quantity sold) in `internal/application/analytics/get_top_items.go`
- [ ] T143 [P] [US4] Create GetSettlementStatus use case in `internal/application/analytics/get_settlement.go`
- [ ] T144 [US4] Implement GET /dashboard/restaurant/{id}/analytics HTTP handler in `internal/interfaces/http/dashboard.go`
- [ ] T145 [P] [US4] Create analytics dashboard Templ template in `internal/interfaces/templates/dashboard_analytics.templ`
- [ ] T146 [P] [US4] Add date range filtering support in GetSalesAnalytics use case
- [ ] T147 [US4] Add daily settlement batch job (end of day 11:59 PM) in `internal/application/payment/settle_daily.go`
- [ ] T148 [US4] Verify analytics display orders, sales, average order value (FR-024)
- [ ] T149 [US4] Verify order history shows complete details (FR-025)
- [ ] T150 [US4] Verify top items ranked correctly (FR-026)
- [ ] T151 [US4] Verify settlement status displays correctly (FR-027)
- [ ] T152 [US4] Verify code quality: functions <50 lines, classes <300 lines, cyclomatic complexity <10

**Checkpoint**: All user stories should now be independently functional. Complete platform operational with analytics.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T153 [P] Documentation: Add Go doc comments for all public APIs in domain layer
- [ ] T154 [P] Documentation: Create API documentation from contracts in `docs/api.md`
- [ ] T155 [P] Documentation: Update README.md with setup instructions
- [ ] T156 Code cleanup: Verify all functions <50 lines, all structs <300 lines
- [ ] T157 Code cleanup: Run gocyclo and refactor any functions with complexity >10
- [ ] T158 Performance: Verify API endpoints <200ms p95 latency
- [ ] T159 Performance: Verify menu page load <2 seconds on 3G
- [ ] T160 Performance: Verify payment completion <10 seconds end-to-end
- [ ] T161 Performance: Verify order status updates <5 seconds
- [ ] T162 [P] Test coverage: Verify overall 80% coverage minimum
- [ ] T163 [P] Test coverage: Verify 95% coverage for payment processing
- [ ] T164 [P] Test coverage: Verify 95% coverage for order management
- [ ] T165 [P] Test coverage: Verify 95% coverage for kitchen display
- [ ] T166 Security: Run Go security scanner (gosec)
- [ ] T167 Security: Review Strike API integration for security best practices
- [ ] T168 UX: Verify consistent error messages across all templates
- [ ] T169 UX: Verify loading states for operations >200ms
- [ ] T170 UX: Verify accessibility (WCAG 2.1 AA) for all templates
- [ ] T171 UX: Test responsive design on mobile, tablet, desktop
- [ ] T172 Validation: Run quickstart.md validation (all setup steps work)
- [ ] T173 Validation: Test concurrent orders (FR-034)
- [ ] T174 Validation: Test offline menu browsing (PWA)
- [ ] T175 Validation: Test order lookup by order number (FR-037)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - RECOMMENDED MVP
  - User Story 2 (P2): Can start after Foundational - integrates with US1 for end-to-end flow
  - User Story 3 (P3): Can start after Foundational - prerequisite for US1/US2 (menu data)
  - User Story 4 (P4): Can start after Foundational - reads data from US1/US2/US3
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires US3 menu data to function (create sample menu or implement US3 first)
- **User Story 2 (P2)**: Requires US1 orders to display (can test with mock orders)
- **User Story 3 (P3)**: No dependencies - can be implemented independently
- **User Story 4 (P4)**: Requires US1/US2/US3 data to display analytics

**Recommended Implementation Order**:
1. Complete Foundational (Phase 2)
2. Implement User Story 3 (P3) - Menu setup (provides data for US1)
3. Implement User Story 1 (P1) - Customer ordering (MVP core)
4. Implement User Story 2 (P2) - Kitchen fulfillment (completes transaction flow)
5. Implement User Story 4 (P4) - Analytics (business insights)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Domain entities before use cases
- Use cases before HTTP handlers
- HTTP handlers before templates
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, User Stories 3 and 4 can start in parallel
- After US3 completes, User Stories 1 and 2 can proceed in parallel
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (write first, ensure they fail):
Task T026-T040: All unit, contract, integration, performance tests

# Launch all Strike API client methods in parallel:
Task T041: Strike client struct
Task T042: CreateInvoice method
Task T043: GetInvoiceStatus method
Task T044: GetExchangeRate method

# Launch all use cases that don't depend on each other:
Task T045: GetMenu use case
Task T046-T048: Cart use cases (parallel)
Task T049-T050: Payment use cases (parallel)

# Launch all HTTP handlers in parallel after use cases complete:
Task T053: GET /menu handler
Task T054-T056: Cart handlers (parallel)
Task T057-T058: Payment handlers (parallel)
Task T059-T060: Order handlers (parallel)

# Launch all templates in parallel:
Task T063-T066: All Templ templates (parallel)
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 3)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 5: User Story 3 (Menu Setup) - provides menu data
4. Complete Phase 3: User Story 1 (Customer Ordering) - MVP core
5. **STOP and VALIDATE**: Test end-to-end ordering flow independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 3 → Test independently → Menu management works
3. Add User Story 1 → Test independently → Customer ordering works (MVP!)
4. Add User Story 2 → Test independently → Kitchen fulfillment works
5. Add User Story 4 → Test independently → Analytics works
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 3 (Menu Setup)
   - Developer B: User Story 4 (Analytics) - parallel with US3
3. Once US3 completes:
   - Developer A: User Story 1 (Customer Ordering)
   - Developer B: User Story 2 (Kitchen Fulfillment) - parallel with US1
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests MUST be written first and fail before implementing (TDD per Constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Count Summary

- **Total Tasks**: 175
- **Setup (Phase 1)**: 8 tasks
- **Foundational (Phase 2)**: 17 tasks
- **User Story 1 (Phase 3)**: 46 tasks (15 tests + 31 implementation)
- **User Story 2 (Phase 4)**: 24 tasks (8 tests + 16 implementation)
- **User Story 3 (Phase 5)**: 39 tasks (13 tests + 26 implementation)
- **User Story 4 (Phase 6)**: 18 tasks (5 tests + 13 implementation)
- **Polish (Phase 7)**: 23 tasks

**Parallel Opportunities**: 89 tasks marked [P] can be executed in parallel

**MVP Scope**: User Story 1 + User Story 3 (85 tasks) - Customer ordering with menu management

**Independent Test Criteria**:
- US1: Scan QR → browse menu → add to cart → pay with Lightning → track order
- US2: View kitchen display → see orders → mark preparing/ready → customer notified
- US3: Sign up → create menu → upload photos → generate QR code
- US4: Login → view sales → see order history → check settlement status

