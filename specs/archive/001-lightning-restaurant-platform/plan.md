# Implementation Plan: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

**Branch**: `001-lightning-restaurant-platform` | **Date**: 2025-11-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-lightning-restaurant-platform/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

BitMerchant v1.0 enables restaurants to accept Lightning Network payments with a zero-friction customer experience. Customers scan QR codes, browse menus, order food, and pay instantly with Bitcoin - all without creating accounts. The system provides real-time order tracking, kitchen display management, and basic analytics for restaurant owners.

**Technical Approach**: Clean Architecture with Domain-Driven Design (DDD) implemented in Go 1.25+ using Echo web framework. Real-time updates via Server-Sent Events (SSE) through Watermill event bus and Datastar. Initial implementation uses in-memory repositories with PostgreSQL-ready interfaces for future migration. Progressive Web App (PWA) frontend with type-safe Templ templates.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.25+ (Go 1.21+ minimum for required features)  
**Primary Dependencies**: 
  - `github.com/labstack/echo/v4` - Web framework with HTTP/2 SSE support
  - `github.com/a-h/templ` - Type-safe Go templates
  - `github.com/delaneyj/datastar` - Real-time hypermedia updates via SSE
  - `github.com/ThreeDotsLabs/watermill` - Event streaming and pub/sub
  - Strike API - Lightning Network payment processing  
**Storage**: In-memory repositories (v1.0) with PostgreSQL-ready interface design. Future: PostgreSQL with hand-written SQL (no ORM).  
**Testing**: Go standard `testing` package, `testify` for assertions. Integration tests for Strike API, contract tests for HTTP endpoints.  
**Target Platform**: Linux server (backend), Progressive Web App (PWA) for frontend - works on any device with modern browser  
**Project Type**: Web application (single codebase, server-rendered with SSE for real-time updates)  
**Performance Goals**: 
  - API endpoints: <200ms p95 latency (Constitution requirement)
  - Menu page load: <2 seconds on 3G (SC-003)
  - Lightning payment: <10 seconds end-to-end (SC-002)
  - Order status updates: <5 seconds propagation (SC-004, SC-005)
  - Critical user flow (order → pay): <2 minutes total (SC-001)  
**Constraints**: 
  - Functions max 50 lines, classes max 300 lines (Constitution)
  - Cyclomatic complexity max 10 per function (Constitution)
  - Minimum 80% test coverage, 95% for payment/critical paths (Constitution)
  - Frontend bundle <200KB gzipped (Constitution)
  - PWA offline support for menu browsing
  - Single tenant per deployment (v1.0)  
**Scale/Scope**: 
  - 10 restaurants active daily (SC-009)
  - 1,000+ successful orders processed (SC-010)
  - 20+ orders per restaurant per day (SC-011)
  - Single restaurant per deployment initially

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Assessment

**Code Quality**: 
- ✅ Architecture supports small functions: Clean Architecture with DDD enforces separation of concerns, domain logic isolated in small functions
- ⚠️ **Action Required**: Implement code review checklist to verify functions <50 lines, classes <300 lines. Use `gocyclo` linter to enforce complexity <10.
- ✅ Standard library first approach aligns with code quality principles

**Testing Standards**: 
- ✅ **Coverage Requirements**: 
  - Payment processing (Strike API integration): 95% coverage (critical path)
  - Order management, kitchen display: 95% coverage (critical path)
  - Menu management, analytics: 80% coverage (standard)
- ✅ **Test Types Required**:
  - Unit tests: All domain logic, use cases, repositories
  - Integration tests: Strike API client, SSE event streaming, photo upload/optimization
  - Contract tests: All HTTP endpoints (menu, orders, payments, kitchen display)
  - Performance tests: Payment flow (<10s), menu load (<2s), order status propagation (<5s)
- ✅ TDD workflow: Write tests → Get approval → Tests fail → Implement → Tests pass

**User Experience Consistency**: 
- ✅ **Design Patterns**: Type-safe Templ templates ensure consistent UI components across customer menu, kitchen display, owner dashboard
- ✅ **Error Messages**: User-friendly error handling for payment failures, network issues, validation errors (FR-030)
- ✅ **Accessibility**: PWA with semantic HTML, WCAG 2.1 AA compliance (Constitution requirement)
- ✅ **Loading States**: SSE provides real-time updates without page refresh, loading indicators for operations >200ms
- ✅ **Responsive Design**: Mobile-first PWA works across all device sizes

**Performance Requirements**: 
- ✅ **API Endpoints**: <200ms p95 latency (Constitution) - Go + Echo provides low-latency HTTP handling
- ✅ **Critical Flows**: 
  - Order → Pay flow: <2 minutes total (SC-001) - achievable with optimized menu load + Lightning payment
  - Payment completion: <10 seconds (SC-002) - Lightning Network + Strike API
