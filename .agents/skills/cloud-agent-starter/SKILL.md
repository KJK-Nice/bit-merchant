---
name: cloud-agent-starter
description: Minimal Cloud-agent playbook for running and testing BitMerchant quickly, with practical auth, env-toggle, and area-by-area test workflows.
---

# BitMerchant Cloud Agent Starter

Use this skill at the start of any Cloud task that touches runtime behavior, auth, routing, templates, or tests.

## 1) Quick bootstrap (first commands to run)

Run from repo root:

1. `go mod download`
2. `npm install`
3. `npx playwright install --with-deps chromium`
4. `templ generate`

Fast local app start (in-memory mode, no Postgres required):

```bash
export PORT=8080
export PUBLIC_BASE_URL=http://localhost:8080
export CUSTOMER_BASE_URL=http://order.localhost:8080
export MERCHANT_BASE_URL=http://merchant.localhost:8080
export DISABLE_RATE_LIMIT=true
unset DATABASE_URL
go run ./cmd/server
```

Notes:

- `*.localhost` host-surface routing works without `/etc/hosts` edits.
- Use in-memory mode by leaving `DATABASE_URL` unset for quick iteration.
- If you changed `.templ` files, always run `templ generate` before testing.

## 2) Auth & merchant area (passkey login)

Important behavior:

- Auth is passkey-first (`/auth/signup`, `/auth/login`), not password-based.
- Post-login can redirect to `/auth/select-restaurant` for multi-membership users.
- Protected merchant routes are `/dashboard/*`, `/admin/*`, and `/kitchen/*` (role-based).

Cloud-friendly testing workflow:

1. Run merchant auth/navigation E2E:
   - `npm run e2e:test -- e2e/playwright/specs/merchant-auth-navigation-journey.spec.ts`
2. Run multi-restaurant and invite flows when touching auth context logic:
   - `npm run e2e:test -- e2e/playwright/specs/merchant-multi-restaurant-selection.spec.ts`
   - `npm run e2e:test -- e2e/playwright/specs/merchant-invite-kitchen-journey.spec.ts`

Practical note:

- Playwright merchant actors already use a virtual authenticator, so passkey flow runs headlessly in Cloud.

## 3) Customer ordering area

Key routes:

- Menu: `/menu?restaurantID=restaurant_1`
- Cart: `/cart`
- Checkout/create order: `/order/confirm`, `/order/create`
- Order status: `/order/:orderNumber`

Cloud-friendly testing workflow:

1. `npm run e2e:test -- e2e/playwright/specs/customer-order-journey.spec.ts`
2. `npm run e2e:test -- e2e/playwright/specs/customer-order-resilience.spec.ts`

Use these when changing customer menu/cart/order handlers, templates, or redirects.

## 4) Admin/menu/QR area (owner-only)

Key routes:

- Admin dashboard/menu management: `/admin/dashboard`
- QR management: `/admin/qr`

Cloud-friendly testing workflow:

1. `npm run e2e:test -- e2e/playwright/specs/admin-menu-qr-management.spec.ts`
2. Optional HTTP contract checks for route/handler regressions:
   - `go test ./tests/contract/http/...`

Use this when editing menu CRUD, item availability toggles, QR generation/settings, or admin templates.

## 5) Kitchen operations area

Key route:

- Kitchen display and lifecycle actions: `/kitchen`

Cloud-friendly testing workflow:

1. `npm run e2e:test -- e2e/playwright/specs/kitchen-order-lifecycle.spec.ts`
2. If touching kitchen domain/app logic, also run:
   - `go test ./tests/unit/application/kitchen/...`
   - `go test ./tests/integration/kitchen/...`

## 6) Persistence & infra area (in-memory vs Postgres)

Default quick mode (recommended for most Cloud tasks):

- Keep `DATABASE_URL` unset (in-memory repositories, fast startup).

Postgres mode (when editing adapters/migrations/persistence):

1. `docker compose up -d postgres`
2. `export DATABASE_URL='postgres://bitmerchant:bitmerchant@localhost:5432/bitmerchant?sslmode=disable'`
3. `go run ./cmd/server` (startup runs Goose migrations automatically)
4. `go test -v ./tests/integration/postgres/...`

Optional explicit migration run:

- `go run ./cmd/migrate/main.go`

## 7) Env toggles and "feature flag" equivalents

There is no broad product feature-flag framework yet; use these env toggles to control test behavior:

- `DISABLE_RATE_LIMIT=true` -> disables rate limiting for deterministic automation.
- `DATABASE_URL` unset -> mocks persistence with in-memory repositories.
- `PUBLIC_BASE_URL`, `CUSTOMER_BASE_URL`, `MERCHANT_BASE_URL` -> controls host-surface routing and redirects.
- `COOKIE_SECURE=true|false` -> force secure cookie behavior in HTTPS vs local HTTP.
- S3 vars unset -> effectively disable photo upload path when not under test.

## 8) Baseline command sets by codebase area

Use these compact command bundles after edits:

- Server/config wiring:
  - `go test ./cmd/server/...`
- Domain + application logic:
  - `go test ./tests/unit/...`
- HTTP route/handler contracts:
  - `go test ./tests/contract/http/...`
- In-memory integration slices:
  - `go test ./tests/integration/admin/... ./tests/integration/auth/... ./tests/integration/customer/... ./tests/integration/dashboard/... ./tests/integration/kitchen/...`
- Host routing/session/auth e2e smoke:
  - `npm run e2e:test -- e2e/playwright/specs/routing-host-surfaces.spec.ts`
  - `npm run e2e:test -- e2e/playwright/specs/session-cookie-isolation.spec.ts`

## 9) How to update this skill (keep it useful)

Whenever you discover a new testing trick or runbook step:

1. Update this file in the same PR that proved the new workflow.
2. Add it to the correct area section (Auth, Customer, Admin, Kitchen, Persistence).
3. Include exact command(s), preconditions, and whether Docker is required.
4. Prefer smallest reliable test command over broad/full-suite commands.
5. Remove obsolete steps immediately when routes/env vars/test paths change.
