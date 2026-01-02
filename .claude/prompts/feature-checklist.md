# Full-Stack Feature Checklist

Use this checklist when implementing a complete feature spanning backend and frontend.

---

## Pre-Implementation

```
[ ] Requirements are clear and documented
[ ] API contract is defined (endpoints, request/response shapes)
[ ] Database schema changes identified
[ ] Security implications reviewed
[ ] Existing patterns identified to follow
```

---

## Backend Implementation

### Database Layer

```
[ ] Migration created (up + down)
[ ] Migration tested (up → down → up)
[ ] Indexes added for query patterns
[ ] Foreign keys with appropriate ON DELETE
[ ] GORM model created/updated
```

### Service Layer

```
[ ] Service struct created with db field
[ ] Constructor follows New{Entity}Service pattern
[ ] Sentinel errors defined (ErrNotFound, etc.)
[ ] Methods accept context.Context as first param
[ ] Errors wrapped with fmt.Errorf context
[ ] Business logic is in service (not handler)
[ ] Transactions used for multi-step operations
```

### Handler Layer

```
[ ] Handler struct with service dependency
[ ] Input validation (request body, params)
[ ] Proper HTTP status codes
[ ] Consistent error response format
[ ] Routes registered in main.go
[ ] Authentication middleware applied
[ ] Authorization checks (RBAC)
```

### Testing

```
[ ] Unit tests for service methods
[ ] Table-driven tests with edge cases
[ ] Integration tests for handlers (if critical)
[ ] Test coverage for error paths
```

---

## Frontend Implementation

### Types

```
[ ] TypeScript interface matches Go model
[ ] Request types (Create, Update)
[ ] Response types (Single, List with meta)
[ ] Types exported from index
```

### Query Keys

```
[ ] Added to queryKeys factory
[ ] Follows pattern: all → lists → list → details → detail
[ ] Uses 'as const' assertion
```

### API Service

```
[ ] Service file created in services/{entity}/
[ ] Methods for all CRUD operations
[ ] Proper typing on requests and responses
[ ] Error handling (ApiError)
[ ] Exported from index
```

### Query Hooks

```
[ ] useQuery hook for fetching
[ ] Uses queryKeys from factory
[ ] Proper enabled condition
[ ] Loading/error states accessible
[ ] Exported from hooks/queries/index
```

### Mutation Hooks

```
[ ] useMutation hooks for write operations
[ ] Invalidates correct queries on success
[ ] Shows toast on success/error
[ ] Updates optimistically (if appropriate)
[ ] Exported from hooks/mutations/index
```

### UI Components

```
[ ] Loading skeleton
[ ] Error state display
[ ] Empty state display
[ ] Form validation (Zod + react-hook-form)
[ ] Accessible (keyboard, screen reader)
[ ] Responsive design
```

### Testing

```
[ ] Hook tests with mock service
[ ] Component tests (if complex logic)
[ ] Verify cache invalidation works
```

---

## Integration

```
[ ] Backend builds: go build ./...
[ ] Backend tests pass: go test ./...
[ ] Frontend typechecks: npm run typecheck
[ ] Frontend tests pass: npm run test
[ ] E2E test for critical path (if applicable)
[ ] Manual testing in development
```

---

## Documentation

```
[ ] API endpoint documented (OpenAPI or README)
[ ] CLAUDE.md updated (if new pattern)
[ ] Migration notes (if schema change)
[ ] Feature flag (if gradual rollout)
```

---

## Deployment

```
[ ] Environment variables documented
[ ] Migration runs before app deploy
[ ] Feature flag configured (if using)
[ ] Monitoring/alerts set up
[ ] Rollback plan documented
```

---

## Quick Reference

### File Locations

| What | Where |
|------|-------|
| Migration | `backend/migrations/000XXX_{name}.up.sql` |
| Go Model | `backend/internal/models/models.go` |
| Go Service | `backend/internal/services/{entity}_service.go` |
| Go Handler | `backend/internal/handlers/{entity}_handlers.go` |
| Routes | `backend/cmd/main.go` |
| TS Types | `frontend/app/types/{entity}.ts` |
| API Service | `frontend/app/services/{entity}/` |
| Query Keys | `frontend/app/lib/query-keys.ts` |
| Query Hooks | `frontend/app/hooks/queries/use-{entity}.ts` |
| Mutations | `frontend/app/hooks/mutations/use-{entity}-mutations.ts` |

### Commands

```bash
# Backend
make migrate-up          # Apply migrations
go build ./...           # Build
go test ./...            # Test

# Frontend
npm run typecheck        # Type check
npm run test            # Test
npm run dev             # Dev server

# Full stack
make dev                # Start everything
make test               # Test everything
```
