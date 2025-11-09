# Implementation Plan: Cash Payment with Hypermedia UI

**Branch**: `002-cash-payment-hypermedia` | **Date**: 2025-01-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-cash-payment-hypermedia/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

BitMerchant v1.0 enables restaurants to accept cash payments with a zero-friction customer experience. Customers scan QR codes, browse menus, order food, and confirm cash payment - all without creating accounts. The system provides real-time order tracking, kitchen display management, and basic analytics for restaurant owners. Architecture is designed to support multiple payment methods (cash initially, Lightning Network in future) without requiring refactoring.

**Technical Approach**: Clean Architecture with Domain-Driven Design (DDD) implemented in Go 1.25+ using Echo web framework. Hypermedia-driven UI with server-rendered HTML templates using Templ and Datastar for partial page updates and real-time SSE updates. Payment method abstraction layer supports cash payments now and Lightning payments in future. Initial implementation uses in-memory repositories with PostgreSQL-ready interfaces for future migration. Progressive Web App (PWA) frontend.

## Technical Context

**Language/Version**: Go 1.25+ (Go 1.21+ minimum for required features)  
**Primary Dependencies**: 
  - `github.com/labstack/echo/v4` - Web framework with HTTP/2 SSE support
  - `github.com/a-h/templ` - Type-safe Go templates for server-rendered HTML
  - `github.com/delaneyj/datastar` - Hypermedia-driven UI with SSE for partial page updates and real-time DOM updates (REQUIRED - not optional)
  - `github.com/ThreeDotsLabs/watermill` - Event streaming and pub/sub for domain events
  - Templ UI components from `templui.io` - Pre-built UI components for Templ templates
**Storage**: In-memory repositories (v1.0) with PostgreSQL-ready interface design. Future: PostgreSQL with hand-written SQL (no ORM).  
**Testing**: Go standard `testing` package, `testify` for assertions. Integration tests for SSE, contract tests for HTTP endpoints.  
**Target Platform**: Linux server (backend), Progressive Web App (PWA) for frontend - works on any device with modern browser  
**Project Type**: Web application (single codebase, server-rendered HTML with SSE for real-time updates)  
**Performance Goals**: 
  - API endpoints: <200ms p95 latency (Constitution requirement)
  - Menu page load: <2 seconds on 3G (SC-003)
  - Cash payment confirmation: immediate (no external processing delay) (SC-002)
  - Order status updates: <5 seconds propagation (SC-004, SC-005)
  - Critical user flow (order → confirm cash payment): <2 minutes total (SC-001)  
**Constraints**: 
  - Functions max 50 lines, classes max 300 lines (Constitution)
  - Cyclomatic complexity max 10 per function (Constitution)
  - Minimum 80% test coverage, 95% for payment/critical paths (Constitution)
  - Frontend bundle <200KB gzipped (Constitution)
  - PWA offline support for menu browsing
  - Single tenant per deployment (v1.0)
  - Payment method abstraction must support future Lightning integration without refactoring (FR-031)
**Scale/Scope**: 
  - 10 restaurants active daily (SC-010)
  - 1,000+ successful orders processed (SC-011)
  - 20+ orders per restaurant per day (SC-012)
  - Single restaurant per deployment initially

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Assessment

**Code Quality**: 
- ✅ Architecture supports small functions: Clean Architecture with DDD enforces separation of concerns, domain logic isolated in small functions
- ⚠️ **Action Required**: Implement code review checklist to verify functions <50 lines, classes <300 lines. Use `gocyclo` linter to enforce complexity <10.
- ✅ Standard library first approach aligns with code quality principles
- ✅ Payment method abstraction pattern keeps payment logic isolated and testable

**Testing Standards**: 
- ✅ **Coverage Requirements**: 
  - Payment processing (cash confirmation, payment method abstraction): 95% coverage (critical path)
  - Order management, kitchen display: 95% coverage (critical path)
  - Menu management, analytics: 80% coverage (standard)
- ✅ **Test Types Required**:
  - Unit tests: All domain logic, use cases, repositories, payment method abstraction
  - Integration tests: SSE event streaming, photo upload/optimization, hypermedia interactions
  - Contract tests: All HTTP endpoints returning HTML (menu, orders, payments, kitchen display)
  - Performance tests: Order flow (<2min), menu load (<2s), order status propagation (<5s)
- ✅ TDD workflow: Write tests → Get approval → Tests fail → Implement → Tests pass

**User Experience Consistency**: 
- ✅ **Design Patterns**: Type-safe Templ templates ensure consistent UI components across customer menu, kitchen display, owner dashboard
- ✅ **UI Components**: Use templui.io components for consistent design system
- ✅ **Hypermedia Interactions**: Datastar enables form submissions, links, and SSE updates without JavaScript frameworks (FR-041, FR-042, FR-043)
- ✅ **Error Messages**: User-friendly error handling for validation errors, network issues
- ✅ **Accessibility**: PWA with semantic HTML, WCAG 2.1 AA compliance (Constitution requirement)
- ✅ **Loading States**: SSE provides real-time updates without page refresh, loading indicators for operations >200ms
- ✅ **Responsive Design**: Mobile-first PWA works across all device sizes

**Performance Requirements**: 
- ✅ **API Endpoints**: <200ms p95 latency (Constitution) - Go + Echo provides low-latency HTTP handling
- ✅ **Critical Flows**: 
  - Order → Cash confirmation flow: <2 minutes total (SC-001) - achievable with optimized menu load + immediate cash confirmation
  - Cash payment confirmation: immediate (SC-002) - no external payment processing delay
