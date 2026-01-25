# React + Go Starter Kit - Claude Code Guide

## Project Overview

Full-stack SaaS starter with React 19 + Go 1.25. Multi-tenant, WebSocket, Stripe billing.

## Automatic Skill Invocation (Jarvis Mode)

When you express intent, Claude automatically invokes the right skill:

| You Say | Claude Does |
|---------|-------------|
| "Add a user profile feature" | `/scaffold-feature` → Full backend + frontend |
| "Create an auth service" | `/scaffold-service` → Go service with tests |
| "Add a notifications table" | `/add-migration` → SQL migration files |
| "Types are out of sync" | `/sync-models` → Go → TypeScript sync |
| "Debug the API" | `/debug-api` → curl commands |
| "Check my environment" | `/env-check` → Validate setup |

### Global Workflows (Also Auto-Suggested)

| You Say | Claude Does |
|---------|-------------|
| "Fix this bug" | `/workflow bugfix` → Debug → Fix → Test |
| "Review this PR" | `/review-pr` → Security + Performance + Quality |
| "Prepare release" | `/workflow release` → Changelog → Tag → PR |
| "Refactor this" | `/workflow refactor` → Plan → Execute → Verify |

### Specialized Agents (Auto-Routed)

When domain expertise is needed, Claude spawns specialists:

| Domain | Agent | Triggers |
|--------|-------|----------|
| Database | `@db-specialist` | Schema, migrations, queries |
| Security | `@security-auditor` | Auth, vulnerabilities, OWASP |
| Frontend | `@frontend-reviewer` | React, a11y, performance |
| API | `@api-specialist` | REST design, contracts |

### Analysis Prompts

| You Say | Claude Loads |
|---------|--------------|
| "Review architecture" | `prompts/stack-review.md` |
| "Security review" | `prompts/security-audit.md` |
| "Is feature complete?" | `prompts/feature-checklist.md` |
| "Check API contract" | `prompts/api-contract-review.md` |
| "Ready to deploy?" | `prompts/deployment-checklist.md` |

**After completing scaffolding tasks**, always remind the user of next steps (e.g., "Run `make migrate-up` to apply the migration").

## Directory Structure

```
backend/
├── cmd/main.go           # Entry point, route setup
├── internal/
│   ├── auth/             # JWT, OAuth, 2FA
│   ├── handlers/         # HTTP request handlers
│   ├── services/         # Business logic layer
│   ├── models/           # GORM models
│   ├── middleware/       # Auth, rate limiting, security
│   ├── cache/            # Redis/Dragonfly caching
│   ├── stripe/           # Billing integration
│   └── websocket/        # Real-time events
└── migrations/           # SQL migrations

frontend/
├── app/
│   ├── hooks/
│   │   ├── queries/      # TanStack Query hooks
│   │   └── mutations/    # Mutation hooks
│   ├── lib/
│   │   ├── query-keys.ts # Query key factory
│   │   └── guards.ts     # Route guards
│   ├── routes/           # File-based routing
│   └── services/api/     # API client
└── components/ui/        # ShadCN components
```

## Code Patterns

### Frontend Query Keys (MUST use this pattern)

```typescript
// frontend/app/lib/query-keys.ts
queryKeys.users.detail(id)       // ["users", "detail", id]
queryKeys.billing.subscription() // ["billing", "subscription"]
queryKeys.organizations.members(orgSlug) // ["organizations", orgSlug, "members"]
```

### Frontend Mutations (invalidate related queries)

```typescript
onSuccess: () => {
  queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
}
```

### Backend Services (business logic layer)

```go
// Services handle business logic, handlers call services
type OrgService struct { db *gorm.DB }
func (s *OrgService) GetBySlug(ctx context.Context, slug string) (*Organization, error)
```

### Backend Handlers (thin, call services)

```go
func (h *OrgHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
    org, err := h.service.GetBySlug(r.Context(), chi.URLParam(r, "orgSlug"))
    // Handle error, return JSON
}
```

### Database Queries (ALWAYS use context)

```go
s.db.WithContext(ctx).Where("slug = ?", slug).First(&org)
```

## Common Gotchas

