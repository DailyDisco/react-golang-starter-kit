# React + Go Starter Kit - Claude Code Guide

## Project Overview

Full-stack SaaS starter with React 19 + Go 1.25. Multi-tenant, WebSocket, Stripe billing.

## Skill Auto-Invocation Rules

**IMPORTANT:** When the user's request matches these patterns, invoke the corresponding skill automatically:

| User Intent | Invoke Skill |
|-------------|--------------|
| "Add a feature", "new feature", "implement X feature" | `/scaffold-feature` |
| "Add a service", "new service", "backend for X" | `/scaffold-service` |
| "Add migration", "new table", "add column", "change schema" | `/add-migration` |
| "Sync types", "update types", "types are out of sync" | `/sync-models` |
| "Review architecture", "review this design" | Load `prompts/stack-review.md` |
| "Security review", "check for vulnerabilities" | Load `prompts/security-audit.md` |
| "Is feature complete?", "what am I missing?" | Load `prompts/feature-checklist.md` |

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
- **Backend**: `go test ./...` or `make test-backend`
- **E2E**: `npm run test:e2e` (Playwright)

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

---

## Project Prompts

Load these prompts for specialized analysis:

| Prompt | Use For |
|--------|---------|
| `stack-review.md` | Architecture review for React+Go |
| `feature-checklist.md` | Full-stack feature completion checklist |
| `security-audit.md` | Security review specific to this stack |

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

## File Structure

```
.claude/
├── CLAUDE.md                          # This file
├── settings.local.json.template       # Settings template
├── skills/
│   ├── scaffold-service/              # Go service scaffolding
│   ├── scaffold-feature/              # Full-stack feature scaffolding
│   ├── add-migration/                 # Database migration generator
│   └── sync-models/                   # Go → TypeScript sync
├── hooks/
│   ├── go-service-pattern.sh          # Go service validation
│   ├── query-key-pattern.sh           # Query key validation
│   └── hook-naming.sh                 # Hook naming validation
├── prompts/
│   ├── stack-review.md                # Architecture review
│   ├── feature-checklist.md           # Feature completion checklist
│   └── security-audit.md              # Security review
└── agents/
    └── fullstack-reviewer/            # Stack-aware code reviewer
```
