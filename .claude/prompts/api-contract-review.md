# API Contract Review - Go ↔ TypeScript Alignment

Use this prompt when reviewing API contract alignment between backend and frontend.

---

## Overview

This stack uses:
- **Backend**: Go structs with JSON tags → HTTP responses
- **Frontend**: TypeScript interfaces → TanStack Query hooks

Misalignment causes runtime errors that TypeScript can't catch at compile time.

---

## Review Checklist

### 1. Response Types Match

Compare Go handler response with TypeScript interface:

**Go (handler response):**
```go
type UserResponse struct {
    ID        uint      `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**TypeScript (frontend type):**
```typescript
interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  createdAt: string;  // dates come as ISO strings
  updatedAt: string;
}
```

**Check:**
```
[ ] Field names match (Go json tags = TS keys)
[ ] Field types compatible (see type mapping below)
[ ] Optional fields marked correctly (omitempty ↔ ?)
[ ] All fields present (no extra/missing)
```

### 2. Type Mapping

| Go Type | JSON | TypeScript |
|---------|------|------------|
| `string` | `"text"` | `string` |
| `int`, `uint`, `int64` | `123` | `number` |
| `float64` | `1.23` | `number` |
| `bool` | `true` | `boolean` |
| `time.Time` | `"2024-01-01T00:00:00Z"` | `string` (ISO 8601) |
| `[]T` | `[...]` | `T[]` |
| `*T` (pointer) | `null` or value | `T \| null` |
| `map[string]T` | `{...}` | `Record<string, T>` |
| `interface{}` | any | `unknown` |

### 3. Nullable vs Optional

**Go:**
```go
// Nullable (can be null in JSON)
Name *string `json:"name"`

// Optional (omitted if zero value)
Name string `json:"name,omitempty"`

// Required and nullable
Name *string `json:"name"`

// Optional and nullable
Name *string `json:"name,omitempty"`
```

**TypeScript:**
```typescript
// Nullable
name: string | null;

// Optional (might not be present)
name?: string;

// Optional and nullable
name?: string | null;
```

### 4. Error Response Format

Verify error responses match:

**Go:**
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
    Code    string `json:"code,omitempty"`
}
```

**TypeScript:**
```typescript
interface ApiError {
  error: string;
  message?: string;
  code?: string;
}
```

### 5. Query Key Alignment

Check query keys match endpoints:

| Endpoint | Query Key |
|----------|-----------|
| `GET /api/v1/users` | `queryKeys.users.all` |
| `GET /api/v1/users/:id` | `queryKeys.users.detail(id)` |
| `GET /api/v1/orgs/:slug` | `queryKeys.organizations.detail(slug)` |
| `GET /api/v1/orgs/:slug/members` | `queryKeys.organizations.members(slug)` |

---

## Common Misalignment Patterns

### 1. Case Mismatch

```go
// Go uses camelCase in json tags
Name string `json:"userName"`  // ❌ should be "name" or match TS
```

```typescript
// TS expects what Go sends
interface User {
  userName: string;  // Must match json tag exactly
}
```

### 2. Date Handling

```go
// Go sends ISO string
CreatedAt time.Time `json:"createdAt"`
```

```typescript
// TS receives string, not Date object
interface User {
  createdAt: string;  // ✅ string
  // createdAt: Date; // ❌ won't work without parsing
}

// Parse when needed:
const date = new Date(user.createdAt);
```

### 3. Enum Misalignment

```go
type Role string
const (
    RoleAdmin  Role = "admin"
    RoleMember Role = "member"
)
```

```typescript
type Role = "admin" | "member";  // Must match Go constants
// OR
enum Role {
  Admin = "admin",
  Member = "member"
}
```

### 4. Nested Objects

```go
type Order struct {
    ID    uint     `json:"id"`
    Items []Item   `json:"items"`
    User  *User    `json:"user,omitempty"`
}
```

```typescript
interface Order {
  id: number;
  items: Item[];       // Array, not Items
  user?: User | null;  // omitempty + pointer = optional + nullable
}
```

---

## Review Process

### Step 1: Find All Response Types

```bash
# Go response structs
grep -r "json:\"" backend/internal/handlers/ backend/internal/models/

# TypeScript API types
cat frontend/app/services/types/index.ts
```

### Step 2: Compare Each Pair

For each endpoint:
1. Find Go handler function
2. Identify response struct
3. Find corresponding TypeScript type
4. Check all fields match

### Step 3: Check Query Keys

```bash
# View query keys
cat frontend/app/lib/query-keys.ts

# View hooks using them
ls frontend/app/hooks/queries/
```

### Step 4: Verify Mutations

Check mutation request bodies match Go input structs:

```go
// Go expects
type CreateUserInput struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required"`
}
```

```typescript
// TS sends
interface CreateUserInput {
  email: string;
  password: string;
  name: string;
}
```

---

## Output Format

```markdown
## API Contract Review: {Feature/Module}

### Aligned ✅
- `User` - Go and TypeScript match
- `Organization` - All fields aligned

### Misaligned ❌

#### 1. UserProfile
**Go:**
\`\`\`go
AvatarURL *string `json:"avatarUrl,omitempty"`
\`\`\`

**TypeScript:**
\`\`\`typescript
avatarUrl: string;  // Missing optional marker
\`\`\`

**Fix:** Change TypeScript to `avatarUrl?: string | null;`

#### 2. Order.createdAt
**Issue:** Go sends ISO string, TS expects Date object
**Fix:** Update TS type to `createdAt: string`

### Query Keys
- ✅ `users.all` → `GET /api/v1/users`
- ❌ `orders.list` → endpoint is `/api/v1/orders` (plural mismatch)

### Recommendations
1. Run `/sync-models` to regenerate TypeScript types
2. Update query key for orders endpoint
```

---

## Automation

Use `/sync-models` skill to auto-generate TypeScript types from Go structs.
