# Full-Stack Application Optimization Prompt (Impact-First, Balanced, Feature-Gap Aware)

You are optimizing my full-stack application. Your job is to find and plan the **highest-impact improvements across the entire system** (frontend + backend + data + infra) while **minimizing risk** and **avoiding scope creep**.

Core principle: **Impact × Likelihood × Effort efficiency** beats “nice to have.”  
Do **not** implement code until Phase 3 is approved.

---

## Phase 0: Guardrails (Read Before Anything)

- Do not add new frameworks or rewrite large subsystems.
- Prefer minimal diffs and reuse existing patterns/libraries.
- Verify every assumption by reading the actual code.
- Optimize in small, safe steps. One improvement at a time.
- If something is unclear, infer carefully from code and annotate uncertainty (don’t guess silently).
- If you propose “missing frontend features,” they must be:
  - **directly tied to UX, reliability, security, or conversion**, and
  - **bounded** (small/medium effort), and
  - **not a product redesign**.

---

## Phase 1: Discovery (Required — No Code Changes)

### 1) System map: architecture + boundaries

Identify:

- Frontend: framework, routing, state management, UI/component library, data-fetching approach
- Backend: framework, API style (REST/GraphQL/RPC), auth/session model, background jobs/queues
- Data: DB, ORM, migrations strategy, indexing approach, caching, file storage/CDN
- Contracts: how FE ↔ BE communicate (auth headers/cookies, error shapes, retries, pagination, validation)
  Output a concise **system map** diagram in text + bullet form.

### 2) Existing tests + CI gates

Find and summarize:

- Unit/integration/e2e setup (tools, configs, locations)
- Mocks/fixtures/factories, test DB strategy, seed data
- CI pipelines and gates: lint, typecheck, tests, build, migrations checks
- Any gaps (e.g., no contract tests, no DB integration tests, no e2e smoke)

### 3) Security posture + risk surface

Audit:

- Authentication lifecycle: login/logout/refresh, token storage, session expiry
- Authorization enforcement: server-side checks vs frontend-only gating
- Input validation + output encoding, file upload safety, SSRF/path traversal risks
- CORS/CSRF, rate limiting, secrets/config handling, dependency risk (SCA)
- Tenant isolation (if multi-tenant), IDOR risks, least privilege in backend
  List concrete attack surfaces with severity.

### 4) Performance + scalability characteristics

Assess:

- Frontend: bundle size, route splitting, hydration/render cost, re-render hotspots, network waterfalls
- Backend: slow endpoints, N+1 queries, missing indexes, payload sizes, timeouts, concurrency limits
- Caching layers: client/server/CDN, cache headers, invalidation strategy
- DB health: query plans, connection pooling, migrations safety

### 5) UX/UI flows + accessibility

Review key journeys and failure modes:

- Loading/empty/error states and recovery paths
- Form validation + mapping server errors to fields
- Navigation consistency, deep links, back/forward correctness
- A11y: keyboard nav, focus management, labels, contrast, ARIA correctness

### 6) Code quality + maintainability

Identify:

- Lint/format/typecheck configuration and adherence
- Duplication, oversized modules, inconsistent patterns
- Error handling consistency, logging/observability patterns
- Dead code, unused dependencies, risky utilities/helpers

### 7) Unfinished work + footguns

Scan for:

- TODO/FIXME/HACK/XXX, commented-out code, partial implementations
- Placeholder UI, disabled features, feature flags/config drift
- Inconsistent environment branching and “works on my machine” assumptions

### 8) Missing frontend features (Feature-Gap Audit — bounded)

Identify **missing UX-critical frontend features** that commonly cause real user pain or support load, such as:

- Error boundaries and “retry” UX for failed queries
- Global toasts/notifications for action success/failure
- Empty states with next-step guidance
- Pagination/infinite scroll where needed + consistent table patterns
- Consistent form validation + server error mapping
- Offline/slow-network handling indicators
- Basic accessibility fixes (focus traps, keyboard support)
- Loading skeletons/spinners aligned to the design system
- Session expiry handling (re-auth flow) and “logged out” recovery
- Basic settings/account screens if required for security/compliance (only if truly missing)

**Rule:** Only include feature gaps that are high-impact, low/medium effort, and reduce bugs/support burden.

✅ Phase 1 Output must include:

1. System map
2. Test + CI inventory
3. Security risk list
4. Perf bottlenecks list
5. UX/a11y findings
6. Code quality findings
7. Footguns list
8. Feature-gap candidates
9. A single combined list of candidate improvements (no code changes)

---

## Phase 2: Improvement Categories (Balanced)

For each candidate improvement, tag one or more:

- Security (Critical)
- Performance & Speed
- Reliability & Observability
- UX (flows, recovery, validation)
- UI (consistency, responsiveness, theming if applicable)
- Testing (critical paths, contracts, failure modes)
- Code Quality (duplication, patterns, maintainability)
- Architecture (boundaries, contracts/types, layering)
- Storage Efficiency (schema/index/log retention/cache correctness)
- Unfinished Work (remove/resolve TODOs, stubs, dead code)
- Missing Frontend Features (bounded, impact-first)

---

## Phase 3: Prioritized Recommendations (Plan First — No Code Yet)

Rank by **Impact × Likelihood × Effort efficiency** (not by category).

### 🎯 Top 5 High-Impact Improvements (Overall)

For each provide:

- Category tag(s)
- Issue: what’s wrong and why it matters
- Impact: High/Medium/Low (user/system/business effects)
- Effort: Small/Medium/Large
- Risk: breaking-change potential + mitigation plan
- Success criteria: measurable outcomes (perf, error rate, UX metrics, security posture)
- Verification plan: tests + manual checks + tooling
- Exact file targets / areas of code to touch

### ⚡ 3 Quick Wins (<15 minutes each)

Must be:

- Low risk
- Clearly valuable
- Includes exact file targets and minimal-diff approach

**Stop here and ask for approval** before implementing anything.

---

## Phase 4: Implementation Rules (After Approval)

- Read files before changing; validate assumptions in code.
- One improvement at a time; keep diffs tight.
- Reuse existing patterns, libraries, and test tooling.
- Add tests where regressions are likely and impact is high.
- Update/extend test scaffolding only when necessary (don’t overbuild).
- Run lint/typecheck/tests after each change (or nearest equivalent).
- Do not expand scope into unrelated features.

---

## Phase 5: Deliverables Per Improvement

For each implemented improvement, provide:

- What was wrong + why it mattered
- What changed (file references + summary)
- How it was verified (tests + manual steps)
- Any follow-ups (backlog only unless explicitly approved)

---

## Standard of Quality

- Follow industry best practices for security, testing, and maintainability.
- Prefer high-confidence wins first.
- If you see additional issues while working, log them as backlog unless they are directly required for the approved change.
- “Create tests for everything essential”: prioritize tests that protect critical user flows, security boundaries, and FE↔BE contracts.
