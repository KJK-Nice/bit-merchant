# AI MVP Roadmap (2 Weeks)

This roadmap is designed to ship useful AI features quickly with low infrastructure risk.
It prioritizes owner/kitchen value first, then customer conversion lift.

## Goals

- Reduce owner setup time
- Increase order throughput with fewer stock mistakes
- Improve owner decisions using actionable summaries
- Keep implementation lightweight (no custom model training)

## Scope for MVP

Ship 3 features:

1. Menu Copilot (owner onboarding + menu editing)
2. Auto 86 Risk Alert (kitchen/owner stock risk hints)
3. Smart Daily Summary (owner digest)

Out of scope for this MVP:

- Custom fine-tuning
- Multi-language voice agent
- Fully autonomous pricing changes

---

## Feature 1: Menu Copilot

### User Value

- Owner pastes rough notes (or short descriptions) and gets polished menu items.
- Faster setup for new restaurants.

### Inputs

- Raw dish text
- Optional target tone (`casual`, `premium`, `street-food`)
- Optional currency context

### Outputs

- Suggested `name`
- Suggested `description`
- Suggested `category`
- Optional suggested price band (min/max, non-binding)

### Codebase Integration

- New owner/admin action near menu creation flows:
  - `internal/interfaces/http/admin.go`
  - `internal/interfaces/templates/admin/*`
- Reuse existing menu item create/update workflows (no schema change required initially).

### Acceptance

- Owner can generate 5+ item drafts in one action.
- Can edit and save generated drafts with existing forms.
- Generation response time under 5s for small payloads.

---

## Feature 2: Auto 86 Risk Alert

### User Value

- Warn when an item likely sells out soon based on recent order velocity.
- Helps kitchen avoid over-promising unavailable items.

### Inputs

- Recent order history (last 1-3 days)
- Current item availability + rough stock hint (optional manual number)

### Outputs

- Risk level per item: `low`, `medium`, `high`
- Suggested action: keep/open, prep more, toggle unavailable soon

### Codebase Integration

- Read from existing order query data:
  - `internal/dashboard/app/query/*`
  - `internal/ordering/*`
- Show hints in dashboard/admin menu management screens.

### Acceptance

- Owner sees risk hints on menu management page.
- High-risk items clearly labeled.
- No automatic menu toggles (human confirms action).

---

## Feature 3: Smart Daily Summary

### User Value

- End-of-day AI summary with concrete actions for tomorrow.

### Inputs

- Orders today
- Top items / low performers
- Busy windows
- Restaurant open/close state context

### Outputs

- 5-10 bullet summary:
  - Revenue/order snapshot
  - Top movers
  - Underperformers
  - Recommended changes for next day

### Codebase Integration

- Build from existing dashboard stats/history/top items:
  - `internal/dashboard/app/query/*`
  - `internal/interfaces/http/dashboard.go`

### Acceptance

- Summary visible on dashboard for selected date range.
- Includes at least 3 actionable recommendations.
- Owner can regenerate summary.

---

## Architecture (MVP-safe)

- Add one AI gateway interface in application layer:
  - `GenerateMenuDrafts(...)`
  - `PredictSelloutRisk(...)`
  - `GenerateDailySummary(...)`
- Start with a single provider adapter (OpenAI API) behind interface.
- Add strict prompt templates + JSON schema validation on responses.
- Add timeout/retry + safe fallback:
  - If AI fails, app still works with normal non-AI flow.

## Data + Safety

- Do not send PII beyond what is needed.
- Mask session/user identifiers before provider call.
- Log request IDs and latency, not raw sensitive payloads.
- Add feature flag per AI feature for staged rollout.

---

## Delivery Plan (14 Days)

### Week 1

1. Day 1-2: AI gateway + provider adapter + config/env wiring
2. Day 3-4: Menu Copilot UI + endpoint + tests
3. Day 5: Internal QA + prompt tuning + guardrails

### Week 2

1. Day 6-8: Auto 86 Risk Alert computation + dashboard/admin UI
2. Day 9-10: Smart Daily Summary generation + dashboard integration
3. Day 11-12: Test hardening, latency checks, fallback behavior
4. Day 13-14: Rollout flags, docs, and production readiness checklist

---

## Metrics to Track

Menu Copilot:
- Time to first complete menu
- % of generated drafts saved

Auto 86:
- # of high-risk alerts shown
- # of manual availability toggles after alerts

Daily Summary:
- % owners viewing summary
- % summaries regenerated
- Correlation with next-day item updates

Global:
- AI error rate
- AI response latency p50/p95
- Cost per restaurant per day

---

## Launch Strategy

1. Enable for internal/demo restaurant first.
2. Enable for 2-3 pilot restaurants.
3. Collect qualitative feedback for 1 week.
4. Adjust prompts and thresholds.
5. Expand to all restaurants with feature flags still available for rollback.