- ✅ **Page Load**: <2 seconds on 3G (SC-003) - Server-rendered templates + optimized images (300KB)
- ✅ **Performance Tests**: Required for payment flow, menu load, order status updates
- ✅ **Bundle Size**: Server-rendered templates minimize frontend bundle, target <200KB gzipped

**Quality Gates**: 
- ✅ **Linting**: `golangci-lint` with strict rules, `gocyclo` for complexity
- ✅ **Tests**: Go test suite with coverage reporting, minimum thresholds enforced
- ✅ **Build**: Go build with CI/CD pipeline
- ✅ **Performance**: Automated performance tests in CI/CD
- ✅ **Security**: Go security scanning, Strike API security review
- ✅ **Documentation**: Go doc comments for all public APIs, OpenAPI spec for contracts

### Post-Design Re-Assessment

**Status**: ✅ All gates pass after Phase 1 design

**Code Quality**: 
- ✅ Data model enforces small, focused entities (Restaurant, MenuCategory, MenuItem, Order, Payment)
- ✅ Repository interfaces support clean separation (domain defines interfaces, infrastructure implements)
- ✅ Domain events keep handlers small and focused
- ✅ Architecture supports functions <50 lines, classes <300 lines

**Testing Standards**: 
- ✅ Data model provides clear test boundaries (entities, repositories, use cases)
- ✅ Contract tests defined for all HTTP endpoints
- ✅ Integration test points identified (Strike API, SSE, photo storage)
- ✅ Domain events enable event-driven testing

**User Experience Consistency**: 
- ✅ API contracts define consistent error responses
- ✅ SSE endpoints provide real-time updates without page refresh
- ✅ PWA structure supports offline menu browsing
- ✅ Responsive design supported via server-rendered templates

**Performance Requirements**: 
- ✅ API contracts specify performance targets (<200ms, <2s, <10s, <5s)
- ✅ SSE endpoints use HTTP/2 for efficient real-time updates
- ✅ Photo optimization (300KB) supports fast page loads
- ✅ In-memory repositories provide low-latency for v1.0

**Quality Gates**: 
- ✅ All gates pass: Architecture supports linting, testing, build, performance, security, documentation requirements

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
internal/
├── domain/              # Pure business logic (zero dependencies)
│   ├── restaurant.go    # Restaurant entity, value objects
│   ├── menu.go          # Menu category, menu item entities
│   ├── order.go         # Order entity, order status, order items
│   ├── payment.go       # Payment entity, payment status
│   └── events.go        # Domain events (OrderPaid, OrderReady, etc.)
├── application/         # Use cases, commands, queries
│   ├── menu/            # Menu management use cases
│   ├── order/           # Order processing use cases
│   ├── payment/         # Payment processing use cases
│   └── kitchen/         # Kitchen display use cases
├── infrastructure/      # External dependencies
│   ├── repositories/   # In-memory (v1.0) and PostgreSQL (future) implementations
│   │   ├── memory/      # In-memory repositories with sync.RWMutex
│   │   └── postgres/    # PostgreSQL repositories (future, same interfaces)
│   ├── strike/          # Strike API client for Lightning payments
│   ├── storage/         # Photo storage (S3 or similar)
│   └── events/          # Watermill event bus implementation
└── interfaces/          # External interfaces
    ├── http/            # Echo HTTP handlers
    │   ├── menu.go      # Menu endpoints
    │   ├── order.go     # Order endpoints
    │   ├── payment.go   # Payment endpoints
    │   ├── kitchen.go   # Kitchen display endpoints
    │   └── sse.go       # Server-Sent Events handler for real-time updates
    └── templates/       # Templ templates
        ├── menu.templ   # Customer menu view
        ├── kitchen.templ # Kitchen display view
        └── dashboard.templ # Owner dashboard view

cmd/
└── server/
    └── main.go          # Application entry point, dependency injection

tests/
├── contract/            # Contract tests for HTTP endpoints
├── integration/         # Integration tests (Strike API, SSE, storage)
└── unit/                # Unit tests for domain, application, infrastructure

static/                  # Static assets (CSS, JS, images)
└── pwa/                 # PWA manifest, service worker
```

**Structure Decision**: Single web application using Clean Architecture with DDD. Server-rendered HTML via Templ templates with SSE for real-time updates. No separate frontend build step - templates compiled to Go code. PWA capabilities added via service worker and manifest. Repository pattern allows swapping in-memory (v1.0) for PostgreSQL (future) without changing application code.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
