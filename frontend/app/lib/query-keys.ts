// Query Keys Factory for Type-Safe Query Management
export const queryKeys = {
  users: {
    all: ["users"] as const,
    lists: () => [...queryKeys.users.all, "list"] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.users.lists(), filters] as const,
    details: () => [...queryKeys.users.all, "detail"] as const,
    detail: (id: number) => [...queryKeys.users.details(), id] as const,
  },
  auth: {
    user: ["auth", "user"] as const,
    session: ["auth", "session"] as const,
  },
  health: {
    status: ["health", "status"] as const,
  },
  featureFlags: {
    all: ["featureFlags"] as const,
    user: () => [...queryKeys.featureFlags.all, "user"] as const,
  },
  settings: {
    preferences: ["settings", "preferences"] as const,
  },
} as const;

// Type-safe query key inference
export type QueryKeys = typeof queryKeys;
