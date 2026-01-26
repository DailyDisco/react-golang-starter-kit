# Testing Guide

Comprehensive testing guide for both backend (Go) and frontend (React) development.

## Table of Contents

- [Quick Start](#quick-start)
- [Backend Unit Testing](#backend-unit-testing)
- [Backend Integration Testing](#backend-integration-testing)
- [Test Utilities](#test-utilities)
- [Frontend Testing](#frontend-testing)
- [E2E Testing](#e2e-testing)
- [CI Pipeline](#ci-pipeline)

---

## Quick Start

### Run All Tests

```bash
# From project root
npm run test              # Frontend + Backend
npm run test:frontend     # Frontend only
npm run test:backend      # Backend unit tests only

# Or using make
make test                 # All tests
make test-backend         # Backend only
make test-frontend        # Frontend only
```

### Backend Tests

```bash
cd backend

# Unit tests (no database required)
go test ./...

# Integration tests (requires test database)
INTEGRATION_TEST=true go test ./internal/services/... -v

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend Tests

```bash
cd frontend

# Run tests once (CI mode)
npm run test:fast

# Watch mode (development)
npm test

# With coverage
npm run test:coverage

# Visual UI
npm run test:ui
```

---

## Backend Unit Testing

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"valid with subdomain", "user@mail.example.com", false},
        {"missing @", "userexample.com", true},
        {"missing domain", "user@", true},
        {"empty string", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

### Using testify

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserService_Create(t *testing.T) {
    // require stops test on failure, assert continues
    repo := &MockUserRepository{}
    svc := NewUserService(repo)

    user, err := svc.Create(context.Background(), "test@example.com", "Test User")

    require.NoError(t, err)           // Stops if error
    assert.NotZero(t, user.ID)        // Continues if fails
    assert.Equal(t, "Test User", user.Name)
}
```

### Mocking with Interfaces

Services depend on repository interfaces, making them easy to mock:

```go
// Mock implementation
type MockUserRepository struct {
    Users    []models.User
    FindByIDFn func(ctx context.Context, id uint) (*models.User, error)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    if m.FindByIDFn != nil {
        return m.FindByIDFn(ctx, id)
    }
    for _, u := range m.Users {
        if u.ID == id {
            return &u, nil
        }
    }
    return nil, gorm.ErrRecordNotFound
}

// In test
func TestGetUser_NotFound(t *testing.T) {
    repo := &MockUserRepository{
        FindByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
            return nil, gorm.ErrRecordNotFound
        },
    }
    svc := NewUserServiceWithRepo(nil, repo)

    _, err := svc.GetByID(context.Background(), 999)

    assert.ErrorIs(t, err, ErrUserNotFound)
}
```

---

## Backend Integration Testing

Integration tests use a real PostgreSQL database for realistic testing.

### Setup Test Database

```bash
# Option 1: Start test database container
docker run -d --name starter-test-db \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -e POSTGRES_DB=starter_kit_test \
  -p 5433:5432 \
  postgres:16-alpine

# Option 2: Use docker compose
docker compose -f compose.test.yml up -d
```

### Running Integration Tests

```bash
cd backend

# Run all integration tests
INTEGRATION_TEST=true go test ./internal/services/... -v

# Run specific test file
INTEGRATION_TEST=true go test ./internal/services/org_service_integration_test.go -v

# With custom database
TEST_DB_HOST=localhost TEST_DB_PORT=5434 \
INTEGRATION_TEST=true go test ./internal/services/... -v
```

### Available Integration Test Suites

| File | Coverage |
|------|----------|
| `org_service_integration_test.go` | Organizations, memberships, invitations |
| `session_service_integration_test.go` | Sessions, device tracking |
| `settings_service_integration_test.go` | System settings, batched updates |
| `usage_service_integration_test.go` | Usage tracking, limits, alerts |
| `totp_service_integration_test.go` | 2FA setup and verification |
| `file_service_integration_test.go` | File operations with S3/LocalStack |
| `health_service_integration_test.go` | Health check endpoints |
| `user_preferences_service_integration_test.go` | User preferences CRUD |

### Test Transaction Pattern

Use `NewTestTransaction` for automatic rollback and test isolation:

```go
func TestOrgService_Create(t *testing.T) {
    testutil.SkipIfNotIntegration(t)
    db := testutil.SetupTestDB(t)
    tt := testutil.NewTestTransaction(t, db)
    defer tt.Rollback()  // Automatic cleanup

    // Create service with transaction DB
    svc := services.NewOrgService(tt.DB)

    // Test
    org, err := svc.Create(context.Background(), "Test Org", "test-org", userID)

    require.NoError(t, err)
    assert.Equal(t, "Test Org", org.Name)
    // No cleanup needed - transaction rolls back automatically
}
```

---

## Test Utilities

The `backend/internal/testutil/` package provides testing helpers.

### Database Helpers

```go
// Skip if not running integration tests
testutil.SkipIfNotIntegration(t)

// Get shared test database connection
db := testutil.GetTestDB(t)

// Setup with migrations
db := testutil.SetupTestDB(t)

// Start isolated transaction
tt := testutil.NewTestTransaction(t, db)
defer tt.Rollback()

// Truncate all tables (for cleanup)
testutil.TruncateTables(t, db)

// Clean specific test data
testutil.CleanupTestData(t, db, "test@%")  // Pattern-based
```

### Test Fixtures

Use factories to create test data with sensible defaults:

```go
// Create user with defaults
user := testutil.NewUserFactory().Build()

// Customize as needed
admin := testutil.NewUserFactory().
    WithName("Admin User").
    WithEmail("admin@test.com").
    AsAdmin().
    Build()

// Create in database
user := testutil.CreateTestUser(t, db)
adminUser := testutil.CreateTestUser(t, db, testutil.WithRole("admin"))

// Create organization with owner
org := testutil.CreateTestOrganization(t, db, "My Org", ownerID)

// Create org membership
member := testutil.CreateTestOrgMember(t, db, orgID, userID, "member")
```

### HTTP Test Helpers

```go
// Create test request with context
req := testutil.NewRequest(t, "GET", "/api/v1/users/1", nil)

// With authentication
req := testutil.NewAuthenticatedRequest(t, "POST", "/api/v1/orgs", body, userID)

// Assert response
testutil.AssertStatus(t, resp, http.StatusOK)
testutil.AssertJSON(t, resp, &result)
```

### Mock Implementations

Available mocks in `testutil/mocks/`:

| Mock | Purpose |
|------|---------|
| `MockStripeClient` | Stripe API calls |
| `MockS3Client` | S3 file operations |
| `MockEmailService` | Email sending |
| `MockSessionRepository` | Session data access |
| `MockUserRepository` | User data access |
| `MockOrganizationRepository` | Org data access |

---

## Frontend Testing

### Vitest Configuration

Tests are configured in `vite.config.ts` with:
- Happy DOM for fast DOM simulation
- Global test functions (no imports needed)
- Coverage reporting

### Running Tests

```bash
cd frontend

npm run test:fast     # Run once (CI)
npm test              # Watch mode
npm run test:coverage # With coverage report
npm run test:ui       # Visual test runner
```

### Component Testing

```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);

    fireEvent.click(screen.getByText('Click'));

    expect(handleClick).toHaveBeenCalledOnce();
  });

  it('is disabled when loading', () => {
    render(<Button loading>Submit</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Testing Hooks

```tsx
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useUser } from './useUser';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useUser', () => {
  it('fetches user data', async () => {
    const { result } = renderHook(() => useUser(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.name).toBe('Test User');
  });
});
```

### Testing Zustand Stores

```tsx
import { act } from '@testing-library/react';
import { useAuthStore } from './auth-store';

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store between tests
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
    });
  });

  it('sets user on login', () => {
    const user = { id: 1, name: 'Test', email: 'test@example.com' };

    act(() => {
      useAuthStore.getState().login(user);
    });

    expect(useAuthStore.getState().user).toEqual(user);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('clears user on logout', () => {
    // Setup: user is logged in
    useAuthStore.setState({ user: { id: 1 }, isAuthenticated: true });

    act(() => {
      useAuthStore.getState().logout();
    });

    expect(useAuthStore.getState().user).toBeNull();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});
```

---

## E2E Testing

End-to-end tests use Playwright.

### Setup

```bash
cd frontend

# Install Playwright browsers
npx playwright install

# Run E2E tests
npm run test:e2e

# With UI mode
npm run test:e2e -- --ui

# Headed mode (see browser)
npm run test:e2e -- --headed
```

### Writing E2E Tests

```typescript
// e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('user can log in', async ({ page }) => {
    await page.goto('/login');

    await page.fill('[name="email"]', 'test@example.com');
    await page.fill('[name="password"]', 'Password123!');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('text=Welcome')).toBeVisible();
  });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.fill('[name="email"]', 'wrong@example.com');
    await page.fill('[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Invalid credentials')).toBeVisible();
  });
});
```

---

## CI Pipeline

### What Runs in CI

GitHub Actions runs on every PR:

1. **Lint** - `npm run lint` (frontend)
2. **Type Check** - `npm run typecheck` (frontend)
3. **Backend Tests** - `go test ./...`
4. **Frontend Tests** - `npm run test:fast`
5. **Build** - Verify production build works

### Integration Tests in CI

Integration tests run with a PostgreSQL service container:

```yaml
# .github/workflows/test.yml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: starter_kit_test
    ports:
      - 5433:5432

steps:
  - name: Run integration tests
    env:
      INTEGRATION_TEST: true
      TEST_DB_HOST: localhost
      TEST_DB_PORT: 5433
    run: go test ./internal/services/... -v
```

### Coverage Requirements

- No strict percentage required
- Focus on critical path coverage
- PRs should not decrease coverage significantly

---

## Best Practices

### Do

- Use table-driven tests for comprehensive coverage
- Mock at boundaries (database, external APIs)
- Use `testutil.NewTestTransaction` for database isolation
- Name tests descriptively: `TestUserService_Create_WithDuplicateEmail`
- Run tests in parallel when possible

### Don't

- Don't test implementation details
- Don't create flaky tests (random data, timing issues)
- Don't skip error case testing
- Don't test third-party libraries
- Don't share state between parallel tests

---

## Additional Resources

- [BACKEND_GUIDE.md](BACKEND_GUIDE.md) - Backend architecture
- [FRONTEND_GUIDE.md](FRONTEND_GUIDE.md) - Frontend development
- [Testify docs](https://github.com/stretchr/testify)
- [Vitest docs](https://vitest.dev/)
- [Playwright docs](https://playwright.dev/)
