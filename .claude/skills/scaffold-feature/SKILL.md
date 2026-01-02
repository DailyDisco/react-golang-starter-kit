---
name: scaffold-feature
description: End-to-end feature scaffolding for backend and frontend. Use when adding a complete new feature spanning Go service and React hooks.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(go:*), Bash(npm:*), AskUserQuestion, Skill
requires:
  - scaffold-service
context-files:
  - frontend/app/lib/query-keys.ts
  - frontend/app/hooks/queries/use-users.ts
  - frontend/app/hooks/mutations/use-org-mutations.ts
  - backend/internal/services/org_service.go
---

# Full-Stack Feature Scaffolding

You are my full-stack feature scaffolding assistant for this React + Go starter kit.

## Objective

Generate a complete feature spanning backend and frontend:
- Go service + handlers + tests (via /scaffold-service)
- TypeScript types matching Go structs
- Query keys following factory pattern
- API service with typed methods
- TanStack Query hooks
- Mutation hooks with cache invalidation
- Frontend tests

## Hard Rules

1. Backend first, then frontend (types depend on Go structs)
2. Use `/scaffold-service` for backend (don't duplicate)
3. Query keys MUST follow the factory pattern in `query-keys.ts`
4. All hooks must use `queryKeys.{entity}` for cache management
5. Mutations must invalidate related queries on success
6. Frontend tests must cover hook behavior
7. Run typecheck and tests before completing

## Guided Workflow

### Phase 1: Gather Requirements

Ask the user:

1. **Feature name** (singular, lowercase): e.g., "notification", "payment"
2. **Feature description**: Brief explanation of what it does
3. **Operations needed**:
   - [ ] Create
   - [ ] Read (single)
   - [ ] List (with filters)
   - [ ] Update
   - [ ] Delete
   - [ ] Custom (describe)
4. **Model fields**: List the fields for the entity

Example:
```
Feature: notification
Description: User notification system for alerts and updates
Operations: Create, Read, List, MarkAsRead, Delete
Fields:
  - id: number
  - userId: number
  - title: string
  - message: string
  - type: "info" | "warning" | "error"
  - read: boolean
  - createdAt: string
```

### Phase 2: Backend Service

Invoke `/scaffold-service` with the gathered requirements.

This creates:
- `backend/internal/services/{entity}_service.go`
- `backend/internal/services/{entity}_service_test.go`
- `backend/internal/handlers/{entity}_handlers.go`

### Phase 3: TypeScript Types

Create `frontend/app/types/{entity}.ts`:

```typescript
// Generated from Go model
export interface {Entity} {
  id: number;
  // ... fields from requirements
  createdAt: string;
  updatedAt: string;
}

export interface Create{Entity}Request {
  // ... writable fields only
}

export interface Update{Entity}Request {
  // ... updatable fields only
}

export interface {Entity}ListResponse {
  data: {Entity}[];
  meta: {
    total: number;
    page: number;
    perPage: number;
  };
}

export interface {Entity}Filters {
  // ... filter options
}
```

### Phase 4: Query Keys

Add to `frontend/app/lib/query-keys.ts`:

```typescript
// Add inside queryKeys object
{entity}: {
  all: ["{entity}"] as const,
  lists: () => [...queryKeys.{entity}.all, "list"] as const,
  list: (filters: Record<string, unknown>) => [...queryKeys.{entity}.lists(), filters] as const,
  details: () => [...queryKeys.{entity}.all, "detail"] as const,
  detail: (id: number) => [...queryKeys.{entity}.details(), id] as const,
},
```

### Phase 5: API Service

Create `frontend/app/services/{entity}/{entity}Service.ts`:

```typescript
import { apiClient } from "../api/client";
import type {
  {Entity},
  Create{Entity}Request,
  Update{Entity}Request,
  {Entity}ListResponse,
  {Entity}Filters,
} from "../../types/{entity}";

const BASE_PATH = "/api/v1/{entities}";

export const {Entity}Service = {
  async getAll(filters?: {Entity}Filters): Promise<{Entity}ListResponse> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) params.append(key, String(value));
      });
    }
    const url = params.toString() ? `${BASE_PATH}?${params}` : BASE_PATH;
    return apiClient.get<{Entity}ListResponse>(url);
  },

  async getById(id: number): Promise<{Entity}> {
    return apiClient.get<{Entity}>(`${BASE_PATH}/${id}`);
  },

  async create(data: Create{Entity}Request): Promise<{Entity}> {
    return apiClient.post<{Entity}>(BASE_PATH, data);
  },

  async update(id: number, data: Update{Entity}Request): Promise<{Entity}> {
    return apiClient.put<{Entity}>(`${BASE_PATH}/${id}`, data);
  },

  async delete(id: number): Promise<void> {
    return apiClient.delete(`${BASE_PATH}/${id}`);
  },
};
```

Also create `frontend/app/services/{entity}/index.ts`:
```typescript
export * from "./{entity}Service";
```

### Phase 6: Query Hooks

Create `frontend/app/hooks/queries/use-{entity}.ts`:

```typescript
import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { {Entity}Service } from "../../services/{entity}";
import type { {Entity}Filters } from "../../types/{entity}";

export const use{Entities} = (filters?: {Entity}Filters) => {
  return useQuery({
    queryKey: queryKeys.{entity}.list(filters ?? {}),
    queryFn: () => {Entity}Service.getAll(filters),
  });
};

export const use{Entity} = (id: number | undefined) => {
  return useQuery({
    queryKey: queryKeys.{entity}.detail(id!),
    queryFn: () => {Entity}Service.getById(id!),
    enabled: !!id,
  });
};
```

### Phase 7: Mutation Hooks

Create `frontend/app/hooks/mutations/use-{entity}-mutations.ts`:

```typescript
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { queryKeys } from "../../lib/query-keys";
import { {Entity}Service } from "../../services/{entity}";
import type { Create{Entity}Request, Update{Entity}Request } from "../../types/{entity}";

export function useCreate{Entity}() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Create{Entity}Request) => {Entity}Service.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.{entity}.all });
      toast.success("{Entity} created successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useUpdate{Entity}() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Update{Entity}Request }) =>
      {Entity}Service.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.{entity}.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.{entity}.detail(id) });
      toast.success("{Entity} updated successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useDelete{Entity}() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => {Entity}Service.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.{entity}.all });
      toast.success("{Entity} deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}
```

### Phase 8: Hook Tests

Create `frontend/app/hooks/queries/use-{entity}.test.ts`:

```typescript
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { use{Entity}, use{Entities} } from "./use-{entity}";
import { {Entity}Service } from "../../services/{entity}";
import { createQueryWrapper } from "../../test-utils/query-wrapper";

vi.mock("../../services/{entity}");

describe("use{Entity}", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("fetches {entity} by id", async () => {
    const mock{Entity} = { id: 1, title: "Test" };
    vi.mocked({Entity}Service.getById).mockResolvedValue(mock{Entity});

    const { result } = renderHook(() => use{Entity}(1), {
      wrapper: createQueryWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mock{Entity});
  });

  it("does not fetch when id is undefined", () => {
    const { result } = renderHook(() => use{Entity}(undefined), {
      wrapper: createQueryWrapper(),
    });

    expect(result.current.isFetching).toBe(false);
    expect({Entity}Service.getById).not.toHaveBeenCalled();
  });
});

describe("use{Entities}", () => {
  it("fetches all {entities}", async () => {
    const mockResponse = { data: [{ id: 1 }], meta: { total: 1, page: 1, perPage: 10 } };
    vi.mocked({Entity}Service.getAll).mockResolvedValue(mockResponse);

    const { result } = renderHook(() => use{Entities}(), {
      wrapper: createQueryWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockResponse);
  });
});
```

### Phase 9: Update Exports

Add to `frontend/app/hooks/queries/index.ts`:
```typescript
export * from "./use-{entity}";
```

Add to `frontend/app/hooks/mutations/index.ts`:
```typescript
export * from "./use-{entity}-mutations";
```

Add to `frontend/app/services/index.ts`:
```typescript
export * from "./{entity}";
```

### Phase 10: Verification

Run verification commands:

```bash
# Frontend
cd frontend && npm run typecheck
cd frontend && npm run test -- --run use-{entity}

# Backend
cd backend && go build ./...
cd backend && go test ./internal/services/... -run {Entity}
```

## Output Format

After each phase, show:
```
## Phase N Complete: {phase name}
- Created/Modified: {file paths}
- Next: {what's next}
```

At the end:
```
## Feature Scaffolding Complete

Backend files:
- backend/internal/services/{entity}_service.go
- backend/internal/services/{entity}_service_test.go
- backend/internal/handlers/{entity}_handlers.go

Frontend files:
- frontend/app/types/{entity}.ts
- frontend/app/services/{entity}/{entity}Service.ts
- frontend/app/services/{entity}/index.ts
- frontend/app/hooks/queries/use-{entity}.ts
- frontend/app/hooks/queries/use-{entity}.test.ts
- frontend/app/hooks/mutations/use-{entity}-mutations.ts

Modified files:
- frontend/app/lib/query-keys.ts
- frontend/app/hooks/queries/index.ts
- frontend/app/hooks/mutations/index.ts
- frontend/app/services/index.ts

Next steps:
1. Add model to backend/internal/models/models.go
2. Add route registration to backend/cmd/main.go
3. Run database migration if needed
4. Build UI components using the hooks

Verification:
- [ ] TypeScript: passing
- [ ] Frontend tests: passing
- [ ] Backend build: passing
- [ ] Backend tests: passing
```

## Constraints

- Entity names must be lowercase, singular
- Feature must have at least Read or List operation
- Max 10 operations per feature
- All generated code must pass typecheck and tests
