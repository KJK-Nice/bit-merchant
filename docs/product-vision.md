# BitMerchant — Product Vision

> Last updated: 2026-04-18

## Mission

**Help every small restaurant get paid instantly, keep more of every sale, and run smarter — powered by Bitcoin and AI, usable by anyone with a phone.**

## North Star

**The Lightning-native POS for cash-and-sats economies.** Cash stays as a first-class fallback; Lightning becomes the default digital rail. Zero card fees, instant settlement, no chargebacks, cross-border by default.

## Why this wedge

- **Cash is not dead.** In SEA, LATAM, South Asia, and informal economies, 60–80% of F&B transactions are cash. Western POS vendors treat cash as second-class; we treat it as default.
- **No hardware lock-in.** A street-side stall with one phone can run BitMerchant. That is a ~10x larger TAM than terminal-based POS.
- **Lightning economics.** ~0.1% routing vs. 2–3% card fees. On a $300K/yr restaurant, that is $6–9K/yr back in the owner's pocket — more than our SaaS fee.
- **Tourist-heavy beachheads** (Chiang Mai, Lugano, San Salvador, Cape Town, Lisbon, Buenos Aires): Bitcoiners travel with sats and actively seek merchants who accept them. Being on BTC Map is free CAC.
- **5-minute onboarding.** Just a Lightning address (LNURL / BOLT12) — no merchant account, no terminal, no KYC.

## 12-month business goals

1. **Land 500 paying merchants** in one beachhead city — prove density and word-of-mouth before geo-expanding.
2. **<5 min onboarding** from signup to first order — measured, not aspirational.
3. **$15–25/mo SaaS** with **0% transaction fee** — undercut Toast/Square by being the only vendor not taking a payment cut.

## 3-year bet

Expand from ordering → **inventory, supplier payments, and working-capital loans** underwritten by order-flow data. The ordering app is the wedge; **embedded fintech for SMB restaurants** is the business.

---

## Product pillars

### 1. Payments — Lightning-first, cash-always

| Method | Role | Notes |
|--------|------|-------|
| **Lightning** | Default digital rail | **Non-custodial by default.** Vendors link their own Lightning address / LNURL-pay / BOLT12 offer. Payments settle directly at the vendor's wallet or node — we never touch the funds. |
| **Cash** | First-class fallback | Manual mark-paid flow from kitchen. Core to emerging-market coverage. |
| **Cards** | Out of scope | Explicitly not on the roadmap. The whole point is to route around card rails. |

**Why non-custodial-first.** We are a *software interface*, not a money transmitter. Settlement-at-the-edge eliminates custody liability, AML/KYC burden in most jurisdictions, and the security cost of holding other people's money. Vendors keep their keys and their trust in us is higher because of it.

Optional auto-convert percentage on Lightning receipts (e.g. "20% sats / 80% bank") is a wallet-side concern we surface as guidance, not something we broker.

### 2. AI features — narrow, measurable, life-changing

Every AI feature ships tied to a metric. No generic chatbots.

**For merchants**
- **Menu photographer.** Phone pic → cleaned background, multilingual description, allergen tags, price suggestion from local comps.
- **Dynamic prep-time predictor.** Learns from kitchen throughput + current queue; shows customers honest ETAs. *Metric: prep-time MAE.*
- **Daily close-out coach.** "You sold 40% fewer pad krapow than Tuesdays. Chili supplier price spiked 12%. Consider a weekly special." Actionable, not dashboard porn.
- **Voice order entry in the kitchen** (Thai / Spanish / English). Hands stay on the wok.

**For customers**
- **Order like a local.** Photo-first menu in the user's language, dietary filter, a confidence chip ("3 friends-of-friends ordered this").
- **Split-the-bill in sats.** Group scans one QR, AI parses the receipt, each person pays their Lightning share.
- **Return-visit memory.** "Last time you loved the #4 with extra basil. Same again?" Opt-in, session-scoped, no account required.

### 3. Monetization — SaaS subscription, USD-pegged, settled in sats

Because payments are non-custodial, we cannot skim a transaction fee — and we do not want to. We charge vendors directly for the software.