1. **N+1 Queries**: Use `.Preload()` for associations
2. **Cache Keys**: Use `strconv.FormatUint()` for IDs, NOT `string(rune())`
3. **WebSocket**: Messages broadcast via `hub.Broadcast(type, payload)`
4. **Auth**: JWT in cookie, refresh token rotation enabled
5. **API Versioning**: Routes mounted at both `/api/v1` and `/api`

## Testing

- **Frontend**: `npm run test` (Vitest)
- **Backend Unit**: `go test ./...` or `make test-backend`
- **Backend Integration**: `INTEGRATION_TEST=true go test ./...`
- **E2E**: `npm run test:e2e` (Playwright)

### Backend Integration Tests

Integration tests use real PostgreSQL via testcontainers-go or a pre-configured test database. They test database-dependent code paths that unit tests skip.

**Location:** `backend/internal/services/*_integration_test.go`

**Available integration test suites:**

- `org_service_integration_test.go` - Organization CRUD, memberships, invitations
- `session_service_integration_test.go` - Session management, device tracking
- `settings_service_integration_test.go` - System settings, batched updates
- `usage_service_integration_test.go` - Usage tracking, limits, alerts
- `totp_service_integration_test.go` - 2FA setup and verification
- `file_service_integration_test.go` - File operations with S3/LocalStack

**Running integration tests:**

```bash
# Option 1: Start test database container
docker run -d --name starter-test-db \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -e POSTGRES_DB=starter_kit_test \
  -p 5433:5432 \
  postgres:16-alpine

# Run integration tests
cd backend && INTEGRATION_TEST=true go test ./internal/services/... -v

# Option 2: Use custom test database
TEST_DB_HOST=localhost \
TEST_DB_PORT=5434 \
TEST_DB_USER=postgres \
TEST_DB_PASSWORD=mypass \
TEST_DB_NAME=test_db \
INTEGRATION_TEST=true go test ./internal/services/... -v
```

**Test utilities:** `backend/internal/testutil/`

| File | Purpose |
| ---- | ------- |
| `database.go` | Test DB setup, transactions, migrations |
| `containers.go` | Testcontainers-go for PostgreSQL/Redis |
| `mocks/repository.go` | Mock repositories for unit tests |

**Test transaction pattern:**

```go
func testSetup(t *testing.T) (*Service, func()) {
    testutil.SkipIfNotIntegration(t)
    db := testutil.SetupTestDB(t)
    tt := testutil.NewTestTransaction(t, db)
    svc := NewService(tt.DB)
    return svc, func() { tt.Rollback() }
}
```

## Commands

```bash
make dev              # Start all services
make test             # Run all tests
make migrate-up       # Apply migrations
make seed             # Seed test data
```

## Key Files

| Purpose | File |
|---------|------|
| Query keys | `frontend/app/lib/query-keys.ts` |
| API client | `frontend/app/services/api/client.ts` |
| Route guards | `frontend/app/lib/guards.ts` |
| Main routes | `backend/cmd/main.go` |
| Models | `backend/internal/models/models.go` |
| Auth middleware | `backend/internal/auth/middleware.go` |

---

## Project Memory

This project uses `.claude/memory/memory.md` for persistent context across sessions.

**Location:** [.claude/memory/memory.md](.claude/memory/memory.md)

**Features:**

- Automatically loaded at session start (via `memory-sync.sh` hook)
- Timestamp updated at session end
- Version controlled - committed to git

**Usage:**

- Add notes: `.claude/hooks/memory-sync.sh add-note "Your note"`
- Add learnings: `.claude/hooks/memory-sync.sh add-learning "What you learned"`
- Add decisions: `.claude/hooks/memory-sync.sh add-decision "Decision made"`

---

## Architecture Decision Records

Decisions are tracked in `.claude/decisions/` using MADR format.

**Index:** [.claude/decisions/index.md](.claude/decisions/index.md)

**Creating a new ADR:**

1. Copy template: `cp .claude/decisions/0000-template.md .claude/decisions/0001-my-decision.md`
2. Fill in the template (Status, Context, Decision, Consequences)
3. Index auto-updates on save (via `adr-index.sh` hook)
4. Commit: `git commit -m "docs(adr): add ADR-0001 my decision"`

