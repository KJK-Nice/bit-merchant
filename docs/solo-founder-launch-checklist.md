# Solo Founder Launch Checklist

This checklist is designed for running BitMerchant as a one-person business.
It is ordered by impact and speed to launch.

## Goal

Launch a paid, operable product with:

- A clear way to acquire customers
- A reliable production setup
- A supportable operations workflow
- A path to get paid as the operator

---

## Phase 1 (Week 1-2): Must-Have Before Real Customers

### 1) Public Website and Conversion Funnel

- [ ] Create a real marketing homepage (`/`) instead of redirecting to `/menu`
- [ ] Add core pages: Pricing, Features, FAQ, Contact
- [ ] Add legal pages: Terms of Service, Privacy Policy
- [ ] Add clear CTA: `Start Free Trial` / `Book Demo`
- [ ] Add lead capture form and save submissions (email + restaurant name + country)

Current code touchpoints:

- `cmd/server/routes.go` (`GET /` currently redirects)
- `internal/interfaces/templates/` (new marketing templates)

### 2) Production-Safe Tenant Onboarding

- [ ] Remove default restaurant fallback from customer menu flow
- [ ] Gate access by valid `restaurantID` or QR table link only
- [ ] Make owner signup + first restaurant setup a guided flow
- [ ] Add onboarding completion checks:
  - Menu has at least 1 category
  - Menu has at least 3 items
  - At least 1 table QR generated
  - Restaurant is marked open

Current code touchpoints:

- `internal/interfaces/http/menu.go` (fallback to `restaurant_1`)
- `cmd/server/seed.go` (seed behavior)
- `internal/interfaces/http/owner.go`
- `internal/interfaces/http/admin.go`

### 3) Reliability and Data Protection

- [ ] Enforce Postgres in production startup (fail fast if `DATABASE_URL` missing)
- [ ] Add automated DB backup job (daily + restore test weekly)
- [ ] Add monitoring/alerts:
  - App down
  - High 5xx rate
  - DB connection failure
- [ ] Add error tracking (Sentry or equivalent)

Current code touchpoints:

- `cmd/server/main.go`
- `cmd/server/database.go`
- deployment scripts/infra configs (to be added)

### 4) Basic Solo Ops Runbook

- [ ] Write `docs/runbook.md` with:
  - How to deploy
  - How to rollback
  - How to restore backups
  - How to handle restaurant outage reports
- [ ] Add `docs/support-playbook.md`:
  - Common support issues
  - Ready-to-send replies
  - Escalation rules

### 5) “You Get Paid” Layer

- [ ] Define your own business model first:
  - Flat monthly SaaS
  - Per-location fee
  - Setup fee + monthly support
- [ ] Add billing system for your customers (Stripe Billing or equivalent)
- [ ] Add account status checks (active/past-due/canceled)
- [ ] Add invoice + receipt visibility for restaurant owners

Note:

- Cash payment for restaurant orders is already in app flow.
- What is missing is *your* SaaS billing workflow.

---

## Phase 2 (Week 3-6): Strongly Recommended After First Launch

### 6) Product Analytics and Funnel Tracking

- [ ] Track website funnel: landing -> signup -> activated restaurant
- [ ] Track product activation:
  - First menu created
  - First QR generated
  - First order placed
  - First day with 10+ orders
- [ ] Build a simple weekly KPI dashboard:
  - New signups
  - Activated restaurants
  - Churned restaurants
  - MRR/ARR

### 7) Support Inbox and SLA

- [ ] Create one support channel (`support@` + help form)
- [ ] Add ticket tagging (`onboarding`, `bug`, `billing`, `feature`)
- [ ] Define response targets:
  - Critical outage: <1 hour
  - Regular support: <24 hours

### 8) CI/CD Safety Net

- [ ] Add GitHub Actions CI:
  - `templ generate` check
  - `go test ./...`
  - lint checks
- [ ] Add deployment pipeline with rollback option

Current code touchpoints:

- `.github/workflows/` (currently empty)
- `Taskfile.yml`

### 9) Security and Compliance Basics

- [ ] Security headers and cookie checks in production
- [ ] Secret rotation checklist
- [ ] Minimal audit logging for admin actions
- [ ] Data retention policy and account deletion process

---

## Phase 3 (Later): Scale and Efficiency

### 10) Multi-Location and Team Features

- [ ] Better multi-restaurant onboarding UX
- [ ] Team permissions beyond owner/kitchen
- [ ] Activity logs per user

