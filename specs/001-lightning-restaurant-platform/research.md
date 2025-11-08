# Research: BitMerchant v1.0 Technical Architecture

**Date**: 2025-11-08  
**Feature**: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

## Architecture Decisions

### Decision: Clean Architecture with Domain-Driven Design (DDD)

**Rationale**: 
- Enforces separation of concerns: domain logic isolated from infrastructure
- Enables testability: domain layer has zero dependencies, easily unit testable
- Supports future migration: swap in-memory repositories for PostgreSQL without changing application code
- Aligns with Constitution code quality: small, focused functions in domain layer
- Industry standard pattern for Go web applications

**Alternatives Considered**:
- **MVC (Model-View-Controller)**: Rejected - too tightly coupled, harder to test domain logic in isolation
- **Hexagonal Architecture**: Considered - similar benefits but Clean Architecture provides clearer layer boundaries
- **Monolithic with direct DB access**: Rejected - violates Constitution testing standards, harder to achieve 95% coverage on critical paths

**Implementation Pattern**:
```
internal/
├── domain/          # Pure business logic, zero dependencies
├── application/     # Use cases orchestrate domain
├── infrastructure/ # External adapters (repos, APIs, events)
└── interfaces/      # HTTP handlers, templates
```

**Reference**: "Clean Architecture" by Robert C. Martin, adapted for Go by Three Dots Labs patterns

---

### Decision: Go 1.25+ with Echo Web Framework

**Rationale**:
- Go provides excellent performance: low-latency HTTP handling meets <200ms p95 requirement
- Echo framework: lightweight, HTTP/2 support for SSE, middleware ecosystem
- Standard library first: aligns with Constitution principle of minimal dependencies
- Type safety: compile-time checks reduce runtime errors
- Concurrency: goroutines handle concurrent orders efficiently (FR-034)

**Alternatives Considered**:
- **Gin**: Similar to Echo, but Echo has better HTTP/2 SSE documentation
- **Fiber**: Faster but less standard library alignment
- **net/http standard library**: Rejected - too low-level, would require building too much framework code

**Key Features Used**:
- HTTP/2 Server-Sent Events (SSE) for real-time updates
- Middleware for authentication, logging, error handling
- Context propagation for request-scoped data

**Reference**: Echo documentation, Go HTTP/2 SSE examples

---

### Decision: Templ for Type-Safe Templates

**Rationale**:
- Type safety: templates compile to Go code, catch errors at compile time
- Performance: compiled templates faster than runtime template parsing
- Consistency: ensures UI components are consistent across customer menu, kitchen display, dashboard (UX Constitution requirement)
- No separate build step: templates compiled with Go code, simpler deployment

**Alternatives Considered**:
- **html/template (standard library)**: Rejected - runtime errors, no type safety
- **React/Next.js frontend**: Rejected - adds complexity, separate build step, violates "standard library first" principle
- **HTMX**: Considered but Templ provides better type safety for Go

**Implementation Pattern**:
- Templates in `internal/interfaces/templates/`
- Compile to Go code: `templ generate`
- Render in HTTP handlers: `templ.Handler()`

**Reference**: https://templ.guide/

---

### Decision: Watermill Event Bus + Datastar for Real-Time Updates

**Rationale**:
- **Watermill**: Provides event streaming infrastructure, supports in-memory pub/sub (v1.0) and can migrate to Kafka/RabbitMQ (future)
- **Datastar**: Hypermedia-driven real-time updates via SSE, aligns with server-rendered architecture
- **SSE over WebSockets**: Simpler for server-to-client push, HTTP/2 native support, no connection management complexity
- **Event-driven architecture**: Domain events (OrderPaid, OrderReady) trigger real-time updates without tight coupling

**Alternatives Considered**:
- **WebSockets**: Rejected - more complex connection management, not needed for unidirectional server-to-client updates
- **Polling**: Rejected - violates performance requirements (<5s update propagation)
- **HTMX + SSE**: Considered but Datastar provides better abstraction for hypermedia updates

**Implementation Pattern**:
1. Domain event occurs (e.g., `OrderPaid`)
2. Watermill publishes to in-memory event bus
3. SSE handler streams to Datastar clients via Echo HTTP/2
4. Datastar updates DOM automatically (no JavaScript needed)

**Reference**: 
- https://threedots.tech/post/live-website-updates-go-sse-htmx/ (adapt HTMX to Datastar)
- Watermill documentation: https://watermill.io/
- Datastar: https://github.com/delaneyj/datastar

---

### Decision: In-Memory Repositories (v1.0) with PostgreSQL-Ready Interfaces

**Rationale**:
- **v1.0 simplicity**: No database setup required, faster development, easier testing
- **PostgreSQL-ready**: Repository interfaces allow swapping implementations without changing application code
- **Migration path**: When scale requires persistence, swap `memory.Repository` for `postgres.Repository` in `main.go`
- **Testing**: In-memory repos make unit tests fast and isolated (Constitution requirement: tests complete in <5 minutes)