**Status Lifecycle:**

- **Proposed** - Under discussion
- **Accepted** - Approved and in effect
- **Deprecated** - No longer applies
- **Superseded** - Replaced by another ADR

---

## Project Skills

This project includes Claude Code skills for scaffolding new features.

### `/scaffold-service` - Go Service Scaffolding

Generates a complete Go service layer following project patterns:

```bash
/scaffold-service
```

**Creates:**
- `backend/internal/services/{entity}_service.go` - Service with sentinel errors
- `backend/internal/services/{entity}_service_test.go` - Table-driven tests
- `backend/internal/handlers/{entity}_handlers.go` - HTTP handlers
- Route registration snippet for `cmd/main.go`

### `/scaffold-feature` - Full-Stack Feature

Generates end-to-end feature spanning backend and frontend:

```bash
/scaffold-feature
```

**Creates:**
- All backend files (via scaffold-service)
- `frontend/app/types/{entity}.ts` - TypeScript types
- `frontend/app/services/{entity}/` - API service
- `frontend/app/hooks/queries/use-{entity}.ts` - Query hooks
- `frontend/app/hooks/mutations/use-{entity}-mutations.ts` - Mutation hooks
- Updates to `query-keys.ts` and index exports

---

## Project Hooks

Pattern validation hooks run automatically when editing relevant files.

| Hook | Trigger | Validates |
|------|---------|-----------|
| `go-service-pattern.sh` | `*_service.go` | Sentinel errors, constructor, context usage |
| `query-key-pattern.sh` | `query-keys.ts` | Factory pattern, as const, spread usage |
| `hook-naming.sh` | `hooks/**/*.ts` | use{Entity} / use{Action}{Entity} naming |

### Hook Configuration

To enable hooks, copy the settings template:

```bash
cp .claude/settings.local.json.template .claude/settings.local.json
```

Or add manually to your Claude Code settings:

```json
{
  "hooks": {
    "PostEdit": [
      {
        "matcher": "backend/internal/services/*_service.go",
        "command": ".claude/hooks/go-service-pattern.sh \"$FILE\" \"$CONTENT\""
      },
      {
        "matcher": "frontend/app/lib/query-keys.ts",
        "command": ".claude/hooks/query-key-pattern.sh \"$FILE\" \"$CONTENT\""
      },
      {
        "matcher": "frontend/app/hooks/**/*.ts",
        "command": ".claude/hooks/hook-naming.sh \"$FILE\" \"$CONTENT\""
      }
    ]
  }
}
```

---

## Additional Skills

### `/add-migration` - Database Migration Generator

Generate safe, reversible database migrations following project conventions:

```bash
/add-migration
```

**Features:**
- Uses project's 000XXX naming convention
- Knows existing table structure (20+ tables)
- Generates both .up.sql and .down.sql
- Optionally updates GORM models

### `/sync-models` - Go to TypeScript Type Sync

Keep frontend types aligned with backend models:

```bash
/sync-models
```

**Features:**
- Parses Go GORM structs
- Generates TypeScript interfaces
- Detects drift between Go and TS types
- Preserves custom type extensions

### `/env-check` - Environment Validation

Validate environment setup before starting development:

```bash
/env-check
```

**Features:**
- Checks `.env` file exists and has required vars
- Validates formats (JWT_SECRET length, DB_PORT numeric, etc.)
- Verifies Docker containers are running
- Tests database and cache connectivity

### `/debug-api` - API Testing Helper

Generate curl commands for testing API endpoints:

```bash
/debug-api
```

**Features:**
- Generates curl commands with proper JWT cookie auth
- Handles CSRF tokens for mutations
- Shows expected response shapes
- Includes troubleshooting tips

---

## Project Prompts

Load these prompts for specialized analysis:

| Prompt | Use For |
|--------|---------|
| `stack-review.md` | Architecture review for React+Go |
| `feature-checklist.md` | Full-stack feature completion checklist |
| `security-audit.md` | Security review specific to this stack |
| `api-contract-review.md` | Go ↔ TypeScript type alignment |
| `deployment-checklist.md` | Pre-production readiness checklist |

