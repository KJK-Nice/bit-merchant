# BitMerchant

BitMerchant is a lightning-fast restaurant ordering platform designed for cash-first payments with a hypermedia-driven UI.

## Features

- **Customer Ordering**: Scan QR code, browse menu, order, and pay with cash. No account required.
- **Kitchen Display**: Real-time order management for kitchen staff.
- **Owner Dashboard**: Menu management and sales analytics.
- **PWA**: Installable as an app, works offline (menu browsing).
- **Performance**: Server-rendered HTML for speed, SSE for real-time updates.

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Echo v4
- **Templating**: Templ (Type-safe Go templates)
- **UI Library**: Datastar (Hypermedia) + TemplUI
- **Database**: In-memory by default, optional PostgreSQL-backed auth + core persistence via `DATABASE_URL`

## Getting Started

### Prerequisites

- Go 1.25+
- Make (optional)

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
3. Open http://localhost:8080

### Environment variables

Configuration is loaded from the process environment in [`cmd/server/config.go`](cmd/server/config.go). Copy [`.env.example`](.env.example) as reference; the Go binary does not load `.env` files automatically—use your shell, a tool like `direnv`, or Docker Compose.

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `PORT` | No | `8080` | HTTP listen port. |
| `BASE_URL` | No | `http://localhost:8080` | Public site URL (scheme + host [+ port]). Used for WebAuthn RP ID (hostname), absolute menu URLs in QR codes, and secure-cookie heuristics. Set correctly in every deployed environment. |
| `COOKIE_SECURE` | No | (off) | If `true`, session cookies are marked `Secure` (use behind HTTPS). |
| `DATABASE_URL` | No | *(empty)* | Postgres connection string. If unset, the app uses **in-memory** repositories only (no persistence). If set, **Goose migrations run automatically on startup** after the DB is reachable. |
| `S3_BUCKET_NAME` | No | *(empty)* | S3 bucket for menu item photos. If missing (with `AWS_REGION`), photo uploads are effectively disabled. |
| `AWS_REGION` | No | *(empty)* | AWS region for S3. Use `auto` for some S3-compatible providers (e.g. Cloudflare R2). |
| `S3_ENDPOINT` | No | *(empty)* | Custom S3 API base URL for **non-AWS** storage (MinIO, **Cloudflare R2**, Wasabi, etc.). Example R2: `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`. Omit for real AWS S3. |
| `S3_USE_PATH_STYLE` | No | `true` when `S3_ENDPOINT` is set, else path-style off for AWS | Path-style URLs (`endpoint/bucket/key`) vs virtual-hosted. Many compat servers need `true`. Set `false` explicitly if your provider requires virtual-hosted style. |
| `S3_PUBLIC_BASE_URL` | No | *(empty)* | Public base URL stored with menu items after upload (what the browser loads). Use your **R2 public bucket URL**, **custom domain**, or CDN origin—**not** always the same as `S3_ENDPOINT`. If unset with a custom endpoint, URLs default to `S3_ENDPOINT/bucket/key` (path-style), which may not be publicly readable until you expose that path or set this to a public hostname. |

For S3, the AWS SDK uses its normal credential chain (e.g. `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`, shared config, or instance role). Those keys are **not** defined in `config.go` but are standard for local and cloud setups.

### Optional PostgreSQL persistence

If `DATABASE_URL` is set, all repositories use PostgreSQL and Goose migrations run on startup (see [`cmd/server/database.go`](cmd/server/database.go)):

- auth: users, memberships, invitations, sessions
- core: restaurants, menu categories, menu items, orders, order items, payments, session restaurant visits

Example:

```bash
export DATABASE_URL="postgres://bitmerchant:bitmerchant@localhost:5432/bitmerchant?sslmode=disable"
go run cmd/server/main.go
```

Goose migration files live under `internal/infrastructure/migrations/sql/`.

### Docker Compose

[`docker-compose.yml`](docker-compose.yml) loads **[`.env.docker`](.env.docker)** for the `app` and `postgres` containers (database URL uses host `postgres`, not `localhost`). Run:

```bash
docker compose up --build
```

Adjust `.env.docker` for passwords, `BASE_URL` if you expose the app on another host, or optional S3 variables. If you change `POSTGRES_*`, update the Postgres `healthcheck` in `docker-compose.yml` to use the same user and database name.

### Multi-restaurant context switcher

Authenticated users can switch their active restaurant context:

- `GET /auth/select-restaurant`
- `POST /auth/select-restaurant`

During passkey login:

- users with one membership are routed directly to dashboard/kitchen based on role
- users with multiple memberships are routed to restaurant selection first

The active restaurant is stored in the server-side session and enforced by role middleware.

### Development

- Run tests:
  ```bash
  go test ./...
  ```

## Current Status

- Dashboard analytics now support `today`, `week`, and `month` date ranges with unit/integration coverage.
- CSRF protection is enabled in the Echo middleware stack, with token handling for both form submissions and Datastar-driven POST requests.
- Dashboard and mobile footer navigation were revalidated, and menu closed-state rendering was fixed (`circle-alert` icon usage).

## License

Proprietary.

