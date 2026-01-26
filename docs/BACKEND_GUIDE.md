# Backend Development Guide

Comprehensive guide for Go backend development with Chi router, GORM, and layered architecture.

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Service Layer](#service-layer)
- [Repository Pattern](#repository-pattern)
- [Middleware Stack](#middleware-stack)
- [Error Handling](#error-handling)
- [Database Patterns](#database-patterns)
- [Background Jobs](#background-jobs)
- [Best Practices](#best-practices)

---

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Docker (recommended)

### Local Development

```bash
cd backend

# Install dependencies
go mod download

# Start with hot-reloading (Air)
air

# Or run directly
go run cmd/main.go

# API runs at http://localhost:8080
```

### Docker Development

```bash
# From project root
docker compose up backend

# View logs
docker compose logs -f backend
```

---

## Architecture Overview

The backend follows a layered architecture:

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Handlers                      │
│         (Parse requests, call services)              │
├─────────────────────────────────────────────────────┤
│                   Services                           │
│         (Business logic, orchestration)              │
├─────────────────────────────────────────────────────┤
│                  Repositories                        │
│         (Data access, database queries)              │
├─────────────────────────────────────────────────────┤
│                    Models                            │
│            (GORM entities, structs)                  │
└─────────────────────────────────────────────────────┘
```

### Directory Structure

```
backend/
├── cmd/main.go              # Entry point, route setup
├── internal/
│   ├── auth/                # JWT, OAuth, 2FA, middleware
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic layer
│   ├── repository/          # Data access interfaces + implementations
│   ├── models/              # GORM models
│   ├── middleware/          # HTTP middleware (CSRF, CORS, etc.)
│   ├── cache/               # Redis/Dragonfly caching
│   ├── stripe/              # Billing integration
│   ├── websocket/           # Real-time events
│   └── workers/             # Background job definitions
├── migrations/              # SQL migrations (golang-migrate)
└── config/                  # Configuration loading
```

---

## Service Layer

Services contain business logic and orchestrate operations across repositories.

### Pattern: Sentinel Errors

Define domain-specific errors at the top of each service:

```go
package services

import "errors"

// Sentinel errors for organization operations
var (
    ErrOrgNotFound       = errors.New("organization not found")
    ErrOrgSlugTaken      = errors.New("organization slug is already taken")
    ErrInvalidSlug       = errors.New("invalid slug format")
    ErrNotMember         = errors.New("user is not a member of this organization")
    ErrInsufficientRole  = errors.New("insufficient role permissions")
)
```

### Pattern: Constructor with Dependencies

Services receive their dependencies (repositories) via constructor injection:

```go
// OrgService handles organization business logic
type OrgService struct {
    db         *gorm.DB
    orgRepo    repository.OrganizationRepository
    memberRepo repository.OrganizationMemberRepository
}

// NewOrgServiceWithRepo creates a service with injected repositories.
// Use this constructor for testing with mock repositories.
func NewOrgServiceWithRepo(
    db *gorm.DB,
    orgRepo repository.OrganizationRepository,
    memberRepo repository.OrganizationMemberRepository,
) *OrgService {
    return &OrgService{
        db:         db,
        orgRepo:    orgRepo,
        memberRepo: memberRepo,
    }
}
```

### Pattern: Context Usage

Always accept `context.Context` as the first parameter:

```go
func (s *OrgService) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
    org, err := s.orgRepo.FindBySlug(ctx, slug)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrOrgNotFound
        }
        return nil, fmt.Errorf("find org by slug: %w", err)
    }
    return org, nil
}
```

---

## Repository Pattern

Repositories define data access interfaces for testable services.

### Interface Definition

```go
// repository/interfaces.go

// SessionRepository defines data access operations for user sessions.
type SessionRepository interface {
    Create(ctx context.Context, session *models.UserSession) error
    FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error)
    DeleteByID(ctx context.Context, sessionID, userID uint) (int64, error)
    DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
```

### GORM Implementation

```go
// repository/session_repository.go

type GormSessionRepository struct {
    db *gorm.DB
}

func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
    return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
    return r.db.WithContext(ctx).Create(session).Error
}

func (r *GormSessionRepository) FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error) {
    var sessions []models.UserSession
    err := r.db.WithContext(ctx).
        Where("user_id = ? AND expires_at > ?", userID, now).
        Order("last_active_at DESC").
        Find(&sessions).Error
    return sessions, err
}
```

### Mock Implementation (Testing)

```go
// testutil/mocks/repository.go

type MockSessionRepository struct {
    Sessions []models.UserSession
    CreateFn func(ctx context.Context, session *models.UserSession) error
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
    if m.CreateFn != nil {
        return m.CreateFn(ctx, session)
    }
    m.Sessions = append(m.Sessions, *session)
    return nil
}
```

### Available Repositories

| Repository | Purpose |
|------------|---------|
| `SessionRepository` | User sessions (JWT tokens) |
| `LoginHistoryRepository` | Login attempt records |
| `UserRepository` | User CRUD |
| `OrganizationRepository` | Organization CRUD |
| `OrganizationMemberRepository` | Org membership |
| `OrganizationInvitationRepository` | Org invitations |
| `SubscriptionRepository` | Billing subscriptions |
| `SystemSettingRepository` | System configuration |
| `IPBlocklistRepository` | IP blocking |
| `UsageEventRepository` | Usage tracking events |
| `UsagePeriodRepository` | Usage period summaries |
| `UsageAlertRepository` | Usage limit alerts |

---

## Middleware Stack

Middleware is applied in the router setup (`cmd/main.go`):

```go
r := chi.NewRouter()

// Global middleware (applied to all routes)
r.Use(middleware.RequestID)           // Correlation ID for tracing
r.Use(middleware.RealIP)              // Get real IP behind proxy
r.Use(mw.NewLogger().Handler)         // Structured logging
r.Use(mw.RecoveryMiddleware)          // Panic recovery
r.Use(mw.CORS(corsConfig))            // CORS headers
r.Use(mw.SecurityHeaders(secConfig))  // Security headers (CSP, etc.)
r.Use(mw.Prometheus())                // Metrics collection
```

### Available Middleware

| Middleware | File | Purpose |
|------------|------|---------|
| `RequestID` | `request_id.go` | Adds `X-Request-ID` for tracing |
| `Logger` | `logger.go` | Structured request/response logging |
| `Recovery` | `recovery.go` | Catches panics, logs stack trace |
| `CORS` | `cors.go` | Cross-origin resource sharing |
| `SecurityHeaders` | `security.go` | CSP, X-Frame-Options, etc. |
| `CSRF` | `csrf.go` | CSRF token validation |
| `Prometheus` | `prometheus.go` | HTTP metrics |
| `CacheHeaders` | `cache_headers.go` | Cache-Control headers |
| `Idempotency` | `idempotency.go` | Idempotency key handling |
| `Usage` | `usage.go` | Usage metering |
| `Sentry` | `sentry.go` | Error tracking |

### Auth Middleware

Authentication middleware is in `internal/auth/middleware.go`:

```go
// Protect routes requiring authentication
r.Route("/api/v1/users", func(r chi.Router) {
    r.Use(auth.RequireAuth(jwtService))  // Validates JWT
    r.Get("/me", handlers.GetCurrentUser)
})

// Protect routes requiring specific roles
r.Route("/api/v1/admin", func(r chi.Router) {
    r.Use(auth.RequireAuth(jwtService))
    r.Use(auth.RequireRole("admin", "super_admin"))
    r.Get("/users", handlers.ListAllUsers)
})
```

---

## Error Handling

### Pattern: Wrap Errors with Context

```go
func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound  // Domain error
        }
        return nil, fmt.Errorf("find user by id %d: %w", id, err)
    }
    return user, nil
}
```

### Pattern: Handle Errors in Handlers

```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)

    user, err := h.service.GetByID(r.Context(), uint(userID))
    if err != nil {
        switch {
        case errors.Is(err, services.ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            log.Error().Err(err).Msg("failed to get user")
            http.Error(w, "Internal error", http.StatusInternalServerError)
        }
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

### Structured Error Responses

Use a consistent error response format:

```go
type ErrorResponse struct {
    Error     string `json:"error"`
    Code      string `json:"code,omitempty"`
    RequestID string `json:"request_id,omitempty"`
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message, code string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error:     message,
        Code:      code,
        RequestID: middleware.GetReqID(r.Context()),
    })
}
```

---

## Database Patterns

### Always Use Context

```go
// ✅ Correct - uses context for cancellation/timeout
s.db.WithContext(ctx).Where("id = ?", id).First(&user)

// ❌ Wrong - no context
s.db.Where("id = ?", id).First(&user)
```

### Preload Associations (N+1 Prevention)

```go
// ✅ Correct - eager loads members in one query
s.db.WithContext(ctx).
    Preload("Members").
    Preload("Members.User").
    Where("slug = ?", slug).
    First(&org)

// ❌ Wrong - causes N+1 queries
s.db.WithContext(ctx).Where("slug = ?", slug).First(&org)
for _, member := range org.Members {
    s.db.First(&member.User, member.UserID)  // N queries!
}
```

### Transaction Handling

```go
func (s *OrgService) TransferOwnership(ctx context.Context, orgID, newOwnerID uint) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Remove owner role from current owner
        if err := tx.Model(&models.OrgMember{}).
            Where("org_id = ? AND role = ?", orgID, "owner").
            Update("role", "admin").Error; err != nil {
            return err
        }

        // Add owner role to new owner
        if err := tx.Model(&models.OrgMember{}).
            Where("org_id = ? AND user_id = ?", orgID, newOwnerID).
            Update("role", "owner").Error; err != nil {
            return err
        }

        return nil  // Commit
    })
}
```

---

## Background Jobs

Background jobs use [River](https://github.com/riverqueue/river), a PostgreSQL-backed job queue.

### Job Definition

```go
// workers/data_export.go

type DataExportArgs struct {
    UserID   uint   `json:"user_id"`
    Format   string `json:"format"`
    RequestID string `json:"request_id"`
}

func (DataExportArgs) Kind() string { return "data_export" }

type DataExportWorker struct {
    river.WorkerDefaults[DataExportArgs]
    db *gorm.DB
}

func (w *DataExportWorker) Work(ctx context.Context, job *river.Job[DataExportArgs]) error {
    log.Info().
        Uint("user_id", job.Args.UserID).
        Str("format", job.Args.Format).
        Msg("Starting data export")

    // Export logic here...

    return nil
}
```

### Enqueueing Jobs

```go
// In a handler or service
_, err := riverClient.Insert(ctx, DataExportArgs{
    UserID:    userID,
    Format:    "json",
    RequestID: middleware.GetReqID(ctx),
}, nil)
```

### Configuration

Jobs are configured in `internal/workers/config.go` with retry policies, timeouts, and scheduling.

---

## Best Practices

### 1. Handler Responsibilities

Handlers should be thin - only parse input, call services, and format output:

```go
func (h *Handler) CreateOrg(w http.ResponseWriter, r *http.Request) {
    // Parse input
    var req CreateOrgRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, r, http.StatusBadRequest, "Invalid JSON", "INVALID_JSON")
        return
    }

    // Validate
    if err := validate.Struct(req); err != nil {
        writeError(w, r, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
        return
    }

    // Call service (business logic lives here)
    org, err := h.service.Create(r.Context(), req.Name, req.Slug)
    if err != nil {
        handleServiceError(w, r, err)
        return
    }

    // Format output
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(org)
}
```

### 2. Service Responsibilities

Services contain business logic, validation, and orchestration:

```go
func (s *OrgService) Create(ctx context.Context, name, slug string) (*models.Organization, error) {
    // Business validation
    if !slugRegex.MatchString(slug) {
        return nil, ErrInvalidSlug
    }

    // Check uniqueness
    existing, _ := s.orgRepo.FindBySlug(ctx, slug)
    if existing != nil {
        return nil, ErrOrgSlugTaken
    }

    // Create entity
    org := &models.Organization{
        Name: name,
        Slug: slug,
    }

    if err := s.orgRepo.Create(ctx, org); err != nil {
        return nil, fmt.Errorf("create org: %w", err)
    }

    return org, nil
}
```

### 3. Testing Approach

- **Unit tests**: Mock repositories, test service logic
- **Integration tests**: Real database, test full flow
- **See [TESTING.md](TESTING.md)** for detailed patterns

### 4. Logging

Use structured logging with zerolog:

```go
import "github.com/rs/zerolog/log"

log.Info().
    Str("org_slug", slug).
    Uint("user_id", userID).
    Msg("User joined organization")

log.Error().
    Err(err).
    Str("operation", "create_org").
    Msg("Failed to create organization")
```

---

## Additional Resources

- [TESTING.md](TESTING.md) - Comprehensive testing guide
- [FEATURES.md](FEATURES.md) - Feature documentation (JWT, RBAC, etc.)
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment guide
- [Chi Router Docs](https://go-chi.io/)
- [GORM Docs](https://gorm.io/)
- [River Job Queue](https://riverqueue.com/)