- ✅ **Page Load**: <2 seconds on 3G (SC-003) - Server-rendered templates + optimized images (300KB)
- ✅ **Performance Tests**: Required for order flow, menu load, order status updates
- ✅ **Bundle Size**: Server-rendered templates minimize frontend bundle, target <200KB gzipped

**Quality Gates**: 
- ✅ **Linting**: `golangci-lint` with strict rules, `gocyclo` for complexity
- ✅ **Tests**: Go test suite with coverage reporting, minimum thresholds enforced
- ✅ **Build**: Go build with CI/CD pipeline
- ✅ **Performance**: Automated performance tests in CI/CD
- ✅ **Security**: Go security scanning, payment method abstraction security review
- ✅ **Documentation**: Go doc comments for all public APIs, HTML contract documentation

### Post-Design Re-Assessment

**Status**: ✅ All gates pass after Phase 1 design

**Code Quality**: 
- ✅ Data model enforces small, focused entities (Restaurant, MenuCategory, MenuItem, Order, PaymentMethod)
- ✅ Payment method abstraction keeps payment logic isolated and extensible
- ✅ Repository interfaces support clean separation (domain defines interfaces, infrastructure implements)
- ✅ Domain events keep handlers small and focused
- ✅ Architecture supports functions <50 lines, classes <300 lines

**Testing Standards**: 
- ✅ Data model provides clear test boundaries (entities, repositories, use cases, payment methods)
- ✅ Contract tests defined for all HTML-returning HTTP endpoints
- ✅ Integration test points identified (SSE, photo storage, hypermedia interactions)
- ✅ Domain events enable event-driven testing
- ✅ Payment method abstraction enables testing cash and future Lightning independently

**User Experience Consistency**: 
- ✅ HTML contracts define consistent page structure and hypermedia interactions
- ✅ SSE endpoints provide real-time updates without page refresh
- ✅ PWA structure supports offline menu browsing
- ✅ Responsive design supported via server-rendered templates
- ✅ templui.io components ensure consistent UI design

**Performance Requirements**: 
- ✅ HTML contracts specify performance targets (<200ms, <2s, <5s)
- ✅ SSE endpoints use HTTP/2 for efficient real-time updates
- ✅ Photo optimization (300KB) supports fast page loads
- ✅ In-memory repositories provide low-latency for v1.0
- ✅ Cash payment confirmation has zero external processing delay

**Quality Gates**: 
- ✅ All gates pass: Architecture supports linting, testing, build, performance, security, documentation requirements

## Project Structure

### Documentation (this feature)

```text
specs/002-cash-payment-hypermedia/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── domain/              # Pure business logic (zero dependencies)
│   ├── restaurant.go    # Restaurant entity, value objects
│   ├── menu.go          # Menu category, menu item entities
│   ├── order.go         # Order entity, order status, order items
│   ├── payment.go       # Payment entity, payment status, payment method abstraction
│   └── events.go        # Domain events (OrderPaid, OrderReady, etc.)
├── application/         # Use cases, commands, queries
│   ├── menu/            # Menu management use cases
│   ├── order/           # Order processing use cases
│   ├── payment/         # Payment processing use cases (cash, future: lightning)
│   │   ├── cash/        # Cash payment use cases
│   │   └── method.go    # Payment method abstraction interface
│   └── kitchen/         # Kitchen display use cases
├── infrastructure/      # External dependencies
│   ├── repositories/    # In-memory (v1.0) and PostgreSQL (future) implementations
│   │   ├── memory/      # In-memory repositories with sync.RWMutex
│   │   └── postgres/    # PostgreSQL repositories (future, same interfaces)
│   ├── payment/         # Payment method implementations
│   │   └── cash/        # Cash payment implementation
│   │   # Future: lightning/ directory for Lightning payment implementation
│   ├── storage/         # Photo storage (S3 or similar)
│   └── events/          # Watermill event bus implementation
└── interfaces/          # External interfaces
    ├── http/            # Echo HTTP handlers (return HTML, not JSON)
    │   ├── menu.go      # Menu HTML endpoints
    │   ├── order.go     # Order HTML endpoints
    │   ├── payment.go   # Payment HTML endpoints
    │   ├── kitchen.go   # Kitchen display HTML endpoints
    │   └── sse.go       # Server-Sent Events handler for real-time updates
    └── templates/       # Templ templates
        ├── components/ # Reusable UI components (using templui.io patterns)
        ├── menu.templ   # Customer menu view
        ├── kitchen.templ # Kitchen display view
        └── dashboard.templ # Owner dashboard view

cmd/
└── server/
    └── main.go          # Application entry point, dependency injection

tests/
├── contract/            # Contract tests for HTML endpoints
├── integration/         # Integration tests (SSE, storage, hypermedia)
└── unit/                # Unit tests for domain, application, infrastructure

static/                  # Static assets (CSS, JS, images)
└── pwa/                 # PWA manifest, service worker
```

**Structure Decision**: Single web application using Clean Architecture with DDD. Server-rendered HTML via Templ templates with Datastar for hypermedia-driven UI (partial page updates, form submissions, SSE real-time updates). Payment method abstraction layer in `internal/infrastructure/payment/` supports cash payments now and Lightning payments in future without refactoring. Hypermedia-driven UI using Templ templates with Datastar attributes (ds-post, ds-target, ds-sse-connect) and templui.io components for consistent design. No separate frontend build step - templates compiled to Go code. PWA capabilities added via service worker and manifest. Repository pattern allows swapping in-memory (v1.0) for PostgreSQL (future) without changing application code.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Payment method abstraction layer | Supports future Lightning integration without refactoring (FR-031) | Direct cash payment implementation would require refactoring when adding Lightning, violating extensibility requirement |