- **Primary:** Monthly subscription, **priced in USD** (e.g. $15–25/mo) and **paid in sats** at the spot rate at invoice time. Predictable bill for the vendor; we absorb no FX risk beyond the ~24h invoice validity window.
- **Payment mechanic:** "Prepaid days" model — vendor pays a Lightning invoice, we credit them N days of access. A daily cron flips access off when `expires_at < now()`.
- **Ledger of record:** [TigerBeetle](https://tigerbeetle.com) double-entry accounts — not Postgres. Vendor subscription balance lives as a TigerBeetle account with `DebitsMustNotExceedCredits`, so access enforcement is a DB-level guarantee, not application logic.
- **Fallback tier:** Pay-per-invoice credits for vendors who won't commit to a subscription (e.g. seasonal stalls, weekend markets). Same TigerBeetle ledger, different `code`.
- **Not shipping:** Transaction-percentage fees, LSAT micro-payments for features, LSP referral kickbacks. All considered and deferred — subscription-first keeps the pricing story simple.

### 4. Hypermedia-first UI

Server-rendered HTML, Datastar + SSE for real-time, PWA with offline menu browsing. No SPA, no JSON API for the merchant/customer surfaces. Small bundle, fast on cheap Android phones — the devices our merchants and customers actually own.

---

## Architecture at a glance

DDD Lite with bounded contexts (see [ddd-lite.md](ddd-lite.md) for conventions):

```
internal/
  restaurant/   Open/close, tables, QR, vendor Lightning address
  menu/         Categories, items, photos (AI: photographer)
  ordering/     Cart, order lifecycle, kitchen queue (AI: prep-time, voice)
  payment/      Lightning (non-custodial, via LSP port), cash (fallback); PaymentMethod port
  billing/      Platform subscription: Lightning invoicing, TigerBeetle ledger, access gate
  auth/         WebAuthn passkeys, sessions, memberships
  places/       Customer visit history (powers return-visit memory)
  dashboard/    Reporting + AI close-out coach
```

Stack: Go 1.26, Echo v4, Templ, Datastar, Watermill, Postgres (product state), **TigerBeetle (subscription/credit ledger)**, S3 (optional). Lightning via an LSP port — see open decision below.

---

## Scope boundaries (what we will NOT build)

- **Card payments.** Ever. That is the whole thesis.
- **Generic AI chatbot.** Only narrow, metric-backed AI features.
- **Loyalty points.** Sats *are* the loyalty program — tip and save in sats.
- **Order customisations in v1.** Fixed items. Ship the wedge, add complexity later.
- **Multi-user restaurant management in v1.** One owner, one staff kitchen view.
- **Refund processing.** Out-of-band via the merchant's own wallet for v1.

---

## Historical specs — what changed, what's kept

The original `specs/` directory contains two earlier product specs that seeded this vision. They are archived under `specs/archive/` for reference; this document supersedes them.

### `001-lightning-restaurant-platform` (archived)

- **Purpose:** First design for a Lightning-native ordering platform using the Strike API.
- **Kept:** Core data model (Restaurant, MenuCategory/Item, Order, Payment, OrderItem), kitchen display flow, real-time SSE updates, PWA, fiat↔sats conversion at invoice time, daily settlement to merchant Lightning address.
- **Changed:** Strike API is no longer the only integration path — we now plan a pluggable provider layer (Breez / Voltage / Strike custodial; LND / CLN / Phoenixd self-custodial) with custodial-first onboarding and a "graduate to your own node" upsell.

### `002-cash-payment-hypermedia` (archived)

- **Purpose:** Cash-first MVP with a `PaymentMethod` port designed to admit Lightning later without refactoring.
- **Kept:** The hypermedia architecture (Templ + Datastar + SSE, HTML-returning handlers, no JSON API on customer/merchant surfaces), the `PaymentMethod` abstraction, manual "mark paid" flow from kitchen, offline menu PWA.
- **Changed:** Cash is no longer "v1.0 with Lightning later" — Lightning and cash ship side by side as first-class peers. AI features were not in the original spec; they are now a product pillar.

---

## Key tradeoffs on the table

1. **LSP selection (open).** We need one primary Lightning Service Provider behind the `payment/` port before v1. Shortlist:
   - **Breez SDK** — Go-friendly, supports both custodial and self-custodial (Greenlight) paths; best "graduate to your own node" story.
   - **Alby** — strong developer API, LNURL-native, good webhook ergonomics.
   - **Ibex** — marketplace-oriented API, explicit multi-vendor sub-account model.
   - **Galoy** — open-source stack, strong in LATAM.
   - Decision criteria: Go SDK quality, webhook reliability, fee structure, regulatory posture in beachhead jurisdiction.
2. **Beachhead-city depth vs. horizontal breadth.** Current call: depth over breadth — one city, one vertical (small restaurants), win density, then expand.
3. **Regulatory exposure.** Non-custodial settlement-at-the-edge dramatically reduces our footprint here, but beachhead selection is still a legal decision. Green: El Salvador, select Swiss cantons. Grey: Argentina. Red (as of last check): Thailand bans crypto-for-payments. Legal review required per beachhead before launch.
4. **Subscription flavor — prepaid-days vs. metered daily burn.** Current call: prepaid-days for v1 (simpler cron, clearer vendor mental model). Metered daily burn is a premium-tier option later.
5. **Pricing currency.** USD-pegged bill, settled in sats at invoice time. Vendors get a predictable software cost; we eat no volatility beyond the invoice validity window.

---

## How to use this doc

- **New contributor?** Read this, then [ddd-lite.md](ddd-lite.md), then the README.
- **Writing a new feature spec?** Create it under `specs/NNN-feature-name/` and link back here. Do not re-litigate the mission or payment stance in feature specs — reference this document.
- **Disagree with a pillar?** Open a discussion, not a PR. Changing pillars is a CEO-level call, not a code review.