### 11) Revenue Expansion

- [ ] Optional premium analytics package
- [ ] Optional managed onboarding service
- [ ] Partner/referral program for agencies or POS consultants

---

## Launch Readiness Scorecard (Pass/Fail)

Do not onboard paid customers until all are `PASS`.

- [ ] Marketing site + legal pages are live
- [ ] Owner can self-onboard without manual DB edits
- [ ] No production fallback to demo tenant (`restaurant_1`)
- [ ] Backups + restore drill completed
- [ ] Monitoring + alerting active
- [ ] Support inbox and response workflow live
- [ ] Billing and invoicing for your SaaS are live
- [ ] CI pipeline blocks broken merges

---

## Suggested Build Order for This Repo

1. Replace `/` redirect with real landing page.
2. Remove hardcoded tenant fallback in menu flow.
3. Add production guardrails (DB required + monitoring hooks).
4. Add CI workflow.
5. Add SaaS billing module.
6. Add analytics and support automation.

---

## Prioritized Execution Board (Solo Founder)

Effort scale:

- `S`: 2-6 hours
- `M`: 1-2 days (6-14 hours)
- `L`: 3-5 days (15-30 hours)

### P0 (Do First - Launch Blockers)

| ID | Task | Effort | ETA | Definition of Done |
|---|---|---:|---:|---|
| P0-1 | Build public landing + pricing + CTA flow | L | 3-5 days | `/` is a real marketing page, includes pricing, FAQ, contact, legal links, and lead capture works |
| P0-2 | Remove demo/default tenant fallback | M | 1-2 days | No `restaurant_1` fallback path for production traffic; invalid tenant requests fail safely |
| P0-3 | Harden onboarding path | M | 1-2 days | Owner can sign up and complete setup without manual data fixes |
| P0-4 | Enforce production DB + backup workflow | M | 1-2 days | App fails fast without `DATABASE_URL` in production; daily backups and restore test documented |
| P0-5 | Add monitoring + error tracking | M | 1-2 days | Uptime + 5xx + DB alerts configured; exceptions visible in one dashboard |
| P0-6 | Enable support channel + runbook | S | 2-6 hours | `support@` (or help form) live and `docs/runbook.md` + `docs/support-playbook.md` created |
| P0-7 | Add SaaS billing for your own revenue | L | 3-5 days | Subscription plans, billing status, invoices, and access control based on payment state |
| P0-8 | Add CI gate in GitHub Actions | S | 2-6 hours | PRs run generate/lint/test and block merge on failures |

Target: finish all P0 in 2-3 weeks solo.

### P1 (Do Next - Growth + Stability)

| ID | Task | Effort | ETA | Definition of Done |
|---|---|---:|---:|---|
| P1-1 | Add product + funnel analytics | M | 1-2 days | Track signup -> activation -> first order milestones |
| P1-2 | Weekly KPI dashboard | S | 2-6 hours | You can review signups, activated restaurants, churn, MRR weekly |
| P1-3 | Ticket tagging + SLA workflow | S | 2-6 hours | Support requests tagged and response targets enforced |
| P1-4 | Security baseline hardening | M | 1-2 days | Security headers, secret rotation checklist, admin audit logs added |

Target: 1-2 weeks after P0.

### P2 (Later - Scale Levers)

| ID | Task | Effort | ETA | Definition of Done |
|---|---|---:|---:|---|
| P2-1 | Team roles and deeper permissions | M | 1-2 days | Non-owner role matrix supports real staff operations |
| P2-2 | Multi-location UX improvements | M | 1-2 days | Managing multiple restaurants is simple and low-friction |
| P2-3 | Upsell package (premium analytics/managed onboarding) | M | 1-2 days | At least one upsell offer is live and purchasable |
| P2-4 | Referral/partner program | S | 2-6 hours | Referral source tracking and reward rules are operational |

---

## Recommended 14-Day Sprint Plan

### Week 1

1. `P0-2` Remove tenant fallback
2. `P0-3` Harden onboarding
3. `P0-8` Add CI gate
4. `P0-6` Support channel + runbooks

### Week 2

1. `P0-1` Launch landing/pricing/legal
2. `P0-4` DB/backup guardrails
3. `P0-5` Monitoring + error tracking
4. Start `P0-7` SaaS billing

### Week 3 (if needed)

1. Finish `P0-7` SaaS billing
2. Move to `P1-1` analytics baseline
