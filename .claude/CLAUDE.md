# React + Go Starter Kit - Claude Code Guide

## Project Overview

Full-stack SaaS starter with React 19 + Go 1.25. Multi-tenant, WebSocket, Stripe billing.

> **Note:** Global skills, output styles, and automation hooks are documented in `~/.claude/CLAUDE.md`. This file covers project-specific patterns only.

## Quick Start

```bash
cp .env.example .env
cp .claude/settings.local.json.template .claude/settings.local.json
make dev
```

## Directory Structure

```
backend/
├── cmd/main.go           # Entry point, route setup
├── internal/
│   ├── auth/             # JWT, OAuth, 2FA
│   ├── handlers/         # HTTP request handlers
│   ├── services/         # Business logic layer
│   ├── repository/       # Data access interfaces
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
│   │   ├── guards.ts     # Route guards
│   │   └── optimistic-mutations.ts # Optimistic update helpers
│   ├── stores/           # Zustand stores (UI state)
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

### Optimistic Updates

Use factory functions for consistent optimistic mutation handling:

```typescript
// For deletes (optimistic-mutations.ts)
createOptimisticDeleteHandlers<User, number>({
  queryClient,
  listQueryKey: queryKeys.users.lists(),
  getId: (user) => user.id,
  successMessage: "User deleted",
});

// For updates (optimistic-updates.ts)
createOptimisticUpdate<User, UpdateUserInput>({
  queryClient,
  queryKey: queryKeys.users.detail(id),
  updateFn: (old, input) => ({ ...old, ...input }),
});

// Specialized helpers:
createListOptimisticUpdate()    // Add/remove/update items in lists
createToggleOptimisticUpdate()  // Boolean field toggles
createCounterOptimisticUpdate() // Increment/decrement counters
```

### Zustand + TanStack Query Integration

Zustand stores hold **UI state**, TanStack Query holds **server state**.

```typescript
// Mutations can update Zustand on success:
const resetForm = useUserStore((state) => state.resetForm);

onSuccess: () => {
  queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
  resetForm();  // Reset Zustand UI state
}
```

Store responsibilities:
- `auth-store`: User auth state, session initialization
- `user-store`: Selected user, filters, form data
- `org-store`: Selected org, org-level UI state
- `notification-store`: Toast queue, notification preferences

### API Client Circuit Breaker

The API client implements a circuit breaker for 401 handling:
- After 3 consecutive 401 failures, stops retrying for 10s
- Resets after successful auth
- Prevents infinite refresh loops

```typescript
// After successful login:
resetAuthCircuitBreaker()
markAuthenticationComplete()  // 5s grace period for cookie propagation
```

### Backend Services (business logic layer)

```go
// Services handle business logic, handlers call services
type OrgService struct { db *gorm.DB }
func (s *OrgService) GetBySlug(ctx context.Context, slug string) (*Organization, error)
```

### Repository Pattern

Services depend on repository interfaces for testability:

```go
// backend/internal/repository/interfaces.go
type SessionRepository interface {
    Create(ctx context.Context, session *models.UserSession) error
    FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error)
}

