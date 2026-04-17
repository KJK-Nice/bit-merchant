# BitMerchant

BitMerchant is a lightning-fast restaurant ordering platform designed for cash-first payments with a hypermedia-driven UI.

## Features

- **Customer Ordering**: Scan QR code, browse menu, order, and pay with cash. No account required.
- **Kitchen Display**: Real-time order management for kitchen staff.
- **Owner Dashboard**: Menu management and sales analytics.
- **PWA**: Installable as an app, works offline (menu browsing).
- **Performance**: Server-rendered HTML for speed, SSE for real-time updates.

## Tech Stack

- **Language**: Go 1.26+
- **Web Framework**: Echo v4
- **Templating**: Templ (Type-safe Go templates)
- **UI Library**: Datastar (Hypermedia) + TemplUI
- **Database**: In-memory by default, optional PostgreSQL-backed auth + core persistence via `DATABASE_URL`
- **Events**: Watermill (in-process event bus for order lifecycle)
- **Logging**: `log/slog` with [humanslog](https://github.com/ThreeDotsLabs/humanslog) for pretty development output

## Architecture

The project follows **DDD Lite** (Domain-Driven Design Lite) with bounded contexts, inspired by [Three Dots Labs](https://threedots.tech/post/ddd-lite-in-go-introduction/). For conventions, layering, and how to add features in this layout, see `[docs/ddd-lite.md](docs/ddd-lite.md)`.

### Bounded Contexts

```
internal/
  common/                           Shared kernel (IDs, EventBus interface)

  restaurant/                       Restaurant management
    domain/restaurant/              Aggregate: Restaurant (open/close, table count)
    app/command/                    CreateRestaurant, ToggleOpen, UpdateTableCount
    app/query/                      GenerateQR
    adapters/                       Postgres + memory repositories

  menu/                             Catalog management
    domain/menu/                    Aggregates: MenuCategory, MenuItem, PhotoStorage port
    app/command/                    CRUD for categories, items, photos
    app/query/                      GetMenu, GetMenuAdmin
    adapters/                       Postgres + memory repos, S3 storage

  ordering/                         Order lifecycle (cart + kitchen)
    domain/order/                   Aggregate: Order (fulfillment state machine, events)
    app/command/                    CreateOrder, MarkPaid, MarkPreparing, MarkReady
    app/query/                      GetOrder, GetCustomerOrders, GetKitchenOrders
    app/cart/                       CartService (application-level, in-memory)
    adapters/                       Postgres + memory repos

  payment/                          Payment processing
    domain/payment/                 Aggregate: Payment, PaymentMethod port
    adapters/                       Cash method, Postgres + memory repos

  auth/                             Identity and access
    domain/user/                    User aggregate (WebAuthn credentials)
    domain/session/                 Session aggregate
    domain/membership/              Membership aggregate (user-restaurant-role)
    domain/invitation/              Invitation aggregate
    adapters/                       WebAuthn service, Postgres + memory repos

  places/                           Customer visit tracking
    domain/visit/                   SessionRestaurantVisit aggregate
    app/command/                    RecordVisit
    app/query/                      ListVisited
    adapters/                       Postgres + memory repos

  dashboard/                        Reporting (query-only, no domain)
    app/query/                      GetStats, GetHistory, GetTopItems

  interfaces/                       Delivery layer (stays flat)
    http/                           Echo HTTP handlers
    http/middleware/                 Session, auth, CSRF, rate limiting
    templates/                      Templ components and layouts
```

### Shared Kernel

`internal/common/` contains cross-boundary value types:

- `ids.go` -- All ID types (`RestaurantID`, `OrderID`, `UserID`, etc.), role constants, status enums
- `events.go` -- `EventBus` and `DomainEvent` interfaces

### Imports

Use bounded-context packages directly, for example `bitmerchant/internal/<context>/domain/...`, `app/command`, `app/query`, and `app/cart`. Legacy facade imports under `internal/domain/` and `internal/application/` are no longer supported.

### Infrastructure

- `internal/infrastructure/events/` -- Watermill-based event bus (shared)
- `internal/infrastructure/logging/` -- Structured logging
- `internal/infrastructure/migrations/` -- Goose SQL migrations
- `internal/infrastructure/qr/` -- QR code generation

## Getting Started

### Prerequisites

- Go 1.26+
- Docker (for testcontainers integration tests)

### Installation

1. Clone the repository
2. Install dependencies:
  ```bash
   go mod download
  ```
3. Install templ:
  ```bash
   go install github.com/a-h/templ/cmd/templ@latest
  ```

### Running the App

1. Generate templates:
  ```bash
   templ generate
  ```
2. Run the server:
  ```bash
   go run cmd/server/main.go
  ```
3. Open [http://localhost:8080](http://localhost:8080)

### Environment variables

Configuration is loaded from the process environment in `[cmd/server/config.go](cmd/server/config.go)`. Copy `[.env.example](.env.example)` as reference; the Go binary does not load `.env` files automatically--use your shell, a tool like `direnv`, or Docker Compose.


| Variable                 | Required | Default                                                            | Purpose                                                                                                                                                                                   |
| ------------------------ | -------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `APP_ENV`                | No       | *(empty)*                                                          | Set to `production` for JSON log output. Any other value (including unset) uses [humanslog](https://github.com/ThreeDotsLabs/humanslog) pretty output for development.                    |
| `PORT`                   | No       | `8080`                                                             | HTTP listen port.                                                                                                                                                                         |
| `BASE_URL`               | No       | `http://localhost:8080`                                            | Backward-compatible base URL fallback used when surface-specific URLs are not set.                                                                                                        |
| `PUBLIC_BASE_URL`        | No       | `BASE_URL`                                                         | Public/marketing host URL (for `/` landing surface).                                                                                                                                      |
| `CUSTOMER_BASE_URL`      | No       | `BASE_URL`                                                         | Customer app host URL (`/menu`, `/order/`*, `/cart/`*, `/my-places`) and QR menu link base.                                                                                               |
| `MERCHANT_BASE_URL`      | No       | `BASE_URL`                                                         | Merchant app host URL (`/dashboard/*`, `/admin/*`, `/kitchen/*`, `/auth/*`). Also used for WebAuthn RP origin/RPID derivation.                                                            |
| `COOKIE_SECURE`          | No       | (off)                                                              | If `true`, session cookies are marked `Secure` (use behind HTTPS).                                                                                                                        |
| `DATABASE_URL`           | No       | *(empty)*                                                          | Postgres connection string. If unset, the app uses **in-memory** repositories only (no persistence). If set, **Goose migrations run automatically on startup** after the DB is reachable. |
| `AWS_S3_BUCKET_NAME`     | No       | *(empty)*                                                          | S3 bucket for menu item photos. If missing (with region), photo uploads are disabled. Alias: `S3_BUCKET_NAME`.                                                                            |
| `AWS_DEFAULT_REGION`     | No       | *(empty)*                                                          | Region for S3 (SDK + presigning). Use `auto` for some providers (e.g. Cloudflare R2). Alias: `AWS_REGION`.                                                                                |
| `AWS_ENDPOINT_URL`       | No       | *(empty)*                                                          | Custom S3 API base URL for **non-AWS** storage (MinIO, R2, Wasabi). Example R2: `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`. Omit for real AWS S3. Alias: `S3_ENDPOINT`.              |
| `S3_USE_PATH_STYLE`      | No       | `true` when `AWS_ENDPOINT_URL` is set, else path-style off for AWS | Path-style URLs vs virtual-hosted. Many compat servers need `true`.                                                                                                                       |
| `S3_PUBLIC_BASE_URL`     | No       | *(empty)*                                                          | Optional. Used to derive the **object key** from **legacy** menu rows that still store a full public URL (before keys-only storage). Not required for new uploads.                        |
| `S3_PRESIGN_GET_EXPIRES` | No       | `3600`                                                             | Seconds until each **presigned GET** URL for menu photos expires (private buckets). Use a larger value if customers keep the menu open longer than an hour.                               |


Credentials: set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` (or use the SDK default chain, e.g. instance role). They are read by the AWS SDK, not listed in `config.go`.

**Private buckets:** uploads use `PutObject` with credentials; the database stores the **object key**. The customer menu uses **presigned GET** URLs so browsers can load images without making the bucket public.

### Optional PostgreSQL persistence

If `DATABASE_URL` is set, all repositories use PostgreSQL and Goose migrations run on startup (see `[cmd/server/database.go](cmd/server/database.go)`):

- auth: users, memberships, invitations, sessions
- core: restaurants, menu categories, menu items, orders, order items, payments, session restaurant visits

Example:

```bash
export DATABASE_URL="postgres://bitmerchant:bitmerchant@localhost:5432/bitmerchant?sslmode=disable"
go run cmd/server/main.go
```

Goose migration files live under `internal/infrastructure/migrations/sql/`.

### Docker Compose

`[docker-compose.yml](docker-compose.yml)` loads `**[.env.docker](.env.docker)**` for the `app` and `postgres` containers (database URL uses host `postgres`, not `localhost`).

The stack now includes a Caddy TLS reverse-proxy:

- Caddy listens on `:8080` with HTTPS
- App listens internally on `app:8081`
- Postgres remains on `:5432`

1. Add local hostnames once:

```bash
echo "127.0.0.1 bitmerchant.local merchant.bitmerchant.local order.bitmerchant.local" | sudo tee -a /etc/hosts
```

1. Start the stack:

```bash
docker compose up --build
```

1. Trust Caddy's local CA (one-time, if browser warns):

```bash
docker compose cp caddy:/data/caddy/pki/authorities/local/root.crt ./tmp/caddy-root.crt
```

Import `./tmp/caddy-root.crt` into your OS/browser trust store.

Then open:

- `https://bitmerchant.local:8080`
- `https://merchant.bitmerchant.local:8080/auth/login`
- `https://order.bitmerchant.local:8080/menu?restaurantID=restaurant_1`

Adjust `.env.docker` for passwords and host URLs (`BASE_URL` and/or `PUBLIC_BASE_URL` / `CUSTOMER_BASE_URL` / `MERCHANT_BASE_URL`) if you expose the app on another host, or optional S3 variables. If you change `POSTGRES_*`, update the Postgres `healthcheck` in `docker-compose.yml` to use the same user and database name.

### Multi-restaurant context switcher

Authenticated users can switch their active restaurant context:

- `GET /auth/select-restaurant`
- `POST /auth/select-restaurant`

During passkey login:

- users with one membership are routed directly to dashboard/kitchen based on role
- users with multiple memberships are routed to restaurant selection first

The active restaurant is stored in the server-side session and enforced by role middleware.

### Development

- Install git hooks (run once after cloning):
  ```bash
  task hooks:install
  ```
  This registers a `pre-commit` hook via [lefthook](https://github.com/evilmartians/lefthook) that automatically runs `templ generate` and `golangci-lint` before each commit. Requires `lefthook` (`brew install lefthook`) and `golangci-lint` to be installed.
- Run all tests (unit + integration with in-memory repos):
  ```bash
  go test ./...
  ```
- Run Postgres integration tests (requires Docker):
  ```bash
  go test -v ./tests/integration/postgres/...
  ```
  These use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up a real Postgres container, run all Goose migrations, and exercise every adapter.

### E2E tests (Playwright)

The Playwright suite includes:

- host-surface routing/canonical redirect checks
- session cookie isolation checks
- full customer journey (menu -> cart -> checkout -> order status)
- full merchant core journey with real passkey flow (signup + dashboard/admin/qr/kitchen access + logout)

Planned and covered E2E journeys are tracked in `[docs/e2e-user-story-checklist.md](docs/e2e-user-story-checklist.md)`.

Host-surface tests run with `*.localhost` domains:

- public: `http://localhost:8080`
- customer: `http://order.localhost:8080`
- merchant: `http://merchant.localhost:8080`

No `/etc/hosts` edits are required.

Install dependencies and browser:

```bash
npm install
npx playwright install --with-deps chromium
```

Run E2E smoke tests:

```bash
npm run e2e:test
```

Run E2E with mobile viewport emulation:

```bash
npm run e2e:test:mobile
```

Passkey note:

- Merchant journey uses Playwright Chromium CDP virtual authenticator (no backend bypass endpoints).

Debug/UI modes:

```bash
npm run e2e:test:ui
npm run e2e:test:debug
npm run e2e:test:mobile:ui
npm run e2e:test:mobile:debug
```

## License

Proprietary.