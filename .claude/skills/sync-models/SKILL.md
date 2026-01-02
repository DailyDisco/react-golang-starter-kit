---
name: sync-models
description: Sync Go GORM models to TypeScript types. Use when models change and frontend types need updating.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(npm:*)
context-files:
  - backend/internal/models/models.go
  - backend/internal/models/organization.go
  - frontend/app/types/
---

# Go Model to TypeScript Sync

You are my type synchronization assistant. Keep frontend TypeScript types aligned with backend Go GORM models.

## Project Context

- **Go models**: `backend/internal/models/*.go`
- **TypeScript types**: `frontend/app/types/*.ts`
- **GORM** with JSON tags defines the API contract
- **Zod** schemas for runtime validation (optional)

## Objective

- Parse Go structs and generate matching TypeScript interfaces
- Detect drift between existing TS types and Go models
- Preserve custom TypeScript additions (marked with `// @custom`)
- Generate Zod schemas for runtime validation

## Hard Rules

1. ALWAYS read the Go model before generating types
2. NEVER overwrite lines marked with `// @custom`
3. ALWAYS use the `json` tag for field names (not Go field names)
4. Map Go types to TypeScript correctly (see type mapping)
5. Include JSDoc comments from Go comments
6. Generated sections must have `// AUTO-GENERATED` markers

## Type Mapping

| Go Type | TypeScript Type | Notes |
|---------|-----------------|-------|
| `string` | `string` | |
| `int`, `int32`, `int64`, `uint`, `uint32`, `uint64` | `number` | |
| `float32`, `float64` | `number` | |
| `bool` | `boolean` | |
| `time.Time` | `string` | ISO 8601 format |
| `*T` (pointer) | `T \| null` | Nullable |
| `[]T` | `T[]` | Array |
| `map[string]T` | `Record<string, T>` | |
| `json.RawMessage`, `datatypes.JSON` | `unknown` | Or specific type if known |
| `gorm.DeletedAt` | `string \| null` | Soft delete timestamp |
| `uuid.UUID` | `string` | UUID string |

## Workflow

### Phase 1: Parse Go Models

Read the Go model files and extract:
- Struct name
- Fields with types
- JSON tags (for field names)
- GORM tags (for relationships)
- Comments (for documentation)

Example Go model:
```go
// User represents a user account
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Email     string         `gorm:"uniqueIndex;not null" json:"email"`
    Name      string         `json:"name"`
    AvatarURL *string        `json:"avatarUrl,omitempty"`
    Role      string         `gorm:"default:user" json:"role"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

    // Relationships
    Files []File `gorm:"foreignKey:UserID" json:"files,omitempty"`
}
```

### Phase 2: Generate TypeScript Interface

```typescript
// =============================================================================
// AUTO-GENERATED from backend/internal/models/models.go
// Generated: 2025-01-15T10:30:00Z
// DO NOT EDIT between AUTO-GENERATED markers
// =============================================================================

/** User represents a user account */
export interface User {
  id: number;
  email: string;
  name: string;
  avatarUrl?: string | null;
  role: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
  files?: File[];
}

// =============================================================================
// END AUTO-GENERATED
// =============================================================================

// @custom - Add custom types below this line
export interface UserWithStats extends User {
  totalFiles: number;
  storageUsed: number;
}
```

### Phase 3: Generate Request/Response Types

```typescript
/** Request body for creating a user */
export interface CreateUserRequest {
  email: string;
  name: string;
  password: string;
}

/** Request body for updating a user */
export interface UpdateUserRequest {
  name?: string;
  avatarUrl?: string | null;
}

/** API response for user endpoints */
export interface UserResponse {
  data: User;
}

/** API response for user list */
export interface UsersResponse {
  data: User[];
  meta: {
    total: number;
    page: number;
    perPage: number;
  };
}
```

### Phase 4: Generate Zod Schemas (Optional)

```typescript
import { z } from "zod";

export const UserSchema = z.object({
  id: z.number(),
  email: z.string().email(),
  name: z.string(),
  avatarUrl: z.string().url().nullable().optional(),
  role: z.string(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  deletedAt: z.string().datetime().nullable().optional(),
});

export const CreateUserRequestSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  password: z.string().min(8),
});

export type UserFromSchema = z.infer<typeof UserSchema>;
```

### Phase 5: Drift Detection

Compare existing TypeScript types with Go models:

```
## Type Drift Report

Comparing: backend/internal/models/models.go
       vs: frontend/app/types/user.ts

✅ User - In sync (12 fields match)

⚠️  Organization - DRIFT DETECTED
   + planFeatures: Record<string, unknown> (in Go, missing in TS)
   ~ billingEmail: string | null → string (nullability changed)

❌ UsageEvent - Missing in TypeScript
   → Run /sync-models to generate

Recommendation: Regenerate types for Organization, add UsageEvent
```

## Output Format

### When Generating Types

```
## Types Generated

**Source:** backend/internal/models/models.go
**Target:** frontend/app/types/user.ts

**Types created/updated:**
- User (12 fields)
- CreateUserRequest (3 fields)
- UpdateUserRequest (2 fields)

**Preserved custom types:**
- UserWithStats (kept as-is)

**Next steps:**
1. Review generated types
2. Run: `npm run typecheck`
3. Update any affected components
```

### When Detecting Drift

```
## Drift Report

| Model | Status | Action |
|-------|--------|--------|
| User | ✅ In sync | None |
| Organization | ⚠️ Drift | Update TS type |
| UsageEvent | ❌ Missing | Generate new |

**To fix:**
Run `/sync-models` to regenerate affected types.
```

## File Organization

```
frontend/app/types/
├── user.ts           # User, CreateUserRequest, UpdateUserRequest
├── organization.ts   # Organization, Member, Invitation types
├── billing.ts        # Subscription, Plan types
├── file.ts           # File, FileUpload types
├── usage.ts          # UsageEvent, UsagePeriod types
├── index.ts          # Re-exports all types
└── api.ts            # Generic API response wrappers
```

## Constraints

- One Go model file → One TypeScript type file
- Preserve `// @custom` sections during regeneration
- Include generation timestamp for tracking
- Maximum 50 types per file (split if larger)
- Always run typecheck after generation