// GORM implementations in repository/*.go
// Mock implementations in testutil/mocks/repository.go
```

Available repositories:
- SessionRepository, LoginHistoryRepository
- UserRepository, OrganizationRepository, OrganizationMemberRepository
- SubscriptionRepository, SystemSettingRepository, IPBlocklistRepository
- UsageEventRepository, UsagePeriodRepository, UsageAlertRepository

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
6. **Circuit Breaker**: After 3 failed 401s, API client stops retrying for 10s
7. **Auth Grace Period**: 5s grace after login before 401s trigger session-expired
8. **Optimistic Updates**: Use factory functions in `lib/optimistic-*.ts` for consistency
9. **Repository Pattern**: Services depend on repository interfaces for testability
10. **Zustand Stores**: UI state only - server state lives in TanStack Query
11. **Test Transactions**: Always use `NewTestTransaction` for automatic rollback

## Testing

- **Frontend**: `npm run test` (Vitest)
- **Backend Unit**: `go test ./...` or `make test-backend`
- **Backend Integration**: `INTEGRATION_TEST=true go test ./...`
- **E2E**: `npm run test:e2e` (Playwright)

### Backend Integration Tests

Integration tests use real PostgreSQL via testcontainers-go or a pre-configured test database.

**Location:** `backend/internal/services/*_integration_test.go`

**Available test suites:**
- `org_service_integration_test.go` - Organization CRUD, memberships, invitations
- `session_service_integration_test.go` - Session management, device tracking
- `settings_service_integration_test.go` - System settings, batched updates
- `usage_service_integration_test.go` - Usage tracking, limits, alerts
- `totp_service_integration_test.go` - 2FA setup and verification
- `file_service_integration_test.go` - File operations with S3/LocalStack
- `health_service_integration_test.go` - Health check endpoint testing
- `user_preferences_service_integration_test.go` - User preferences CRUD

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
TEST_DB_HOST=localhost TEST_DB_PORT=5434 \
INTEGRATION_TEST=true go test ./internal/services/... -v
```

### Test Utilities (`backend/internal/testutil/`)

| File | Purpose |
|------|---------|
| `database.go` | SetupTestDB, NewTestTransaction, TruncateTables, WithTestDB |
| `containers.go` | SetupPostgresContainer, SetupRedisContainer, TestInfrastructure |
| `fixtures.go` | UserFactory, test data builders |
| `seeder.go` | Database seeding for integration tests |
| `http.go` | HTTP test helpers (request builders, response assertions) |
| `mocks/*.go` | Mock implementations (Stripe, S3, Email, Repository) |

**Test transaction pattern (recommended):**

```go
func TestSomething(t *testing.T) {
    testutil.SkipIfNotIntegration(t)
    db := testutil.SetupTestDB(t)
    tt := testutil.NewTestTransaction(t, db)
    defer tt.Rollback()  // Auto-rollback on cleanup

    svc := NewService(tt.DB)
    // Test with tt.DB...
}
```

**Container-based testing (for CI):**

```go
func TestWithContainers(t *testing.T) {
    pg := testutil.SetupPostgresContainer(t)  // Auto-cleanup
    testutil.WithPostgresDB(t, pg, func(db *gorm.DB) {
        // Test with isolated DB
    })
}
```

**Quick helpers:**
- `testutil.CreateTestUser(t, db, opts...)`
- `testutil.CreateTestOrganization(t, db, name, ownerID)`
- `testutil.CreateTestOrgMember(t, db, orgID, userID, role)`
- `testutil.TruncateTables(t, db)` - Full reset
- `testutil.CleanupTestData(t, db, "test@%")` - Pattern-based cleanup

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
| Optimistic updates | `frontend/app/lib/optimistic-mutations.ts` |
| Auth store | `frontend/app/stores/auth-store.ts` |
| Main routes | `backend/cmd/main.go` |
| Models | `backend/internal/models/models.go` |
| Auth middleware | `backend/internal/auth/middleware.go` |
| Repository interfaces | `backend/internal/repository/interfaces.go` |
| Test utilities | `backend/internal/testutil/database.go` |

---

## Project Memory

Persistent context across sessions at `.claude/memory/memory.md`.

**Usage:**
- Add notes: `.claude/hooks/memory-sync.sh add-note "Your note"`
- Add learnings: `.claude/hooks/memory-sync.sh add-learning "What you learned"`
- Add decisions: `.claude/hooks/memory-sync.sh add-decision "Decision made"`

---

## Architecture Decision Records

Decisions tracked in `.claude/decisions/` using MADR format.

**Index:** [.claude/decisions/index.md](.claude/decisions/index.md)

**Creating a new ADR:**
1. Copy template: `cp .claude/decisions/0000-template.md .claude/decisions/0001-my-decision.md`
2. Fill in the template
3. Index auto-updates via `adr-index.sh` hook
4. Commit: `git commit -m "docs(adr): add ADR-0001 my decision"`

---

## Project Skills

### `/scaffold-service` - Go Service Scaffolding

```bash
/scaffold-service
```

Creates:
- `backend/internal/services/{entity}_service.go` - Service with sentinel errors
- `backend/internal/services/{entity}_service_test.go` - Table-driven tests
- `backend/internal/handlers/{entity}_handlers.go` - HTTP handlers
- Route registration snippet for `cmd/main.go`

### `/scaffold-feature` - Full-Stack Feature

```bash
/scaffold-feature
```

Creates all backend files plus:
- `frontend/app/types/{entity}.ts` - TypeScript types
- `frontend/app/services/{entity}/` - API service
- `frontend/app/hooks/queries/use-{entity}.ts` - Query hooks
- `frontend/app/hooks/mutations/use-{entity}-mutations.ts` - Mutation hooks

### `/add-migration` - Database Migration Generator

```bash
/add-migration
```

Features: 000XXX naming convention, .up.sql/.down.sql pairs, optional GORM model updates.

### `/sync-models` - Go to TypeScript Type Sync

```bash
/sync-models
```

Parses Go GORM structs, generates TypeScript interfaces, detects drift.

### `/env-check` - Environment Validation

```bash
/env-check
```

Validates .env, Docker containers, database/cache connectivity.

### `/debug-api` - API Testing Helper

```bash
/debug-api
```

Generates curl commands with JWT auth and CSRF tokens.

---

## Project Hooks

Pattern validation hooks run automatically. See [.claude/hooks/README.md](.claude/hooks/README.md) for details.

| Hook | Trigger | Validates |
|------|---------|-----------|
| `go-service-pattern.sh` | `*_service.go` | Sentinel errors, constructor, context usage |
| `query-key-pattern.sh` | `query-keys.ts` | Factory pattern, as const, spread usage |
| `hook-naming.sh` | `hooks/**/*.ts` | use{Entity} / use{Action}{Entity} naming |
| `memory-sync.sh` | SessionStart/Stop | Loads/saves project memory |
| `adr-index.sh` | `decisions/*.md` | Auto-generates ADR index |

---

## Project Prompts

| Prompt | Use For |
|--------|---------|
| `stack-review.md` | Architecture review for React+Go |
| `feature-checklist.md` | Full-stack feature completion checklist |
| `security-audit.md` | Security review specific to this stack |
| `api-contract-review.md` | Go <-> TypeScript type alignment |
| `deployment-checklist.md` | Pre-production readiness checklist |

---

## Project Agent

### `fullstack-reviewer`

Specialized code reviewer for React + Go patterns, API contracts, and security.

---

## MCP Servers

| Server | Purpose | Env Var |
|--------|---------|---------|
| `db` | PostgreSQL access via dbhub | `DB_DSN` |
| `github` | GitHub API for PRs/issues | `GITHUB_TOKEN` |
| `memory` | Persistent memory across sessions | - |
