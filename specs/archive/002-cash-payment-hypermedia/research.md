# Research: Cash Payment with Hypermedia UI Technical Architecture

**Date**: 2025-01-27  
**Feature**: Cash Payment with Hypermedia UI

## Architecture Decisions

### Decision: Clean Architecture with Domain-Driven Design (DDD)

**Rationale**: 
- Enforces separation of concerns: domain logic isolated from infrastructure
- Enables testability: domain layer has zero dependencies, easily unit testable
- Supports future migration: swap in-memory repositories for PostgreSQL without changing application code
- Aligns with Constitution code quality: small, focused functions in domain layer
- Payment method abstraction supports future Lightning integration without refactoring (FR-031)
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
├── infrastructure/ # External adapters (repos, APIs, events, payment methods)
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
- Concurrency: goroutines handle concurrent orders efficiently (FR-044)
- HTML-returning handlers: Echo supports rendering HTML templates directly (FR-038, FR-039, FR-040)

**Alternatives Considered**:
- **Gin**: Similar to Echo, but Echo has better HTTP/2 SSE documentation
- **Fiber**: Faster but less standard library alignment
- **net/http standard library**: Rejected - too low-level, would require building too much framework code

**Key Features Used**:
- HTTP/2 Server-Sent Events (SSE) for real-time updates
- Middleware for authentication, logging, error handling
- Context propagation for request-scoped data
- HTML template rendering support

**Reference**: Echo documentation, Go HTTP/2 SSE examples

---

### Decision: Templ for Type-Safe Server-Rendered HTML Templates

**Rationale**:
- Type safety: templates compile to Go code, catch errors at compile time
- Performance: compiled templates faster than runtime template parsing
- Consistency: ensures UI components are consistent across customer menu, kitchen display, dashboard (UX Constitution requirement)
- No separate build step: templates compiled with Go code, simpler deployment
- Hypermedia support: Templ supports form submissions, links, and hypermedia-driven interactions (FR-041)
- Server-side rendering: Returns HTML pages instead of JSON (FR-038, FR-039, FR-040)
- Streaming support: Templ supports HTTP streaming for improved TTFB (Time to First Byte)
- Declarative Shadow DOM: Templ supports Suspense patterns for progressive loading

**Alternatives Considered**:
- **html/template (standard library)**: Rejected - runtime errors, no type safety
- **React/Next.js frontend**: Rejected - adds complexity, separate build step, violates "standard library first" principle, requires JSON API
- **HTMX**: Rejected - Datastar provides better integration with Go/Templ and SSE for hypermedia-driven UI

**Implementation Pattern**:
- Templates in `internal/interfaces/templates/`
- Compile to Go code: `templ generate`
- Render in HTTP handlers: `templ.Handler()` or `templ.Handler(component, templ.WithStreaming())` for streaming
- Use templui.io components for consistent UI design system
- Follow patterns from templ.guide/llms.md for LLM-assisted development

**References**: 
- https://templ.guide/ - Main Templ documentation
- https://templ.guide/llms.md - LLM guidance for writing Templ code (use for AI-assisted development)
- https://templui.io/ - Pre-built UI components library for Templ templates

---

### Decision: Payment Method Abstraction Architecture

**Rationale**:
- Supports multiple payment types without refactoring (FR-031)
- Cash payments implemented first (FR-032)
- Lightning Network payments can be added later as new payment method type (FR-037)
- Isolates payment logic for easier testing (95% coverage requirement)
- Enables independent development and testing of payment methods

**Implementation Pattern**:
```go
// Domain interface
type PaymentMethod interface {
    ProcessPayment(ctx context.Context, order Order) error
    ValidatePayment(ctx context.Context, paymentID string) error
}

// Cash implementation
type CashPaymentMethod struct { ... }

// Future: Lightning implementation
type LightningPaymentMethod struct { ... }
```

**Alternatives Considered**:
- **Direct cash implementation**: Rejected - would require refactoring when adding Lightning, violates extensibility requirement
- **Payment gateway abstraction**: Considered but over-engineered for v1.0 - simple interface sufficient

**Reference**: Strategy pattern, Clean Architecture payment processing patterns

---

### Decision: Hypermedia-Driven UI with Server-Sent Events (SSE)

**Rationale**:
- **Hypermedia interactions**: Form submissions, links, and real-time updates without requiring JavaScript frameworks (FR-041, FR-043)
- **SSE for real-time updates**: Server-Sent Events provide efficient server-to-client push for order status changes (FR-042)
- **No JavaScript frameworks**: Aligns with Constitution "standard library first" principle, reduces bundle size (<200KB gzipped)
- **Progressive enhancement**: Works without JavaScript, enhanced with SSE for real-time updates
- **Better UX**: No full page reloads, maintains scroll position, reduces screen flicker

**Alternatives Considered**:
- **WebSockets**: Rejected - more complex connection management, not needed for unidirectional server-to-client updates
- **Polling**: Rejected - violates performance requirements (<5s update propagation), inefficient
- **HTMX**: Rejected - Datastar provides better integration with Go/Templ and native SSE support for hypermedia-driven UI
- **JavaScript frameworks (React/Vue)**: Rejected - adds complexity, separate build step, violates "standard library first" principle

**Implementation Pattern**:
1. HTML pages rendered via Templ templates (FR-038, FR-039, FR-040)
2. Form submissions handled via Datastar attributes (ds-post, ds-target) for partial page updates (FR-041)
3. Real-time updates via Datastar SSE integration (ds-sse-connect) (FR-042)
4. Page updates without full reload using Datastar DOM updates (FR-043)
5. Hypermedia links for navigation

