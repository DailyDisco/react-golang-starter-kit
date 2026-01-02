# React + Go Stack Architecture Review

Use this prompt when reviewing architecture decisions, refactoring plans, or evaluating new feature designs in this stack.

## Stack Context

**Frontend:** React 19 + TanStack Router + TanStack Query + Zustand + Tailwind + ShadCN
**Backend:** Go 1.25 + Chi + GORM + PostgreSQL + JWT + WebSocket
**Infrastructure:** Docker + GitHub Actions + Caddy

---

## Review Checklist

### 1. Data Flow Architecture

```
[ ] Query keys follow factory pattern (queryKeys.entity.detail(id))
[ ] Mutations invalidate correct queries on success
[ ] Optimistic updates used where appropriate
[ ] Loading/error states handled consistently
[ ] Cache invalidation strategy is clear
```

### 2. Backend Service Layer

```
[ ] Services contain business logic (not handlers)
[ ] Handlers are thin (validate, call service, respond)
[ ] Errors use sentinel pattern (ErrNotFound, etc.)
[ ] Context passed through all layers
[ ] Transactions used for multi-step operations
```

### 3. API Contract

```
[ ] Request/response types match between Go and TypeScript
[ ] Error responses follow standard format
[ ] Pagination uses consistent pattern (cursor or offset)
[ ] Authentication required on protected routes
[ ] Rate limiting configured appropriately
```

### 4. State Management

```
[ ] Server state in TanStack Query (not Zustand)
[ ] Client-only state in Zustand (UI state, preferences)
[ ] No duplicate state between Query and Zustand
[ ] Store actions are minimal and focused
```

### 5. Type Safety

```
[ ] Go structs have correct json tags
[ ] TypeScript types match Go models
[ ] Zod schemas for runtime validation at boundaries
[ ] No 'any' types without justification
[ ] API client is fully typed
```

---

## Architecture Patterns

### Recommended Patterns

| Pattern | Use For | Example |
|---------|---------|---------|
| Service Layer | Business logic | `OrgService.CreateOrganization()` |
| Repository Pattern | Complex queries | When GORM methods get unwieldy |
| Query Keys Factory | Cache management | `queryKeys.users.detail(id)` |
| Optimistic Updates | Fast UI feedback | Likes, toggles, simple updates |
| Pessimistic Updates | Data integrity | Payments, critical operations |

### Anti-Patterns to Flag

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Handler with DB access | Bypasses business logic | Move to service |
| Query without queryKey | Cache won't work | Use queryKeys factory |
| Zustand for server data | Stale data, sync issues | Use TanStack Query |
| Raw SQL in handlers | SQL injection risk | Use GORM or parameterized |
| Missing error handling | Silent failures | Wrap with context |

---

## Review Questions

### For New Features

1. Where does the business logic live? (Should be services)
2. How is the data fetched? (Should use TanStack Query)
3. How is the cache invalidated? (Should use queryKeys)
4. What happens on error? (Should have clear error states)
5. Is the feature tested? (Should have unit + integration tests)

### For Refactoring

1. What is the current pain point?
2. Does the refactor maintain backward compatibility?
3. Are there tests covering the affected code?
4. What's the migration path for existing data?
5. Can it be done incrementally?

### For Performance

1. Are there N+1 query issues? (Use Preload)
2. Is pagination implemented? (For list endpoints)
3. Are indexes present on queried columns?
4. Is caching used appropriately? (Redis/Dragonfly)
5. Are expensive operations async? (Background jobs)

---

## Output Format

When reviewing, provide:

```markdown
## Architecture Review: {Feature/Area}

### Summary
{1-2 sentence overview}

### Strengths
- {What's done well}

### Concerns
- {Potential issues}

### Recommendations
1. {Specific actionable item}
2. {Another item}

### Risk Assessment
- **Breaking changes:** Yes/No
- **Data migration needed:** Yes/No
- **Estimated complexity:** Low/Medium/High
```
