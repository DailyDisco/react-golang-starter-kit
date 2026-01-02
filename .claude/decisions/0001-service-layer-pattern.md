# ADR-0001: Service Layer Pattern for Backend

## Status

Accepted

## Date

2025-12-01

## Context

The backend needs a clear separation between HTTP handling and business logic. Without this separation:

- Handlers become bloated with business logic
- Testing requires HTTP request/response mocking
- Code reuse across different entry points (HTTP, CLI, workers) is difficult
- Error handling becomes inconsistent

We needed to decide how to structure the backend code for maintainability and testability.

## Decision

Implement a **Service Layer Pattern** where:

1. **Handlers** are thin HTTP adapters that:
   - Parse requests and validate input
   - Call service methods
   - Format responses and handle HTTP-specific errors

2. **Services** contain all business logic and:
   - Define sentinel errors for expected conditions (`ErrNotFound`, `ErrUnauthorized`)
   - Accept `context.Context` as first parameter
   - Return domain objects and errors
   - Are injected with dependencies (DB, cache, external services)

3. **Repository pattern is NOT used** - GORM is called directly from services for simplicity

## Consequences

### Positive

- Handlers are easy to read and maintain (~20-30 lines each)
- Services are testable without HTTP mocking
- Business logic can be reused (e.g., in background jobs)
- Consistent error handling via sentinel errors and `errors.Is()`

### Negative

- Additional layer adds some indirection
- Must remember to always use `WithContext(ctx)` in services
- Slightly more boilerplate for simple CRUD operations

### Neutral

- Team must follow convention of thin handlers
- `/scaffold-service` skill enforces this pattern automatically

## Alternatives Considered

### Handler-only approach

Put all logic directly in handlers.

**Rejected because:** Makes testing harder, code reuse impossible, handlers become unmaintainable.

### Full Repository Pattern

Add a repository layer between services and GORM.

**Rejected because:** Over-engineering for our needs. GORM already provides a clean API. Can add later if needed.

### Domain-Driven Design

Full DDD with aggregates, value objects, domain events.

**Rejected because:** Too complex for a starter kit. Service layer provides 80% of benefits with 20% of complexity.

## References

- [backend/internal/services/org_service.go](../../backend/internal/services/org_service.go) - Example service
- [backend/internal/handlers/org_handlers.go](../../backend/internal/handlers/org_handlers.go) - Example handler
- [.claude/skills/scaffold-service/SKILL.md](../skills/scaffold-service/SKILL.md) - Scaffolding enforces pattern
