# ADR 002: Background Jobs with River

## Status

Accepted

## Context

The application needs asynchronous job processing for:

- Sending verification and password reset emails
- Processing Stripe webhook events
- Future async tasks (reports, cleanup, notifications)

Requirements:
- Reliable job execution with retries
- Job persistence across restarts
- Minimal infrastructure dependencies
- Good observability
- Go-native solution

## Decision

We will use [River](https://github.com/riverqueue/river) as our background job system.

River is a PostgreSQL-backed job queue specifically designed for Go. Since we already use PostgreSQL, this eliminates the need for additional infrastructure like Redis.

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Handler   │────>│  River Job  │────>│   Worker    │
│  (enqueue)  │     │   Queue     │     │  (process)  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │  PostgreSQL │
                    │ (river_job) │
                    └─────────────┘
```

### Job Types

| Job | Queue | Max Retries | Use Case |
|-----|-------|-------------|----------|
| SendVerificationEmail | email | 5 | After user registration |
| SendPasswordResetEmail | email | 5 | Password reset request |
| ProcessStripeWebhook | webhooks | 3 | Async webhook handling |

### Configuration

```bash
JOBS_ENABLED=true
JOBS_WORKER_COUNT=10
JOBS_MAX_RETRIES=3
JOBS_TIMEOUT=30s
```

## Consequences

### Positive

- No additional infrastructure (uses existing PostgreSQL)
- ACID guarantees for job state
- Exactly-once processing semantics
- Built-in retries with exponential backoff
- Go-native with type-safe job arguments
- Easy local development

### Negative

- PostgreSQL becomes more critical (single point of failure)
- May need to scale PostgreSQL for high job volume
- Less ecosystem/tooling compared to Redis-based solutions
- Newer library with smaller community

## Alternatives Considered

### Option 1: Asynq (Redis-based)

- Pros: Mature, fast, good tooling, large community
- Cons: Requires Redis infrastructure, additional operational complexity

### Option 2: Temporal

- Pros: Powerful workflow engine, great for complex workflows
- Cons: Heavy infrastructure requirement, overkill for simple jobs

### Option 3: Simple goroutines with channels

- Pros: No dependencies, simple
- Cons: No persistence, no retries, jobs lost on restart

## References

- [River GitHub](https://github.com/riverqueue/river)
- [River Documentation](https://riverqueue.com/docs)
- [backend/internal/jobs/](../../backend/internal/jobs/)