Usage: Reference in conversation or use `/prompt` skill.

---

## Project Agent

### `fullstack-reviewer`

A specialized code reviewer that understands both React and Go patterns in this project.

**Expertise:**
- TanStack Query/Router patterns
- Go Chi/GORM patterns
- API contract alignment
- Security best practices

---

## MCP Servers

Pre-configured MCP servers in `.mcp.json`:

| Server | Purpose | Env Var |
|--------|---------|---------|
| `db` | PostgreSQL access via dbhub | `DB_DSN` |
| `github` | GitHub API for PRs/issues | `GITHUB_TOKEN` |
| `memory` | Persistent memory across sessions | - |

---

## Quick Start for New Developers

1. **Clone and setup:**
   ```bash
   git clone <repo>
   cd react-golang-starter-kit
   cp .env.example .env
   cp .claude/settings.local.json.template .claude/settings.local.json
   ```

2. **Start development:**
   ```bash
   make dev
   ```

3. **Add a new feature:**
   ```bash
   /scaffold-feature
   ```

4. **Add a migration:**
   ```bash
   /add-migration
   ```

---

## Output Styles

Switch interaction modes with `/output-style`:

| Style | Best For |
|-------|----------|
| `teaching` | Learning step-by-step with exercises |
| `executive` | Quick business summaries |
| `minimal` | Code only, no explanation |
| `pair-programming` | Collaborative thinking out loud |
| `debugging` | Systematic hypothesis-driven investigation |

---

## Quick Commands

Lightweight commands (no workflow overhead):

| Command | Purpose |
|---------|---------|
| `/explain [target]` | Explain code or architecture |
| `/quick-fix [issue]` | Simple targeted fixes |
| `/compare [A] vs [B]` | Compare approaches |
| `/benchmark [desc]` | Performance comparison |

---

## Advanced Automation Hooks

These global hooks run automatically (enabled in `settings.local.json`):

| Hook | Trigger | Action |
|------|---------|--------|
| `skill-suggester` | Every prompt | Suggests relevant skills based on intent |
| `agent-router` | Every prompt | Routes to specialist agents automatically |
| `post-edit-dispatch` | File edits | Runs type checks, security scans |
| `auto-test` | File edits | Finds and suggests relevant tests |
| `build-test-gate` | Before commit | Parallel validation (types, tests, lint) |
| `commit-guard` | Before commit | Validates conventional commit format |

### What Happens Automatically

1. **When you describe a task** → Skill suggestions appear
2. **When you edit code** → Type checks and test suggestions run
3. **When you commit** → Build, tests, and lint validate in parallel
4. **When domain expertise needed** → Specialist agents are suggested

---

## File Structure

```
.claude/
├── CLAUDE.md                          # This file
├── settings.local.json.template       # Settings template
├── memory/
│   └── memory.md                      # Persistent session context
├── decisions/
│   ├── 0000-template.md               # MADR template
│   └── index.md                       # Auto-generated index
├── skills/
│   ├── scaffold-service/              # Go service scaffolding
│   ├── scaffold-feature/              # Full-stack feature scaffolding
│   ├── add-migration/                 # Database migration generator
│   ├── sync-models/                   # Go → TypeScript sync
│   ├── env-check/                     # Environment validation
│   └── debug-api/                     # API testing helper
├── hooks/
│   ├── memory-sync.sh                 # SessionStart/Stop memory hook
│   ├── adr-index.sh                   # ADR auto-indexing
│   ├── go-service-pattern.sh          # Go service validation
│   ├── query-key-pattern.sh           # Query key validation
│   └── hook-naming.sh                 # Hook naming validation
├── prompts/
│   ├── stack-review.md                # Architecture review
│   ├── feature-checklist.md           # Feature completion checklist
│   ├── security-audit.md              # Security review
│   ├── api-contract-review.md         # Go ↔ TypeScript alignment
│   └── deployment-checklist.md        # Pre-production checklist
└── agents/
    └── fullstack-reviewer/            # Stack-aware code reviewer
```
