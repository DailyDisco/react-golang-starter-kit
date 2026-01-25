# Backend Testing Guide

This guide documents the testing patterns and infrastructure for the Go backend.

## Testing Strategy

The codebase uses a **two-tier testing strategy**:

| Test Type | Purpose | Speed | Dependencies | Run With |
|-----------|---------|-------|--------------|----------|
| **Unit Tests** | Test business logic in isolation | Fast (~1s) | Mocks only | `go test ./...` |
| **Integration Tests** | Test with real database | Slower (~30s) | PostgreSQL | `INTEGRATION_TEST=true go test ./...` |

## Repository Pattern

Services use **dependency injection** with repository interfaces for testability.

### Interface Definition

Define repository interfaces in `internal/repository/interfaces.go`:

```go
type SessionRepository interface {
    Create(ctx context.Context, session *models.UserSession) error
    FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error)
    DeleteByID(ctx context.Context, sessionID, userID uint) (int64, error)
    // ... other methods
}
```

### GORM Implementation

Implement the interface for production in `internal/repository/gorm_session.go`:

```go
type GormSessionRepository struct {
    db *gorm.DB
}

func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
    return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
    return r.db.WithContext(ctx).Create(session).Error
}
```

### Service with DI

Services accept repository interfaces:

```go
type SessionService struct {
    sessionRepo repository.SessionRepository
    historyRepo repository.LoginHistoryRepository
}

// Production constructor (uses global DB)
func NewSessionService() *SessionService {
    return &SessionService{
        sessionRepo: repository.NewGormSessionRepository(database.DB),
        historyRepo: repository.NewGormLoginHistoryRepository(database.DB),
    }
}

// Testable constructor (accepts any implementation)
func NewSessionServiceWithRepo(sessionRepo repository.SessionRepository, historyRepo repository.LoginHistoryRepository) *SessionService {
    return &SessionService{
        sessionRepo: sessionRepo,
        historyRepo: historyRepo,
    }
}
```

## Mock Implementations

Mocks are in `internal/testutil/mocks/repository.go`:

```go
type MockSessionRepository struct {
    mu       sync.RWMutex
    sessions map[uint][]models.UserSession

    // Error injection
    CreateErr error
    FindByUserIDErr error

    // Call tracking
    CreateCalls int
    FindByUserIDCalls int
}

// AddSession - helper for test setup
func (m *MockSessionRepository) AddSession(session models.UserSession) { ... }

// Reset - clear all data between tests
func (m *MockSessionRepository) Reset() { ... }
```

## Writing Unit Tests

Use table-driven tests with mocks:

```go
func TestSessionService_CreateSessionWithContext(t *testing.T) {
    tests := []struct {
        name    string
        userID  uint
        repoErr error
        wantErr bool
    }{
        {
            name:    "successful creation",
            userID:  1,
            wantErr: false,
        },
        {
            name:    "repository error",
            userID:  1,
            repoErr: errors.New("database error"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            sessionRepo := mocks.NewMockSessionRepository()
            historyRepo := mocks.NewMockLoginHistoryRepository()
            sessionRepo.CreateErr = tt.repoErr

            svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
            req := httptest.NewRequest(http.MethodPost, "/login", nil)
            ctx := context.Background()

            // Act
            session, err := svc.CreateSessionWithContext(ctx, tt.userID, "token", req)

            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && session == nil {
                t.Error("expected session, got nil")
            }
        })
    }
}
```

## Integration Tests with Testcontainers

For tests that need a real database, use testcontainers:

```go
//go:build integration

func TestSessionService_Integration(t *testing.T) {
    // Starts a real PostgreSQL container
    pg := testutil.SetupPostgresContainer(t)

    // Create real repository
    repo := repository.NewGormSessionRepository(pg.DB)
    svc := NewSessionServiceWithRepo(repo, historyRepo)

    // Test with real database
    session, err := svc.CreateSession(...)
    require.NoError(t, err)
}
```

Run integration tests:

```bash
INTEGRATION_TEST=true go test ./...
```

## Existing Test Infrastructure

### testutil/database.go

- `SetupTestDB(t)` - Connect to test database with migrations
- `NewTestTransaction(t, db)` - Wrap test in rollback transaction
- `SkipIfNotIntegration(t)` - Skip if not integration mode

### testutil/fixtures.go

Factory builders for test data:

```go
user := testutil.NewUserFactory().
    WithEmail("test@example.com").
    WithRole("admin").
    Build()
```

### testutil/mocks/

Pre-built mocks:
- `MockSessionRepository` - Session data access
- `MockLoginHistoryRepository` - Login history
- `MockEmailProvider` - Email sending
- `MockS3Client` - S3 operations
- `MockStripeClient` - Stripe API

## Running Tests

```bash
# Unit tests only (fast)
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests (requires Docker)
INTEGRATION_TEST=true go test ./...

# Specific package
go test -v ./internal/services/...

# Specific test
go test -v -run TestSessionService ./internal/services/
```

## Converting Existing Services

To convert a service from global DB to repository pattern:

1. **Define interface** in `internal/repository/interfaces.go`
2. **Implement GORM version** in `internal/repository/gorm_*.go`
3. **Create mock** in `internal/testutil/mocks/repository.go`
4. **Update service** to accept interface via constructor
5. **Add `WithContext` versions** of methods
6. **Write unit tests** using mocks

### Example: Converting OrgService

```go
// 1. Interface (interfaces.go)
type OrganizationRepository interface {
    FindBySlug(ctx context.Context, slug string) (*models.Organization, error)
    Create(ctx context.Context, org *models.Organization) error
}

// 2. GORM implementation (gorm_organization.go)
type GormOrganizationRepository struct { db *gorm.DB }

// 3. Mock (mocks/repository.go)
type MockOrganizationRepository struct { ... }

// 4. Update service
type OrgService struct {
    repo repository.OrganizationRepository
}

func NewOrgServiceWithRepo(repo repository.OrganizationRepository) *OrgService {
    return &OrgService{repo: repo}
}
```

## Best Practices

1. **Always use `context.Context`** - Pass context through all layers
2. **Test error paths** - Include tests for repository errors
3. **Use table-driven tests** - Makes adding cases easy
4. **Mock at boundaries** - Only mock external dependencies
5. **Track mock calls** - Verify correct methods were called
6. **Reset mocks between tests** - Prevent test pollution
7. **Keep mocks simple** - In-memory maps, not complex logic
