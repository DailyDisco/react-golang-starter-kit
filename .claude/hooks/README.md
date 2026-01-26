# Project Hooks

This directory contains Claude Code hooks for pattern validation and project automation.

## Hook Overview

| Hook | Trigger | Purpose | Exit Code |
|------|---------|---------|-----------|
| `go-service-pattern.sh` | PostEdit `*_service.go` | Validates Go service patterns | 0=pass, 1=error |
| `query-key-pattern.sh` | PostEdit `query-keys.ts` | Validates query key factory pattern | 0=pass |
| `hook-naming.sh` | PostEdit `hooks/**/*.ts` | Validates React hook naming | 0=pass |
| `memory-sync.sh` | SessionStart/Stop | Loads/saves project memory | 0=success |
| `adr-index.sh` | PostEdit `decisions/*.md` | Auto-generates ADR index | 0=success |

## Validation Rules

### go-service-pattern.sh

Validates Go service files (`backend/internal/services/*_service.go`) follow project patterns.

**Checks performed:**

| Check | What It Validates | Severity |
|-------|-------------------|----------|
| Sentinel errors | `var ( Err... = errors.New(...) )` block present | Warning |
| Service struct | `type XxxService struct { db *gorm.DB }` defined | Error |
| Constructor | `NewXxxService(db *gorm.DB) *XxxService` function exists | Warning |
| Context param | Methods accept `ctx context.Context` as first parameter | Warning |
| Error wrapping | Errors wrapped with `fmt.Errorf("context: %w", err)` | Warning |
| WithContext | Database ops use `s.db.WithContext(ctx)` | Warning |

**Example compliant service:**

```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrEmailTaken   = errors.New("email already taken")
)

type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*User, error) {
    var user User
    if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return &user, nil
}
```

### query-key-pattern.sh

Validates `frontend/app/lib/query-keys.ts` follows the factory pattern.

**Checks performed:**

| Check | What It Validates |
|-------|-------------------|
| `as const` | Array definitions end with `as const` for type safety |
| `all` key | Each entity has base key: `all: ["entity"] as const` |
| Factory functions | `lists()` exists if `list(filters)` is used |
| Spread pattern | Spreads reference own entity: `...queryKeys.{entity}.all` |

**Expected pattern:**

```typescript
export const queryKeys = {
  users: {
    all: ["users"] as const,
    lists: () => [...queryKeys.users.all, "list"] as const,
    list: (filters: UserFilters) => [...queryKeys.users.lists(), filters] as const,
    details: () => [...queryKeys.users.all, "detail"] as const,
    detail: (id: number) => [...queryKeys.users.details(), id] as const,
  },
} as const;
```

### hook-naming.sh

Validates React hooks in `frontend/app/hooks/` follow naming conventions.

**File naming:**

| Location | Expected Pattern | Example |
|----------|------------------|---------|
| `hooks/queries/` | `use-{entity}.ts` | `use-users.ts` |
| `hooks/mutations/` | `use-{entity}-mutations.ts` | `use-users-mutations.ts` |

**Hook naming:**

| Type | Pattern | Examples |
|------|---------|----------|
| Query hooks | `use{Entity}` or `use{Entity}s` | `useUser`, `useUsers` |
| Mutation hooks | `use{Action}{Entity}` | `useCreateUser`, `useDeleteUser` |

**Valid action verbs for mutations:**

```
Create, Update, Delete, Remove, Add, Set, Toggle, Mark, Submit, Cancel,
Invite, Accept, Reject, Leave, Enable, Disable, Resend, Revoke, Reset,
Confirm, Approve
```

**Additional checks:**

- Query hooks must import from `@tanstack/react-query`
- Query hooks must use `queryKeys` from `../../lib/query-keys`
- Mutation hooks must use `useMutation` and `useQueryClient`
- Mutation hooks must call `invalidateQueries` on success

### memory-sync.sh

Manages project memory across Claude Code sessions.

**SessionStart:** Loads `.claude/memory/memory.md` content into session.

**SessionStop:** Updates the "Last Session" timestamp.

**Manual commands:**

```bash
# Add a note to memory
.claude/hooks/memory-sync.sh add-note "Your note here"

# Add a learning/gotcha
.claude/hooks/memory-sync.sh add-learning "What you learned"

# Add a decision
.claude/hooks/memory-sync.sh add-decision "Decision made"
```

### adr-index.sh

Auto-generates the ADR index when decision files are modified.

**Trigger:** Any `.md` file in `.claude/decisions/` is edited.

**Output:** Updates `.claude/decisions/index.md` with:
- Status counts (Accepted, Proposed, Deprecated, Superseded)
- Table of all ADRs with title, status, and date
- Sorted by ADR number

## Enabling Hooks

Copy the settings template to enable all hooks:

```bash
cp .claude/settings.local.json.template .claude/settings.local.json
```

Or add hooks manually in Claude Code settings:

```json
{
  "hooks": {
    "PostEdit": [
      {
        "matcher": "backend/internal/services/*_service.go",
        "command": ".claude/hooks/go-service-pattern.sh \"$FILE\" \"$CONTENT\""
      },
      {
        "matcher": "frontend/app/lib/query-keys.ts",
        "command": ".claude/hooks/query-key-pattern.sh \"$FILE\" \"$CONTENT\""
      },
      {
        "matcher": "frontend/app/hooks/**/*.ts",
        "command": ".claude/hooks/hook-naming.sh \"$FILE\" \"$CONTENT\""
      }
    ]
  }
}
```

## Troubleshooting

### Hook not running

1. Check file path matches the matcher pattern
2. Verify hook is executable: `chmod +x .claude/hooks/*.sh`
3. Check settings.local.json exists and has correct syntax

### False positives

Some hooks may warn about valid patterns. Common cases:

- **go-service-pattern.sh**: Services with repository injection instead of direct `db *gorm.DB`
- **hook-naming.sh**: Hooks that intentionally deviate from naming convention (e.g., utility hooks)
- **query-key-pattern.sh**: Simple query keys that don't need full factory pattern

Warnings are informational and don't block edits. Only errors (exit code 1) block.

### Testing hooks manually

```bash
# Test go-service-pattern
cat backend/internal/services/user_service.go | \
  .claude/hooks/go-service-pattern.sh "backend/internal/services/user_service.go" "$(cat backend/internal/services/user_service.go)"

# Test query-key-pattern
.claude/hooks/query-key-pattern.sh "frontend/app/lib/query-keys.ts" "$(cat frontend/app/lib/query-keys.ts)"

# Test hook-naming
.claude/hooks/hook-naming.sh "frontend/app/hooks/queries/use-users.ts" "$(cat frontend/app/hooks/queries/use-users.ts)"
```
