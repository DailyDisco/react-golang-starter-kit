# React + Go Starter Kit - Claude Code Guide

## Project Overview

Full-stack SaaS starter with React 19 + Go 1.25. Multi-tenant, WebSocket, Stripe billing.

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
