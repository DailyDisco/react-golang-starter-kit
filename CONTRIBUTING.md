# Contributing to React + Go Starter Kit

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- **Go 1.25+**
- **Node.js 20+** (with npm)
- **PostgreSQL 16+**
- **Redis/Dragonfly** (optional, for caching)
- **Docker** (optional, for containerized development)

### Setup

```bash
# Clone the repository
git clone https://github.com/DailyDisco/react-golang-starter-kit.git
cd react-golang-starter-kit

# Copy environment files
cp .env.example .env

# Install dependencies
cd frontend && npm install && cd ..
cd backend && go mod download && cd ..

# Start development
make dev
```

## Development Workflow

### Branch Naming

Use prefixes for branch names:

| Prefix | Purpose |
|--------|---------|
| `feature/` | New features |
| `fix/` | Bug fixes |
| `hotfix/` | Urgent production fixes |
| `docs/` | Documentation changes |
| `refactor/` | Code refactoring |
| `chore/` | Maintenance tasks |

Examples: `feature/user-authentication`, `fix/login-validation`, `docs/api-reference`

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/). Commits must follow this format:

```
<type>: <description>

[optional body]
```

**Allowed types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code restructuring |
| `chore` | Maintenance tasks |
| `build` | Build system changes |
| `ci` | CI configuration |
| `revert` | Revert a commit |

**Examples:**

```bash
feat: add user profile photo upload
fix: resolve race condition in session refresh
docs: update API authentication guide
chore: upgrade Go dependencies
```

## Code Standards

### Backend (Go)

- Run `go fmt ./...` before committing
- Run `go vet ./...` to check for issues
- Follow patterns in existing code (services, handlers, repository)
- Use `context.Context` as the first parameter for all service methods
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use sentinel errors for expected conditions

```go
// Good
if err != nil {
    return fmt.Errorf("users.GetByID: %w", err)
}

// Bad
if err != nil {
    return err  // No context
}
```

### Frontend (TypeScript/React)

- Run `npm run lint` and `npm run typecheck` before committing
- Use the query key factory pattern (`queryKeys.users.detail(id)`)
- Follow TanStack Query patterns for data fetching
- Use Zustand stores for UI state only (server state in TanStack Query)

```typescript
// Good - use query keys factory
queryClient.invalidateQueries({ queryKey: queryKeys.users.all });

// Bad - hardcoded query keys
queryClient.invalidateQueries({ queryKey: ['users'] });
```

## Testing Requirements

### When Tests Are Required

- **Always**: Business logic, utility functions, API handlers
- **Usually**: React hooks with complex logic, form validation
- **Selectively**: UI components (focus on behavior, not styling)

### Backend Tests

```bash
# Run all backend tests
cd backend && go test ./...

# Run with coverage
cd backend && go test -cover ./...

# Run integration tests (requires database)
cd backend && INTEGRATION_TEST=true go test ./...
```

**Test patterns:**

- Use table-driven tests for multiple cases
- Use `testutil.NewTestTransaction` for database tests (auto-rollback)
- Mock external services (Stripe, S3, email)

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Transform(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Frontend Tests

```bash
# Run frontend tests
cd frontend && npm run test

# Run with coverage
cd frontend && npm run test:coverage
```

## Pull Request Process

### Before Submitting

1. **Sync with main**: `git pull origin master`
2. **Run tests**: `make test` or `npm run test && go test ./...`
3. **Check formatting**: `npm run lint` and `go fmt ./...`
4. **Check types**: `npm run typecheck`

### PR Guidelines

- Keep PRs focused and reviewable (**<400 lines** when possible)
- Include a clear description of changes
- Link related issues
- Add tests for new functionality
- Update documentation if needed

### PR Description Template

```markdown
## Summary
Brief description of what this PR does.

## Changes
- List of specific changes

## Testing
How you tested these changes.

## Related Issues
Fixes #123
```

### Review Checklist

Reviewers will check for:

- [ ] Code follows project patterns
- [ ] Tests are included and passing
- [ ] No security vulnerabilities introduced
- [ ] Error handling is appropriate
- [ ] Documentation is updated if needed
- [ ] No unnecessary dependencies added

## Project Structure

```
backend/
├── cmd/main.go           # Entry point
├── internal/
│   ├── auth/             # JWT, OAuth, 2FA
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── repository/       # Data access
│   ├── models/           # GORM models
│   ├── middleware/       # HTTP middleware
│   └── testutil/         # Test utilities
└── migrations/           # SQL migrations

frontend/
├── app/
│   ├── hooks/            # TanStack Query hooks
│   ├── lib/              # Utilities (query-keys, guards)
│   ├── stores/           # Zustand stores
│   ├── routes/           # File-based routing
│   └── services/api/     # API client
└── components/           # React components
```

## Getting Help

- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
- **Documentation**: Check `/docs` directory for guides

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