**Alternatives Considered**:
- **PostgreSQL from start**: Rejected - adds complexity, violates SLC "Simple" principle for v1.0
- **SQLite**: Considered but PostgreSQL-ready interfaces allow better future migration
- **No repository pattern**: Rejected - violates Clean Architecture, makes testing harder

**Implementation Pattern**:
```go
// Domain defines interface
type OrderRepository interface {
    Save(order *Order) error
    FindByID(id OrderID) (*Order, error)
}

// Infrastructure implements
type MemoryOrderRepository struct { ... }
type PostgresOrderRepository struct { ... } // Future

// Application uses interface, doesn't know implementation
```

**Reference**: Repository pattern from "Domain-Driven Design" by Eric Evans

---

### Decision: Strike API for Lightning Network Payments

**Rationale**:
- **Strike API**: Production-ready Lightning Network payment processing
- **Simplifies integration**: Handles Lightning Network complexity (invoices, routing, settlement)
- **Exchange rates**: Provides fiat-to-satoshi conversion (FR-033)
- **Settlement**: Handles daily settlement to restaurant Lightning address (FR-031)
- **Error handling**: API provides clear error responses for payment failures (FR-030)

**Alternatives Considered**:
- **Direct Lightning Node**: Rejected - too complex, requires Lightning node management, violates "Simple" principle
- **Other Lightning APIs**: Strike chosen for production readiness and documentation
- **Manual Lightning integration**: Rejected - would require significant Lightning Network expertise

**Integration Points**:
- Invoice generation: Create Lightning invoice for customer payment
- Payment polling: Check invoice status every 2-3 seconds (detect completion within 10s per SC-002)
- Settlement: Daily batch settlement to restaurant address

**Reference**: Strike API documentation (to be reviewed during implementation)

---

### Decision: Server-Rendered PWA (No Separate Frontend Build)

**Rationale**:
- **Simplicity**: Single codebase, no separate frontend build step, simpler deployment
- **Performance**: Server-rendered HTML loads faster, meets <2s page load requirement (SC-003)
- **Bundle size**: Minimal JavaScript (only Datastar client), meets <200KB gzipped requirement
- **PWA capabilities**: Service worker for offline menu browsing, installable to home screen
- **Real-time**: SSE provides updates without JavaScript frameworks

**Alternatives Considered**:
- **React/Next.js frontend**: Rejected - adds complexity, larger bundle size, separate build step
- **Vue/Nuxt**: Rejected - same issues as React
- **Pure HTML + vanilla JS**: Considered but Datastar provides better abstraction

**PWA Implementation**:
- Service worker: Cache menu HTML/CSS for offline browsing
- Web App Manifest: Installable to home screen
- HTTPS required: PWA requirement (handled by deployment)

**Reference**: PWA best practices, Service Worker API

---

### Decision: Photo Storage (S3 or Similar) with Optimization

**Rationale**:
- **Cloud storage**: Scalable, CDN delivery for fast image loading
- **Optimization**: Automatic compression to 300KB (FR-020) for display version
- **Limits**: 2MB upload max, 100 photos per restaurant (FR-020, FR-042)
- **CDN**: Fast global delivery meets performance requirements

**Alternatives Considered**:
- **Local file storage**: Rejected - doesn't scale, no CDN benefits
- **Database BLOBs**: Rejected - inefficient, violates separation of concerns
- **Multiple storage providers**: Rejected - keep simple for v1.0, can add later

**Implementation**:
- Upload handler validates 2MB limit
- Image processing: compress to 300KB optimized version
- Store original (for future editing) and optimized (for display)
- CDN URL returned for menu display

**Reference**: AWS S3 best practices, image optimization libraries (to be selected)

---

## Technology Stack Summary

| Component | Technology | Rationale |
|-----------|-----------|------------|
| Language | Go 1.25+ | Performance, standard library first, concurrency |
| Web Framework | Echo v4 | HTTP/2 SSE support, lightweight, middleware |
| Templates | Templ | Type-safe, compile-time checks, consistency |
| Real-time | Watermill + Datastar + SSE | Event-driven, hypermedia updates, simple |
| Storage | In-memory (v1.0) | Simple, fast, PostgreSQL-ready interfaces |
| Payments | Strike API | Production-ready Lightning Network integration |
| Photo Storage | S3 or similar | Scalable, CDN delivery, optimization |
| Testing | Go testing + testify | Standard library, assertions, coverage |

## Unresolved Questions

None - all technical decisions made based on provided architecture and requirements.

## Next Steps

1. Generate data model from entities in spec
2. Create API contracts for HTTP endpoints
3. Design domain events for real-time updates
4. Create quickstart guide for development setup

