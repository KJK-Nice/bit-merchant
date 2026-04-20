# DDD Lite in BitMerchant

This document explains how we apply **DDD Lite** (Domain-Driven Design Lite) in this codebase. The approach is inspired by [Three Dots Labs’ introduction to DDD Lite in Go](https://threedots.tech/post/ddd-lite-in-go-introduction/), which emphasizes clear boundaries and behavior-rich domain code without heavy tactical DDD ceremony.

## Goals

- **Bounded contexts** — Each major area of the product lives in its own tree under `internal/<context>/`, so ownership and dependencies stay obvious.
- **Valid domain state** — Aggregates and entities enforce invariants through constructors and methods rather than anemic structs + scattered validation.
- **Ports and adapters** — Domain and application code depend on interfaces (repositories, payment methods, photo storage). Concrete Postgres, memory, S3, and cash implementations live in `adapters/`.
- **Testable use cases** — Application services take interfaces and can be tested with in-memory fakes or integration tests against real Postgres (see `tests/integration/postgres/`).

DDD Lite here is **not** a mandate for ubiquitous language workshops, event sourcing, or a strict aggregate taxonomy on every type. It is a **structure and discipline** for where code belongs and how layers talk to each other.

## Bounded contexts

| Context | Path | Role |
|--------|------|------|
| **Restaurant** | `internal/restaurant/` | Restaurant lifecycle: open/close, table count, QR-related coordination at the domain boundary. |
| **Menu** | `internal/menu/` | Catalog: categories, items, availability, photo storage port and S3 adapter. |
| **Ordering** | `internal/ordering/` | Orders, fulfillment state, cart service, kitchen/customer queries. |
| **Payment** | `internal/payment/` | Payment model and methods (e.g. cash); persistence adapters. |
| **Auth** | `internal/auth/` | Users, sessions, memberships, invitations, WebAuthn-facing adapters. |
| **Places** | `internal/places/` | Session-scoped “visited restaurant” tracking. |
| **Dashboard** | `internal/dashboard/` | Read-side reporting (stats, history, top items); no separate domain package beyond queries. |

**Delivery and cross-cutting** code:

- `internal/wiring/` — Composition-root shared types: `Config`, `Repositories` (all stores), `InitPhotoStorage`, `ConnectDatabase`, and demo `SeedData`.
- `internal/<context>/service/` — Wires that context’s application handlers and HTTP ports (`package service`). The root `internal/service` package composes these into `app.Application`.
- `internal/<context>/ports/http/` — Echo HTTP handlers (route-specific) for that bounded context (`package http`, imported with an alias such as `authhttp`, `orderinghttp`).
- `internal/common/http/middleware/` — Shared Echo middleware (session, authz, CSRF, rate limiting, surface routing, …).
- `internal/common/http/` (`commonhttp`) — Shared request helpers (auth context keys, layout labels, SSE hub used by ordering projections).
- `internal/common/server/` — Shared HTTP transport: `server.Component` + `Run` (Echo, global middleware, static files, graceful shutdown). `cmd/server` composes the app then runs this component.
- `internal/interfaces/templates/` — Templ UI.
- `internal/infrastructure/events/` — Watermill event bus infrastructure (in-memory or NATS JetStream) and handlers.
- `internal/infrastructure/migrations/` — Goose SQL migrations.
- `internal/infrastructure/logging/`, `internal/infrastructure/qr/` — shared technical services.

## Layout inside a context

Typical shape (example: `ordering`):

```text
internal/ordering/
  domain/order/          # Order aggregate, invariants, domain events, Repository interface
  app/command/           # Write-oriented use cases (create order, mark paid, …)
  app/query/             # Read-oriented use cases
  app/cart/              # Application-level cart (in-memory; not all contexts need this)
  adapters/              # Postgres + memory implementations of domain ports
```

- **`domain/`** — Types that express business rules. Interfaces for persistence live next to the aggregate that needs them (e.g. `order.Repository`).
- **`app/app.go`** — Optional `Application{Commands, Queries}` bundle for the context (used in `restaurant`, `places`, `dashboard`; other contexts are migrating).
- **`app/command` and `app/query`** — Orchestration: load data, call domain methods, persist, publish events. Refactored handlers follow the Three Dots style: small **command/query structs**, `Handle(ctx, …)` methods, exported `XHandler` type aliases to `decorator.CommandHandler` / `QueryHandler` / `CommandResultHandler`, and `NewXHandler(...)` constructors that wrap with `Apply*Decorators` (see `internal/common/decorator/`).
- **`adapters/`** — Infrastructure: SQL, external APIs, in-memory stores for tests and default dev mode.

## Shared kernel

`internal/common/` holds **small, deliberately shared** building blocks:

- **IDs and enums** — e.g. `RestaurantID`, `OrderID`, role constants, fulfillment/payment statuses (`ids.go`).
- **Event contracts** — `DomainEvent`, `EventBus`, and topic name constants (`events.go`).

Keep this package lean. If something is only relevant to one context, it belongs under that context’s `domain/` package, not in `common`.

## Cross-context dependencies

Use cases may depend on **another context’s domain ports** (interfaces), not on that context’s adapters. For example, creating an order loads a restaurant through `restaurant.Repository` and builds line items using menu-related data—dependencies are expressed as interfaces injected at construction time in `cmd/server`.

The **composition root** is `internal/service` (`service.NewApplication`, `service.Application`); `cmd/server` loads config and runs the HTTP `common/server` component. Those packages may import adapters and wire concrete types. Domain and application packages should not reach “out” to HTTP, SQL drivers, or global singletons.

## Domain events

- Event **types** and **names** are owned by the domain packages that produce them (often alongside the order aggregate).
- Publishing goes through `internal/common.EventBus`; the default implementation is shared Watermill-based infrastructure under `internal/infrastructure/events/`.
- **Topic strings** live in `internal/common` (`EventOrderCreated`, etc.) so subscribers and publishers stay aligned.

This keeps the domain aware of *what happened* without coupling it to Watermill’s concrete router implementation.

## Imports

Import bounded-context packages directly from `bitmerchant/internal/<context>/...`.

- Use `domain/<aggregate>` packages for aggregates and repository interfaces.
- Use `app/command`, `app/query`, and `app/cart` packages directly from the owning context.
- Do not add compatibility facades under `internal/domain` or `internal/application`.

## Adding a feature (checklist)

1. Identify the **bounded context** (or whether it is truly cross-cutting infrastructure).
2. Put **rules and state transitions** on domain types in `domain/`.
3. Add or extend **repository interfaces** in the same domain package; implement them in `adapters/` (Postgres and/or memory).
4. Add **command or query** types under `app/command` or `app/query`.
5. Wire dependencies in **`internal/service`** (and add route registration in `cmd/server` if needed).
6. If persistence or SQL changes, add a **Goose migration** and consider extending **`tests/integration/postgres/`** so adapters stay aligned with the schema.

## Further reading

- [DDD Lite in Go — introduction](https://threedots.tech/post/ddd-lite-in-go-introduction/) — mental model and motivation.
- [`README.md`](../README.md) — high-level architecture diagram and bounded-context tree.
- [`docs/auth-design.md`](auth-design.md) — auth-specific flows and file locations.
