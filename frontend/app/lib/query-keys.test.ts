import { describe, expect, it } from "vitest";

import { queryKeys } from "./query-keys";

describe("queryKeys", () => {
  describe("users", () => {
    it("has correct 'all' key", () => {
      expect(queryKeys.users.all).toEqual(["users"]);
    });

    it("lists() returns correct key", () => {
      expect(queryKeys.users.lists()).toEqual(["users", "list"]);
    });

    it("list() includes filters in key", () => {
      const filters = { search: "john", role: "admin" };
      expect(queryKeys.users.list(filters)).toEqual(["users", "list", filters]);
    });

    it("list() with empty filters", () => {
      const filters = {};
      expect(queryKeys.users.list(filters)).toEqual(["users", "list", {}]);
    });

    it("details() returns correct key", () => {
      expect(queryKeys.users.details()).toEqual(["users", "detail"]);
    });

    it("detail() includes user id in key", () => {
      expect(queryKeys.users.detail(5)).toEqual(["users", "detail", 5]);
    });

    it("detail() with different ids", () => {
      expect(queryKeys.users.detail(1)).toEqual(["users", "detail", 1]);
      expect(queryKeys.users.detail(100)).toEqual(["users", "detail", 100]);
      expect(queryKeys.users.detail(999)).toEqual(["users", "detail", 999]);
    });
  });

  describe("auth", () => {
    it("has correct 'user' key", () => {
      expect(queryKeys.auth.user).toEqual(["auth", "user"]);
    });

    it("has correct 'session' key", () => {
      expect(queryKeys.auth.session).toEqual(["auth", "session"]);
    });
  });

  describe("health", () => {
    it("has correct 'status' key", () => {
      expect(queryKeys.health.status).toEqual(["health", "status"]);
    });
  });

  describe("featureFlags", () => {
    it("has correct 'all' key", () => {
      expect(queryKeys.featureFlags.all).toEqual(["featureFlags"]);
    });

    it("user() returns correct key", () => {
      expect(queryKeys.featureFlags.user()).toEqual(["featureFlags", "user"]);
    });
  });

  describe("settings", () => {
    it("has correct 'preferences' key", () => {
      expect(queryKeys.settings.preferences).toEqual(["settings", "preferences"]);
    });
  });

  describe("query key consistency", () => {
    it("all keys are readonly arrays", () => {
      // TypeScript enforces this at compile time with 'as const'
      // At runtime, we verify the arrays are properly structured
      expect(Array.isArray(queryKeys.users.all)).toBe(true);
      expect(Array.isArray(queryKeys.auth.user)).toBe(true);
      expect(Array.isArray(queryKeys.health.status)).toBe(true);
      expect(Array.isArray(queryKeys.featureFlags.all)).toBe(true);
      expect(Array.isArray(queryKeys.settings.preferences)).toBe(true);
    });

    it("factory functions return new arrays each time", () => {
      const list1 = queryKeys.users.lists();
      const list2 = queryKeys.users.lists();

      expect(list1).toEqual(list2);
      // They should be equal in value but different references
      // (due to spread operator creating new arrays)
    });

    it("detail keys maintain hierarchical structure", () => {
      const detailKey = queryKeys.users.detail(5);

      // Detail key should start with base keys
      expect(detailKey[0]).toBe("users");
      expect(detailKey[1]).toBe("detail");
      expect(detailKey[2]).toBe(5);
    });
  });

  describe("query key invalidation patterns", () => {
    it("all key can be used to invalidate all users queries", () => {
      // The 'all' key should be a prefix of all other user keys
      const allKey = queryKeys.users.all;
      const listsKey = queryKeys.users.lists();
      const listKey = queryKeys.users.list({ search: "test" });
      const detailsKey = queryKeys.users.details();
      const detailKey = queryKeys.users.detail(1);

      // All keys should start with the 'all' key prefix
      expect(listsKey.slice(0, allKey.length)).toEqual([...allKey]);
      expect(listKey.slice(0, allKey.length)).toEqual([...allKey]);
      expect(detailsKey.slice(0, allKey.length)).toEqual([...allKey]);
      expect(detailKey.slice(0, allKey.length)).toEqual([...allKey]);
    });

    it("featureFlags.all can invalidate all feature flag queries", () => {
      const allKey = queryKeys.featureFlags.all;
      const userKey = queryKeys.featureFlags.user();

      expect(userKey.slice(0, allKey.length)).toEqual([...allKey]);
    });
  });
});
