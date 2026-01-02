# Project Memory

Persistent context for Claude Code sessions. This file is version controlled and automatically loaded at session start.

---

## Project Overview

React + Go full-stack SaaS starter kit with multi-tenant support, Stripe billing, WebSocket real-time updates, and OAuth authentication.

## Stack

| Layer | Technology |
|-------|------------|
| Frontend | React 19, TanStack Router/Query, Zustand, ShadCN UI, Tailwind |
| Backend | Go 1.25, Chi router, GORM, PostgreSQL |
| Auth | JWT + httpOnly cookies, OAuth (Google/GitHub), 2FA (TOTP) |
| Billing | Stripe subscriptions and usage-based billing |
| Real-time | WebSocket with hub pattern |
| Cache | Redis/Dragonfly |

## Current Focus

*Updated during active development sessions*

## Key Decisions

| Date | Decision | Rationale |
|------|----------|-----------|
| - | TanStack Router over React Router | Type-safe routing, file-based, better DX |
| - | Chi over Gin/Fiber | Lightweight, stdlib-compatible, middleware composition |
| - | GORM over sqlx | Rapid development, automatic migrations, good enough perf |
| - | Service layer pattern | Separation of concerns, testable business logic |
| - | Sentinel errors | Explicit error handling, type-safe error checking |

## Established Patterns

| Pattern | Description |
|---------|-------------|
| Query Keys | Factory pattern in `query-keys.ts` for cache management |
| Service Layer | Business logic in services, handlers are thin HTTP adapters |
| Sentinel Errors | `var ErrNotFound = errors.New("not found")` for expected conditions |
| Context Everywhere | All DB queries use `WithContext(ctx)` for cancellation |
| Cache Invalidation | Mutations invalidate related query keys via `onSuccess` |

## Learnings & Gotchas

- **Cache Keys**: Use `strconv.FormatUint()` for IDs, NOT `string(rune())` - the latter produces garbage
- **N+1 Queries**: Always use `.Preload()` for associations in GORM
- **WebSocket Auth**: Token passed in first message, not URL query param
- **Stripe Webhooks**: Must verify signature before processing, idempotency key required
- **JWT Refresh**: Rotation enabled - new refresh token on each refresh

## Team Agreements

- All PRs require review
- Tests required for services and handlers
- Conventional commits for all messages
- No `any` types in TypeScript
- All errors must be wrapped with context in Go

## Context for Next Session

*Add notes via: `.claude/hooks/memory-sync.sh add-note "your note"`*

---

## Session History
**Last Session:** 2026-01-02T13:22:27-06:00

*Auto-updated by memory-sync hook*
