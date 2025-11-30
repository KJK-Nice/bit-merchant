# Tasks: Cash Payment with Hypermedia UI

**Input**: Design documents from `/specs/002-cash-payment-hypermedia/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Testing Standards), all features MUST include comprehensive tests. Test tasks are MANDATORY and MUST be written before implementation (TDD). Minimum 80% coverage required (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `internal/`, `tests/` at repository root
- Paths follow Clean Architecture structure from plan.md

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project structure per implementation plan in `internal/domain/`, `internal/application/`, `internal/infrastructure/`, `internal/interfaces/`
- [x] T002 Initialize Go module with dependencies: `github.com/labstack/echo/v4`, `github.com/a-h/templ`, `github.com/starfederation/datastar`, `github.com/ThreeDotsLabs/watermill`, `github.com/stretchr/testify`, `github.com/aws/aws-sdk-go-v2` (and related S3 modules)
- [x] T003 [P] Configure `golangci-lint` with strict rules in `.golangci.yml`
- [x] T004 [P] Configure `gocyclo` for complexity checking (<10 per function)
- [x] T005 [P] Setup Go test coverage reporting with minimum thresholds (80% standard, 95% critical paths)
- [x] T006 [P] Create `.env.example` with configuration template (PORT, RESTAURANT_ID, AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, S3_BUCKET_NAME, etc.)
- [x] T007 [P] Setup `cmd/server/main.go` skeleton with Echo server initialization
- [x] T008 [P] Create `static/pwa/` directory structure for PWA manifest and service worker

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Domain Entities (Zero Dependencies)

- [x] T009 [P] Create `Restaurant` entity in `internal/domain/restaurant.go` with ID, Name, IsOpen, ClosedMessage, ReopeningHours, CreatedAt, UpdatedAt
- [x] T010 [P] Create `MenuCategory` entity in `internal/domain/menu.go` with ID, RestaurantID, Name, DisplayOrder, IsActive, CreatedAt, UpdatedAt
- [x] T011 [P] Create `MenuItem` entity in `internal/domain/menu.go` with ID, CategoryID, RestaurantID, Name, Description, Price, PhotoURL, PhotoOriginalURL, IsAvailable, CreatedAt, UpdatedAt
- [x] T012 [P] Create `Order` entity in `internal/domain/order.go` with ID, OrderNumber, RestaurantID, PaymentMethodType, PaymentStatus, FulfillmentStatus, TotalAmount, Items, CreatedAt, PaidAt, PreparingAt, ReadyAt, CompletedAt
- [x] T013 [P] Create `OrderItem` entity in `internal/domain/order.go` with ID, OrderID, MenuItemID, Name, Price, Quantity, Subtotal
- [x] T014 [P] Create `Payment` entity in `internal/domain/payment.go` with ID, OrderID, RestaurantID, PaymentMethodType, Amount, Status, CreatedAt, PaidAt, FailedAt, FailureReason
- [x] T015 [P] Create value objects: `PaymentStatus`, `FulfillmentStatus`, `PaymentMethodType` enums in `internal/domain/order.go` and `internal/domain/payment.go`
- [x] T016 [P] Create domain events: `OrderCreated`, `OrderPaid`, `OrderPreparing`, `OrderReady`, `OrderCompleted` in `internal/domain/events.go`

### Repository Interfaces (Domain Layer)

- [x] T017 [P] Create `RestaurantRepository` interface in `internal/domain/restaurant.go` with GetByID, Create, Update methods
- [x] T018 [P] Create `MenuCategoryRepository` interface in `internal/domain/menu.go` with GetByRestaurantID, GetByID, Create, Update methods
- [x] T019 [P] Create `MenuItemRepository` interface in `internal/domain/menu.go` with GetByCategoryID, GetByRestaurantID, GetByID, Create, Update methods
- [x] T020 [P] Create `OrderRepository` interface in `internal/domain/order.go` with GetByID, GetByOrderNumber, GetByRestaurantID, GetPendingByRestaurantID, Create, Update methods
- [x] T021 [P] Create `PaymentRepository` interface in `internal/domain/payment.go` with GetByID, GetByOrderID, Create, Update methods

### Payment Method Abstraction (Domain Layer)

- [x] T022 [P] Create `PaymentMethod` interface in `internal/domain/payment.go` with ProcessPayment, ValidatePayment, GetPaymentMethodType methods
- [x] T023 [P] Create `CashPaymentMethod` implementation in `internal/infrastructure/payment/cash/payment.go` implementing PaymentMethod interface

### In-Memory Repository Implementations (Infrastructure Layer)

- [x] T024 [P] Implement `RestaurantRepository` in-memory in `internal/infrastructure/repositories/memory/restaurant.go` with sync.RWMutex
- [x] T025 [P] Implement `MenuCategoryRepository` in-memory in `internal/infrastructure/repositories/memory/menu.go` with sync.RWMutex
- [x] T026 [P] Implement `MenuItemRepository` in-memory in `internal/infrastructure/repositories/memory/menu.go` with sync.RWMutex
- [x] T027 [P] Implement `OrderRepository` in-memory in `internal/infrastructure/repositories/memory/order.go` with sync.RWMutex
- [x] T028 [P] Implement `PaymentRepository` in-memory in `internal/infrastructure/repositories/memory/payment.go` with sync.RWMutex

### Event Bus (Infrastructure Layer)

- [x] T029 [P] Setup Watermill in-memory event bus in `internal/infrastructure/events/bus.go` with pub/sub configuration
- [x] T030 [P] Create event publisher wrapper in `internal/infrastructure/events/publisher.go` for domain events
- [x] T031 [P] Create event subscriber wrapper in `internal/infrastructure/events/subscriber.go` for SSE handlers

### Cart Management (Session-Based)

- [x] T032 [P] Create ephemeral cart service in `internal/application/cart/cart.go` with AddItem, RemoveItem, GetCart, ClearCart methods (session-based, not persisted)

### Base HTTP Handlers & Middleware

- [x] T033 [P] Create session middleware in `internal/interfaces/http/middleware/session.go` for session ID management
- [x] T034 [P] Create error handling middleware in `internal/interfaces/http/middleware/error.go` for HTML error pages
- [x] T035 [P] Create logging middleware in `internal/interfaces/http/middleware/logging.go` for request logging
- [x] T036 [P] Create structured logging service in `internal/infrastructure/logging/logger.go` (JSON format, log levels, rotation policy) for FR-046
- [x] T037 [P] Integrate logging service in CreateOrderUseCase to log order creation events (FR-046)
- [x] T038 [P] Integrate logging service in MarkOrderPaidUseCase and MarkOrderPreparingUseCase to log payment/fulfillment status changes (FR-046)
- [x] T039 Setup Echo routes structure in `cmd/server/main.go` with customer, kitchen, and owner route groups

### PWA Infrastructure

- [x] T040 [P] Create PWA manifest in `static/pwa/manifest.json` with app name, icons, display mode
- [x] T041 [P] Create service worker in `static/pwa/sw.js` for offline menu browsing (cache menu HTML and images)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Customer Orders Food with Cash Payment (Priority: P1) 🎯 MVP

**Goal**: Customer can scan QR code, browse menu, add items to cart, confirm cash payment, receive order number, and track order status in real-time.

**Independent Test**: (1) Scan restaurant QR code, (2) Browse menu, (3) Add items to cart, (4) Confirm cash payment, (5) Receive order confirmation with real-time status updates. Delivers complete end-to-end ordering experience.

### Tests for User Story 1 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Write tests → Get approval → Tests fail → Implement → Tests pass. Minimum 95% coverage required (critical path).**

- [x] T042 [P] [US1] Unit tests for Restaurant entity validation in `tests/unit/domain/restaurant_test.go` (95% coverage)
- [x] T043 [P] [US1] Unit tests for MenuCategory entity validation in `tests/unit/domain/menu_test.go` (95% coverage)
- [x] T044 [P] [US1] Unit tests for MenuItem entity validation in `tests/unit/domain/menu_test.go` (95% coverage)
- [x] T045 [P] [US1] Unit tests for Order entity and state transitions in `tests/unit/domain/order_test.go` (95% coverage)
- [x] T046 [P] [US1] Unit tests for Payment entity and state transitions in `tests/unit/domain/payment_test.go` (95% coverage)
- [x] T047 [P] [US1] Unit tests for Cart service (AddItem, RemoveItem, GetCart) in `tests/unit/application/cart/cart_test.go` (95% coverage)
- [x] T048 [P] [US1] Unit tests for CashPaymentMethod in `tests/unit/infrastructure/payment/cash/payment_test.go` (95% coverage)
- [x] T049 [P] [US1] Unit tests for in-memory RestaurantRepository in `tests/unit/infrastructure/repositories/memory/restaurant_test.go` (95% coverage)
- [x] T050 [P] [US1] Unit tests for in-memory MenuCategoryRepository in `tests/unit/infrastructure/repositories/memory/menu_test.go` (95% coverage)
- [x] T051 [P] [US1] Unit tests for in-memory MenuItemRepository in `tests/unit/infrastructure/repositories/memory/menu_test.go` (95% coverage)
- [x] T052 [P] [US1] Unit tests for in-memory OrderRepository in `tests/unit/infrastructure/repositories/memory/order_test.go` (95% coverage)
- [x] T053 [P] [US1] Unit tests for in-memory PaymentRepository in `tests/unit/infrastructure/repositories/memory/payment_test.go` (95% coverage)
- [x] T054 [P] [US1] Contract test for GET /menu endpoint in `tests/contract/http/menu_test.go` (validates HTML structure, performance <2s)
- [x] T055 [P] [US1] Contract test for POST /cart/add endpoint in `tests/contract/http/cart_test.go` (validates HTML fragment response, Datastar attributes)
- [x] T056 [P] [US1] Contract test for POST /cart/remove endpoint in `tests/contract/http/cart_test.go` (validates HTML fragment response)
- [x] T057 [P] [US1] Contract test for GET /order/confirm endpoint in `tests/contract/http/order_test.go` (validates HTML structure)
- [x] T058 [P] [US1] Contract test for POST /order/create endpoint in `tests/contract/http/order_test.go` (validates HTML response, order creation)
- [x] T059 [P] [US1] Contract test for GET /order/:orderNumber endpoint in `tests/contract/http/order_test.go` (validates order lookup)
- [x] T060 [P] [US1] Integration test for complete ordering flow in `tests/integration/order/order_flow_test.go` (browse → cart → confirm → order)
- [x] T061 [P] [US1] Integration test for SSE order status stream in `tests/integration/sse/order_stream_test.go` (validates SSE events, <5s propagation)
- [x] T062 [P] [US1] Performance test for menu page load in `tests/performance/menu_load_test.go` (<2s on 3G simulation)
- [x] T063 [P] [US1] Performance test for complete ordering flow in `tests/performance/order_flow_test.go` (<2min total)

### Implementation for User Story 1

#### Application Layer (Use Cases)

- [x] T064 [P] [US1] Create GetMenuUseCase in `internal/application/menu/get_menu.go` (returns restaurant menu with categories and items)
- [x] T065 [US1] Create CreateOrderUseCase in `internal/application/order/create_order.go` (creates order with cash payment, publishes OrderCreated event, logs order creation)
- [x] T066 [US1] Create GetOrderByNumberUseCase in `internal/application/order/get_order.go` (looks up order by order number)

#### HTTP Handlers (Return HTML, Not JSON)

- [x] T067 [P] [US1] Create MenuHandler with GET /menu endpoint in `internal/interfaces/http/menu.go` (returns HTML page via Templ template)
- [x] T068 [P] [US1] Create CartHandler with POST /cart/add, POST /cart/remove, GET /cart endpoints in `internal/interfaces/http/cart.go` (returns HTML fragments for Datastar updates)
- [x] T069 [P] [US1] Create OrderHandler with GET /order/confirm, POST /order/create, GET /order/:orderNumber endpoints in `internal/interfaces/http/order.go` (returns HTML pages)
- [x] T070 [US1] Create SSE handler for GET /order/:orderNumber/stream endpoint in `internal/interfaces/http/sse.go` (streams order status updates via Datastar)

#### Templ Templates (Server-Rendered HTML)

- [x] T071 [P] [US1] Create menu page template in `internal/interfaces/templates/menu.templ` (displays categories, items, prices, photos, cart summary, uses templui.io components)
- [x] T072 [P] [US1] Create cart summary fragment template in `internal/interfaces/templates/components/cart_summary.templ` (for Datastar partial updates)
- [x] T073 [P] [US1] Create order confirmation page template in `internal/interfaces/templates/order_confirm.templ` (shows order summary, cash payment confirmation form)
- [x] T074 [P] [US1] Create order status page template in `internal/interfaces/templates/order_status.templ` (shows order number, status, SSE connection via Datastar ds-sse-connect)
- [x] T075 [P] [US1] Create reusable UI components using templui.io patterns in `internal/interfaces/templates/components/` (buttons, forms, cards)

#### Datastar Integration

- [x] T076 [US1] Integrate Datastar library in `internal/interfaces/http/middleware/datastar.go` (setup Datastar middleware for Echo)
- [x] T077 [US1] Add Datastar attributes (ds-post, ds-target) to cart forms in `internal/interfaces/templates/menu.templ`
- [x] T078 [US1] Add Datastar SSE connection (ds-sse-connect) to order status page in `internal/interfaces/templates/order_status.templ`

#### Event Handling

- [x] T079 [US1] Create OrderCreated event handler in `internal/infrastructure/events/handlers/order_created.go` (publishes to SSE stream for kitchen display)
- [x] T080 [US1] Create OrderPaid event handler in `internal/infrastructure/events/handlers/order_paid.go` (publishes to customer SSE stream, logs payment status change)
- [x] T081 [US1] Create OrderPreparing event handler in `internal/infrastructure/events/handlers/order_preparing.go` (publishes to customer SSE stream, logs fulfillment status change)
- [x] T082 [US1] Create OrderReady event handler in `internal/infrastructure/events/handlers/order_ready.go` (publishes to customer SSE stream, logs fulfillment status change)

#### Route Registration

- [x] T083 [US1] Register customer routes (GET /menu, POST /cart/add, POST /cart/remove, GET /cart, GET /order/confirm, POST /order/create, GET /order/:orderNumber, GET /order/:orderNumber/stream) in `cmd/server/main.go`

#### Code Quality Verification

- [x] T084 [US1] Verify code quality: functions <50 lines, types/structs <300 lines, cyclomatic complexity <10 using gocyclo
- [x] T085 [US1] Verify test coverage: 95% for all US1 components (critical path)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Customer can complete full ordering flow with real-time status updates.

---

## Phase 4: User Story 2 - Kitchen Staff Fulfills Orders (Priority: P2)

**Goal**: Kitchen staff can view incoming orders, mark payments received, prepare food, and notify customers when orders are ready - all through a simple HTML interface.

**Independent Test**: (1) Kitchen staff views kitchen display HTML page, (2) Sees new orders appear automatically, (3) Marks orders as "Paid" when cash received, (4) Marks orders as "Preparing", (5) Marks orders as "Ready", (6) Verifies customers receive real-time updates.

### Tests for User Story 2 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 95% coverage required (critical path).**

- [x] T086 [P] [US2] Unit tests for MarkOrderPaidUseCase in `tests/unit/application/kitchen/mark_paid_test.go` (95% coverage, validates state transitions)
- [x] T087 [P] [US2] Unit tests for MarkOrderPreparingUseCase in `tests/unit/application/kitchen/mark_preparing_test.go` (95% coverage, validates payment requirement)
- [x] T088 [P] [US2] Unit tests for MarkOrderReadyUseCase in `tests/unit/application/kitchen/mark_ready_test.go` (95% coverage, validates state transitions)
- [x] T089 [P] [US2] Unit tests for GetKitchenOrdersUseCase in `tests/unit/application/kitchen/get_orders_test.go` (95% coverage, validates chronological ordering)
- [x] T090 [P] [US2] Contract test for GET /kitchen endpoint in `tests/contract/http/kitchen_test.go` (validates HTML structure, order display)
- [x] T091 [P] [US2] Contract test for GET /kitchen/stream endpoint in `tests/contract/http/kitchen_test.go` (validates SSE events for new orders)
- [x] T092 [P] [US2] Contract test for POST /kitchen/order/:id/mark-paid endpoint in `tests/contract/http/kitchen_test.go` (validates HTML fragment response, state transition)
- [x] T093 [P] [US2] Contract test for POST /kitchen/order/:id/mark-preparing endpoint in `tests/contract/http/kitchen_test.go` (validates HTML fragment response, state transition)
- [x] T094 [P] [US2] Contract test for POST /kitchen/order/:id/mark-ready endpoint in `tests/contract/http/kitchen_test.go` (validates HTML fragment response, state transition)
- [x] T095 [P] [US2] Integration test for kitchen workflow in `tests/integration/kitchen/kitchen_workflow_test.go` (order appears → mark paid → mark preparing → mark ready → customer receives updates)
- [x] T096 [P] [US2] Integration test for kitchen SSE stream in `tests/integration/sse/kitchen_stream_test.go` (validates new-order and order-updated events, <5s propagation)
- [x] T097 [P] [US2] Performance test for kitchen display updates in `tests/performance/kitchen_updates_test.go` (<5s for new orders to appear)
- [x] T098 [P] [US2] Integration test for kitchen offline sync in `tests/integration/kitchen/offline_sync_test.go` (validates status changes queued when offline, synced when reconnected) for FR-017

### Implementation for User Story 2

#### Application Layer (Use Cases)

- [x] T099 [P] [US2] Create GetKitchenOrdersUseCase in `internal/application/kitchen/get_orders.go` (returns orders in chronological order, filters by restaurant)
- [x] T100 [US2] Create MarkOrderPaidUseCase in `internal/application/kitchen/mark_paid.go` (updates payment status, publishes OrderPaid event, logs payment status change)
- [x] T101 [US2] Create MarkOrderPreparingUseCase in `internal/application/kitchen/mark_preparing.go` (validates payment confirmed, updates fulfillment status, publishes OrderPreparing event, logs fulfillment status change)
- [x] T102 [US2] Create MarkOrderReadyUseCase in `internal/application/kitchen/mark_ready.go` (updates fulfillment status, publishes OrderReady event, logs fulfillment status change)

#### HTTP Handlers (Return HTML, Not JSON)

- [x] T103 [P] [US2] Create KitchenHandler with GET /kitchen endpoint in `internal/interfaces/http/kitchen.go` (returns HTML page with orders)
- [x] T104 [US2] Add POST /kitchen/order/:id/mark-paid endpoint to KitchenHandler in `internal/interfaces/http/kitchen.go` (returns HTML fragment for Datastar update)
- [x] T105 [US2] Add POST /kitchen/order/:id/mark-preparing endpoint to KitchenHandler in `internal/interfaces/http/kitchen.go` (returns HTML fragment)
- [x] T106 [US2] Add POST /kitchen/order/:id/mark-ready endpoint to KitchenHandler in `internal/interfaces/http/kitchen.go` (returns HTML fragment)
- [x] T107 [US2] Add GET /kitchen/stream SSE endpoint to SSE handler in `internal/interfaces/http/sse.go` (streams kitchen order updates)
- [x] T108 [US2] Create offline queue service in `internal/infrastructure/kitchen/offline_queue.go` (queues status changes when connection lost, syncs when restored) for FR-017
- [x] T109 [US2] Add offline status indicator to kitchen template in `internal/interfaces/templates/kitchen.templ` (shows visual indicator when offline, syncs when reconnected) for FR-017

#### Templ Templates (Server-Rendered HTML)

- [x] T110 [P] [US2] Create kitchen display page template in `internal/interfaces/templates/kitchen.templ` (shows orders in chronological order, uses Datastar ds-sse-connect for updates)
- [x] T111 [P] [US2] Create order card fragment template in `internal/interfaces/templates/components/order_card.templ` (for Datastar partial updates, shows payment/fulfillment status, action buttons)
- [x] T112 [US2] Add audible/visual alert logic to kitchen template when new order arrives (JavaScript-free, CSS-based, HTML5 audio for audible alert) for FR-012

#### Datastar Integration

- [x] T113 [US2] Add Datastar attributes (ds-post, ds-target="closest article") to kitchen action forms in `internal/interfaces/templates/kitchen.templ`
- [x] T114 [US2] Add Datastar SSE connection (ds-sse-connect="/kitchen/stream") to kitchen display page in `internal/interfaces/templates/kitchen.templ`

#### Event Handling

- [x] T115 [US2] Update OrderCreated event handler to publish to kitchen SSE stream in `internal/infrastructure/events/handlers/order_created.go`
- [x] T116 [US2] Update OrderPaid event handler to publish to kitchen SSE stream in `internal/infrastructure/events/handlers/order_paid.go`
- [x] T117 [US2] Update OrderPreparing event handler to publish to kitchen SSE stream in `internal/infrastructure/events/handlers/order_preparing.go`
- [x] T118 [US2] Update OrderReady event handler to publish to kitchen SSE stream in `internal/infrastructure/events/handlers/order_ready.go`

#### Route Registration

- [x] T119 [US2] Register kitchen routes (GET /kitchen, GET /kitchen/stream, POST /kitchen/order/:id/mark-paid, POST /kitchen/order/:id/mark-preparing, POST /kitchen/order/:id/mark-ready) in `cmd/server/main.go`

#### Code Quality Verification

- [x] T120 [US2] Verify code quality: functions <50 lines, types/structs <300 lines, cyclomatic complexity <10
- [x] T121 [US2] Verify test coverage: 95% for all US2 components (critical path)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Kitchen staff can fulfill orders and customers receive real-time updates.

---

## Phase 5: User Story 3 - Owner Sets Up Restaurant Menu (Priority: P3)

**Goal**: Restaurant owner can create account, set up menu categories and items with photos, generate QR code, and manage menu through HTML interface.

**Independent Test**: (1) Owner signs up, (2) Enters restaurant name, (3) Creates menu categories, (4) Adds items with names, descriptions, prices, photos, (5) Generates customer-facing QR code. Delivers complete menu management.

### Tests for User Story 3 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 80% coverage required.**

- [x] T122 [P] [US3] Unit tests for CreateRestaurantUseCase in `tests/unit/application/restaurant/create_test.go` (80% coverage)
- [x] T123 [P] [US3] Unit tests for CreateMenuCategoryUseCase in `tests/unit/application/menu/create_category_test.go` (80% coverage)
- [x] T124 [P] [US3] Unit tests for CreateMenuItemUseCase in `tests/unit/application/menu/create_item_test.go` (80% coverage)
- [x] T125 [P] [US3] Unit tests for UpdateMenuItemUseCase in `tests/unit/application/menu/update_item_test.go` (80% coverage)
- [x] T126 [P] [US3] Unit tests for photo upload/optimization service in `tests/unit/infrastructure/storage/photo_test.go` (80% coverage, validates 2MB limit, 300KB optimization, 100 photo limit)
- [x] T127 [P] [US3] Contract test for GET /owner/signup endpoint in `tests/contract/http/owner_test.go` (validates HTML structure)
- [x] T128 [P] [US3] Contract test for POST /owner/signup endpoint in `tests/contract/http/owner_test.go` (validates account creation)
- [x] T129 [P] [US3] Contract test for GET /dashboard/menu endpoint in `tests/contract/http/owner_test.go` (validates menu management interface)
- [x] T130 [P] [US3] Contract test for POST /dashboard/menu/category endpoint in `tests/contract/http/owner_test.go` (validates category creation, HTML fragment response)
- [x] T131 [P] [US3] Contract test for POST /dashboard/menu/item endpoint in `tests/contract/http/owner_test.go` (validates item creation with photo upload, HTML fragment response, <30s for item, <10s for photo)
- [x] T132 [P] [US3] Contract test for POST /dashboard/menu/item/:id/photo endpoint in `tests/contract/http/owner_test.go` (validates photo upload, optimization, 2MB limit, 100 photo limit enforcement)
- [x] T133 [P] [US3] Contract test for GET /dashboard/qr-code endpoint in `tests/contract/http/owner_test.go` (validates QR code generation)
- [x] T134 [P] [US3] Integration test for complete menu setup flow in `tests/integration/menu/setup_test.go` (signup → create category → add items → upload photos → generate QR)
- [x] T135 [P] [US3] Integration test for photo storage and optimization in `tests/integration/storage/photo_test.go` (validates upload, compression to 300KB, URL generation, 100 photo limit)
- [x] T136 [P] [US3] Integration test for menu cache invalidation in `tests/integration/menu/cache_test.go` (validates menu changes reflect in customer view within 5 seconds) for FR-024

### Implementation for User Story 3

#### Application Layer (Use Cases)

- [x] T137 [P] [US3] Create CreateRestaurantUseCase in `internal/application/restaurant/create.go` (creates restaurant account with name)
- [x] T138 [P] [US3] Create CreateMenuCategoryUseCase in `internal/application/menu/create_category.go` (creates category with name and display order, invalidates menu cache)
- [x] T139 [P] [US3] Create CreateMenuItemUseCase in `internal/application/menu/create_item.go` (creates item with name, description, price, category, invalidates menu cache)
- [x] T140 [P] [US3] Create UpdateMenuItemUseCase in `internal/application/menu/update_item.go` (updates item details, marks out of stock, invalidates menu cache)
- [x] T141 [P] [US3] Create UploadMenuItemPhotoUseCase in `internal/application/menu/upload_photo.go` (handles photo upload, optimization, storage, validates 2MB limit, enforces 100 photo limit per restaurant, invalidates menu cache) for FR-049

#### Photo Storage Infrastructure

- [x] T142 [P] [US3] Create photo storage interface in `internal/infrastructure/storage/photo.go` (abstracts S3 operations)
- [x] T143 [P] [US3] Implement photo optimization service in `internal/infrastructure/storage/photo_optimizer.go` (compresses to 300KB, validates 2MB upload limit)
- [x] T144 [US3] Implement photo storage service using AWS S3 SDK in `internal/infrastructure/storage/s3_storage.go` (stores original and optimized versions in S3 bucket)
- [x] T145 [P] [US3] Create photo count service in `internal/infrastructure/storage/photo_count.go` (counts photos per restaurant, validates 100 photo limit before upload) for FR-049

#### HTTP Handlers (Return HTML, Not JSON)

- [x] T146 [P] [US3] Create OwnerHandler with GET /owner/signup, POST /owner/signup endpoints in `internal/interfaces/http/owner.go` (returns HTML pages)
- [x] T147 [P] [US3] Add GET /dashboard/menu endpoint to OwnerHandler in `internal/interfaces/http/owner.go` (returns menu management HTML page)
- [x] T148 [P] [US3] Add POST /dashboard/menu/category endpoint to OwnerHandler in `internal/interfaces/http/owner.go` (returns HTML fragment for Datastar update)
- [x] T149 [P] [US3] Add POST /dashboard/menu/item endpoint to OwnerHandler in `internal/interfaces/http/owner.go` (handles multipart form data, returns HTML fragment)
- [x] T150 [P] [US3] Add POST /dashboard/menu/item/:id/photo endpoint to OwnerHandler in `internal/interfaces/http/owner.go` (handles photo upload, validates 100 photo limit, returns HTML fragment)
- [x] T151 [P] [US3] Add GET /dashboard/qr-code endpoint to OwnerHandler in `internal/interfaces/http/owner.go` (generates QR code image, returns HTML page)

#### QR Code Generation

- [x] T152 [P] [US3] Create QR code generation service in `internal/infrastructure/qr/qr.go` (generates QR code linking to restaurant menu URL)

#### Menu Cache Management

- [x] T153 [P] [US3] Create menu cache service in `internal/infrastructure/cache/menu_cache.go` (in-memory cache for menu data, invalidates on menu changes) for FR-024
- [x] T154 [US3] Integrate menu cache in GetMenuUseCase to serve cached menu data and invalidate on updates for FR-024

#### Templ Templates (Server-Rendered HTML)

- [x] T155 [P] [US3] Create owner signup page template in `internal/interfaces/templates/owner_signup.templ` (simple form for restaurant name)
- [x] T156 [P] [US3] Create menu management page template in `internal/interfaces/templates/menu_manage.templ` (shows categories, items, add/edit forms, uses templui.io components) for FR-022
- [x] T157 [P] [US3] Create category form fragment template in `internal/interfaces/templates/components/category_form.templ` (for Datastar partial updates)
- [x] T158 [P] [US3] Create item form fragment template in `internal/interfaces/templates/components/item_form.templ` (for Datastar partial updates, includes photo upload, edit mode)
- [x] T159 [P] [US3] Create QR code display page template in `internal/interfaces/templates/qr_code.templ` (shows QR code image, printable format, shareable link)

#### Datastar Integration

- [x] T160 [US3] Add Datastar attributes (ds-post, ds-target) to menu management forms in `internal/interfaces/templates/menu_manage.templ`

#### Route Registration

- [x] T161 [US3] Register owner routes (GET /owner/signup, POST /owner/signup, GET /dashboard/menu, POST /dashboard/menu/category, POST /dashboard/menu/item, POST /dashboard/menu/item/:id/photo, GET /dashboard/qr-code) in `cmd/server/main.go`

#### Code Quality Verification

- [x] T162 [US3] Verify code quality: functions <50 lines, types/structs <300 lines, cyclomatic complexity <10
- [x] T163 [US3] Verify test coverage: 80% for all US3 components

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently. Owners can set up menus and customers can order from them.

---

## Phase 6: User Story 4 - Owner Views Sales Dashboard (Priority: P4)

**Goal**: Restaurant owner can view daily sales, order count, top-selling items, and payment status through HTML dashboard.

**Independent Test**: (1) Owner logs into dashboard HTML page, (2) Views today's sales and order count, (3) Sees average order value, (4) Views top-selling items, (5) Confirms payment status. Delivers business insights.

### Tests for User Story 4 (MANDATORY - Constitution Principle II) ⚠️

> **NOTE: Per Constitution, tests MUST be written FIRST using TDD. Minimum 80% coverage required.**

- [x] T164 [P] [US4] Unit tests for GetDashboardStatsUseCase in `tests/unit/application/dashboard/get_stats_test.go` (80% coverage, validates calculations)
- [x] T165 [P] [US4] Unit tests for GetOrderHistoryUseCase in `tests/unit/application/dashboard/get_history_test.go` (80% coverage)
- [x] T166 [P] [US4] Unit tests for GetTopSellingItemsUseCase in `tests/unit/application/dashboard/get_top_items_test.go` (80% coverage, validates ranking)
- [x] T167 [P] [US4] Unit tests for ToggleRestaurantOpenUseCase in `tests/unit/application/restaurant/toggle_open_test.go` (80% coverage)
- [x] T168 [P] [US4] Contract test for GET /dashboard endpoint in `tests/contract/http/dashboard_test.go` (validates HTML structure, stats display, weekly summary)
- [x] T169 [P] [US4] Contract test for POST /dashboard/toggle-open endpoint in `tests/contract/http/dashboard_test.go` (validates restaurant status toggle)
- [x] T170 [P] [US4] Integration test for dashboard data accuracy in `tests/integration/dashboard/stats_test.go` (validates calculations match actual orders)

### Implementation for User Story 4

#### Application Layer (Use Cases)

- [x] T171 [P] [US4] Create GetDashboardStatsUseCase in `internal/application/dashboard/get_stats.go` (calculates orders today, total sales, average order value, supports date range for weekly summary)
- [x] T172 [P] [US4] Create GetOrderHistoryUseCase in `internal/application/dashboard/get_history.go` (returns orders with timestamps, items, amounts, statuses, supports date range filtering)
- [x] T173 [P] [US4] Create GetTopSellingItemsUseCase in `internal/application/dashboard/get_top_items.go` (ranks items by quantity sold, calculates revenue per item)
- [x] T174 [P] [US4] Create ToggleRestaurantOpenUseCase in `internal/application/restaurant/toggle_open.go` (toggles IsOpen status, updates ClosedMessage/ReopeningHours)

#### HTTP Handlers (Return HTML, Not JSON)

- [x] T175 [P] [US4] Create DashboardHandler with GET /dashboard endpoint in `internal/interfaces/http/dashboard.go` (returns HTML page with stats, supports date range query parameter for weekly summary)
- [x] T176 [P] [US4] Add POST /dashboard/toggle-open endpoint to DashboardHandler in `internal/interfaces/http/dashboard.go` (returns updated dashboard HTML)

#### Templ Templates (Server-Rendered HTML)

- [x] T177 [P] [US4] Create dashboard page template in `internal/interfaces/templates/dashboard.templ` (shows stats, order history table, top items list, toggle open/closed button, date range selector for weekly summary, uses templui.io components)

#### Route Registration

- [x] T178 [US4] Register dashboard routes (GET /dashboard, POST /dashboard/toggle-open) in `cmd/server/main.go`

#### Code Quality Verification

- [x] T179 [US4] Verify code quality: functions <50 lines, types/structs <300 lines, cyclomatic complexity <10
- [x] T180 [US4] Verify test coverage: 80% for all US4 components

**Checkpoint**: All user stories should now be independently functional. Complete system supports customer ordering, kitchen fulfillment, menu management, and analytics.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Error Handling & Error Pages

- [x] T181 [P] Create 404 Not Found HTML page template in `internal/interfaces/templates/errors/404.templ`
- [x] T182 [P] Create 400 Bad Request HTML page template in `internal/interfaces/templates/errors/400.templ`
- [x] T183 [P] Create 503 Service Unavailable HTML page template in `internal/interfaces/templates/errors/503.templ`
- [x] T184 [P] Update error handling middleware to render appropriate error pages in `internal/interfaces/http/middleware/error.go`

### Restaurant Closed State

- [x] T185 [P] Add "Currently Closed" banner to menu page when restaurant is closed in `internal/interfaces/templates/menu.templ` (shows ClosedMessage and ReopeningHours)
- [x] T186 [P] Disable ordering functionality when restaurant is closed (validate in CreateOrderUseCase, return 503 error)

### Order Lookup & Recovery

- [x] T187 [P] Add order lookup by order number functionality (already implemented in US1, verify edge cases)
- [x] T188 [P] Add order recovery page template in `internal/interfaces/templates/order_lookup.templ` (allows customers to enter order number)

### Order Archiving (FR-016) - POST-MVP

**Note**: FR-016 is marked as post-MVP in spec. These tasks can be deferred to post-launch polish or future iteration.

- [ ] T189 [P] Create order archiving service in `internal/application/order/archive.go` (moves completed orders to archived state after 1 hour) for FR-016
- [ ] T190 [P] Create scheduled job for order archiving in `internal/infrastructure/jobs/order_archiver.go` (runs periodically, archives orders marked Ready after 1 hour) for FR-016
- [ ] T191 [P] Add archived orders query to OrderRepository interface in `internal/domain/order.go` (GetArchivedByRestaurantID method) for FR-016
- [ ] T192 [P] Implement archived orders query in in-memory OrderRepository in `internal/infrastructure/repositories/memory/order.go` for FR-016

### PWA Enhancements

- [x] T193 [P] Enhance service worker for offline menu browsing in `static/pwa/sw.js` (cache menu HTML, images, CSS)
- [x] T194 [P] Add PWA install prompt and instructions in menu template

### Performance Optimization

- [x] T195 [P] Optimize menu page load performance (verify <2s on 3G) - image lazy loading, template caching
- [ ] T196 [P] Optimize SSE event propagation (verify <5s) - efficient event bus, connection management
- [x] T197 [P] Add performance monitoring middleware in `internal/interfaces/http/middleware/performance.go` (logs slow requests >200ms)

### Security

- [x] T198 [P] Add input validation and sanitization for all form inputs (prevent XSS, injection attacks)
- [x] T199 [P] Add CSRF protection for form submissions in `internal/interfaces/http/middleware/csrf.go`
- [x] T200 [P] Add rate limiting middleware in `internal/interfaces/http/middleware/ratelimit.go` (prevent abuse)
- [ ] T201 [P] Add graceful degradation for external service failures in `internal/infrastructure/storage/photo.go` (returns error but doesn't crash, allows menu to function without photos) for FR-048 - **POST-MVP**: FR-048 marked as post-MVP in spec

### Accessibility

- [x] T202 [P] Verify WCAG 2.1 AA compliance across all templates (semantic HTML, alt text, keyboard navigation)
- [x] T203 [P] Add ARIA labels and roles to interactive elements in all templates

### Documentation

- [ ] T204 [P] Add Go doc comments to all public APIs in `internal/application/` and `internal/interfaces/http/`
- [x] T205 [P] Update README.md with setup instructions, architecture overview, API documentation

### Code Quality & Testing

- [x] T206 [P] Run full test suite and verify coverage: 80% overall, 95% for critical paths (payment, orders, kitchen)
- [ ] T207 [P] Run gocyclo to verify all functions have complexity <10
- [ ] T208 [P] Run golangci-lint and fix all issues
- [ ] T209 [P] Verify all functions <50 lines, all types/structs <300 lines (refactor if needed)

### Integration & End-to-End Testing

- [ ] T210 [P] Create end-to-end test for complete customer journey in `tests/e2e/customer_journey_test.go` (QR scan → menu → cart → order → status updates)
- [ ] T211 [P] Create end-to-end test for complete kitchen workflow in `tests/e2e/kitchen_workflow_test.go` (order appears → mark paid → preparing → ready)
- [ ] T212 [P] Create end-to-end test for owner menu setup in `tests/e2e/owner_setup_test.go` (signup → categories → items → QR code)

### Quickstart Validation

- [x] T213 [P] Run quickstart.md validation - verify all setup steps work correctly
- [x] T214 [P] Verify development workflow from quickstart.md (templ generate, go run, etc.)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Depends on US1 Order entity and events, but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories (menu setup is independent)
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Depends on US1 Order entity for analytics, but should be independently testable

### Within Each User Story

- Tests (MANDATORY) MUST be written and FAIL before implementation
- Domain entities before repositories
- Repositories before use cases
- Use cases before HTTP handlers
- HTTP handlers before templates
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Domain entities within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all domain entity tests together:
Task: "Unit tests for Restaurant entity validation in tests/unit/domain/restaurant_test.go"
Task: "Unit tests for MenuCategory entity validation in tests/unit/domain/menu_test.go"
Task: "Unit tests for MenuItem entity validation in tests/unit/domain/menu_test.go"
Task: "Unit tests for Order entity and state transitions in tests/unit/domain/order_test.go"
Task: "Unit tests for Payment entity and state transitions in tests/unit/domain/payment_test.go"

# Launch all repository tests together:
Task: "Unit tests for in-memory RestaurantRepository in tests/unit/infrastructure/repositories/memory/restaurant_test.go"
Task: "Unit tests for in-memory MenuCategoryRepository in tests/unit/infrastructure/repositories/memory/menu_test.go"
Task: "Unit tests for in-memory MenuItemRepository in tests/unit/infrastructure/repositories/memory/menu_test.go"
Task: "Unit tests for in-memory OrderRepository in tests/unit/infrastructure/repositories/memory/order_test.go"
Task: "Unit tests for in-memory PaymentRepository in tests/unit/infrastructure/repositories/memory/payment_test.go"

# Launch all contract tests together:
Task: "Contract test for GET /menu endpoint in tests/contract/http/menu_test.go"
Task: "Contract test for POST /cart/add endpoint in tests/contract/http/cart_test.go"
Task: "Contract test for POST /cart/remove endpoint in tests/contract/http/cart_test.go"
Task: "Contract test for GET /order/confirm endpoint in tests/contract/http/order_test.go"
Task: "Contract test for POST /order/create endpoint in tests/contract/http/order_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Customer ordering)
   - Developer B: User Story 2 (Kitchen fulfillment)
   - Developer C: User Story 3 (Menu setup)
   - Developer D: User Story 4 (Dashboard)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- **CRITICAL**: Tests are MANDATORY per Constitution - write tests FIRST, then implement
- **CRITICAL**: Code quality must be verified: functions <50 lines, types/structs <300 lines, complexity <10
- **CRITICAL**: Test coverage must meet minimums: 80% standard, 95% critical paths

---

## Summary

- **Total Tasks**: 214 (updated from 199 - added 15 remediation tasks)
- **Tasks per User Story**:
  - User Story 1 (P1): 44 tasks (MVP)
  - User Story 2 (P2): 36 tasks (added offline sync tasks)
  - User Story 3 (P3): 42 tasks (added cache invalidation and photo limit tasks)
  - User Story 4 (P4): 17 tasks (added weekly summary support)
  - Setup & Foundational: 42 tasks (added logging infrastructure)
  - Polish: 34 tasks (added order archiving, external service failure handling)
- **Remediation Tasks Added**:
  - T036-T038: Logging infrastructure (FR-046)
  - T098, T108-T109: Kitchen offline sync (FR-017)
  - T112: Audible alert enhancement (FR-012)
  - T136, T153-T154: Menu cache invalidation (FR-024)
  - T126, T132, T135, T141, T145: Photo limit validation (FR-049)
  - T189-T192: Order archiving (FR-016)
  - T201: External service failure handling (FR-048)
  - T168, T171-T172, T177: Weekly summary support (US4)
- **Parallel Opportunities**: Many tasks marked [P] can run in parallel within each phase
- **Independent Test Criteria**: Each user story has clear independent test criteria
- **Suggested MVP Scope**: User Story 1 only (Customer Orders Food with Cash Payment)
- **Format Validation**: All tasks follow checklist format (checkbox, ID, labels, file paths)