**Reference**: 
- Datastar: https://github.com/delaneyj/datastar - Hypermedia-driven UI with SSE for Go/Templ
- Server-Sent Events specification: https://html.spec.whatwg.org/multipage/server-sent-events.html
- Hypermedia-driven UI patterns: RESTful Web APIs by Mike Amundsen

---

### Decision: Watermill Event Bus for Domain Events

**Rationale**:
- **Event-driven architecture**: Domain events (OrderPaid, OrderReady) trigger real-time updates without tight coupling
- **In-memory pub/sub**: Supports v1.0 requirements, can migrate to Kafka/RabbitMQ (future)
- **SSE integration**: Events trigger SSE updates to connected clients
- **Decoupling**: Kitchen display, customer updates, analytics can subscribe independently

**Alternatives Considered**:
- **Direct function calls**: Rejected - tight coupling, harder to test, violates separation of concerns
- **Channels**: Considered but Watermill provides better abstraction and future scalability

**Implementation Pattern**:
1. Domain event occurs (e.g., `OrderPaid`)
2. Watermill publishes to in-memory event bus
3. SSE handler streams to Datastar clients via Echo HTTP/2
4. Datastar receives update via SSE, updates DOM automatically (no JavaScript needed)

**Reference**: 
- Datastar: https://github.com/delaneyj/datastar - Hypermedia-driven UI with SSE
- Watermill documentation: https://watermill.io/

---

### Decision: Templ UI Components from templui.io

**Rationale**:
- **Consistent design system**: Pre-built components ensure UI consistency across customer menu, kitchen display, dashboard
- **Type-safe components**: Components built with Templ provide compile-time type safety
- **Accessibility**: Components follow WCAG 2.1 AA standards (Constitution requirement)
- **Responsive design**: Components work across all device sizes (mobile-first)
- **Reduced development time**: Pre-built components accelerate development

**Implementation Pattern**:
- Import templui.io components into Templ templates
- Compose components to build pages (menu, kitchen display, dashboard)
- Customize components as needed while maintaining design consistency
- Follow templ.guide/llms.md patterns for component development

**Reference**: 
- https://templui.io/ - Templ UI component library
- https://templ.guide/llms.md - LLM guidance for Templ development

---

### Decision: In-Memory Repositories (v1.0) with PostgreSQL-Ready Interfaces

**Rationale**:
- **Fast development**: In-memory repositories enable rapid development and testing
- **PostgreSQL-ready**: Repository interfaces designed for future PostgreSQL migration without changing application code
- **Testability**: In-memory repositories simplify unit testing
- **Performance**: Low-latency for v1.0 scale (10 restaurants, 1,000+ orders)

**Alternatives Considered**:
- **PostgreSQL from start**: Rejected - over-engineering for v1.0, adds complexity and deployment requirements
- **SQLite**: Considered but PostgreSQL-ready interfaces provide better migration path

**Implementation Pattern**:
- Domain defines repository interfaces
- Infrastructure implements in-memory repositories (v1.0)
- Future: Infrastructure implements PostgreSQL repositories (same interfaces)
- Application layer uses interfaces, unaware of implementation

**Reference**: Repository pattern, Clean Architecture data access patterns

---

### Decision: Progressive Web App (PWA) Architecture

**Rationale**:
- **Offline support**: Menu browsing works offline (FR-009)
- **Installable**: Customers can install to home screen (SC-016)
- **No app store**: Reduces deployment friction
- **Cross-platform**: Works on iOS, Android, desktop browsers

**Implementation Pattern**:
- PWA manifest: `static/pwa/manifest.json`
- Service worker: `static/pwa/sw.js` for offline caching
- HTTPS required for PWA features
- App icons and splash screens

**Reference**: PWA documentation, Service Worker API

---

## Technology Stack Summary

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.25+ | Performance, concurrency, standard library first |
| Web Framework | Echo v4 | HTTP/2 SSE support, middleware, HTML rendering |
| Templates | Templ | Type-safe, compiled, server-rendered HTML |
| UI Components | templui.io | Consistent design system, pre-built components |
| Hypermedia UI | Datastar | Hypermedia-driven UI with SSE, form submissions, partial updates |
| Real-time Updates | Server-Sent Events (SSE) via Datastar | Efficient server-to-client push, HTTP/2 native, automatic DOM updates |
| Event Bus | Watermill | Domain events, in-memory pub/sub (v1.0) |
| Storage | In-memory (v1.0) | Fast development, PostgreSQL-ready interfaces |
| Frontend | Server-rendered HTML + Datastar | No JavaScript frameworks, hypermedia-driven |
| PWA | Service Worker + Manifest | Offline support, installable |

## Key References

- **Templ Documentation**: https://templ.guide/
- **Templ LLM Guide**: https://templ.guide/llms.md (use for AI-assisted Templ development)
- **Templ UI Components**: https://templui.io/ (pre-built UI component library)
- **Datastar**: https://github.com/delaneyj/datastar - Hypermedia-driven UI with SSE for Go/Templ
- **Echo Framework**: https://echo.labstack.com/
- **Watermill Event Bus**: https://watermill.io/
- **Server-Sent Events**: https://html.spec.whatwg.org/multipage/server-sent-events.html
- **Clean Architecture**: "Clean Architecture" by Robert C. Martin

