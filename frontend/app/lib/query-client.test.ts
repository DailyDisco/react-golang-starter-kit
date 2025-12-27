import { describe, expect, it } from "vitest";

import { queryClient } from "./query-client";

/**
 * Tests for query client retry logic.
 *
 * The retry function should:
 * - NOT retry on 4xx client errors (400, 401, 403, 404, etc.)
 * - Retry on 5xx server errors (500, 502, 503, 504, etc.)
 * - Retry on network errors
 * - Stop after 2 retries (failureCount >= 2)
 */
describe("queryClient", () => {
  describe("default query options", () => {
    it("should have staleTime of 1 minute", () => {
      const options = queryClient.getDefaultOptions().queries;
      expect(options?.staleTime).toBe(60000);
    });

    it("should have gcTime of 5 minutes", () => {
      const options = queryClient.getDefaultOptions().queries;
      expect(options?.gcTime).toBe(300000);
    });

    it("should refetch on window focus", () => {
      const options = queryClient.getDefaultOptions().queries;
      expect(options?.refetchOnWindowFocus).toBe(true);
    });

    it("should refetch on reconnect", () => {
      const options = queryClient.getDefaultOptions().queries;
      expect(options?.refetchOnReconnect).toBe(true);
    });
  });

  describe("retry logic", () => {
    // Get the retry function from query options
    const getRetryFn = () => {
      const options = queryClient.getDefaultOptions().queries;
      return options?.retry as (failureCount: number, error: Error) => boolean;
    };

    describe("4xx client errors - should NOT retry", () => {
      it("should not retry on 400 Bad Request", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 400");
        expect(retryFn(0, error)).toBe(false);
      });

      it("should not retry on 401 Unauthorized", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 401");
        expect(retryFn(0, error)).toBe(false);
      });

      it("should not retry on 403 Forbidden", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 403");
        expect(retryFn(0, error)).toBe(false);
      });

      it("should not retry on 404 Not Found", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 404");
        expect(retryFn(0, error)).toBe(false);
      });

      it("should not retry on 422 Unprocessable Entity", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 422");
        expect(retryFn(0, error)).toBe(false);
      });

      it("should not retry on 429 Too Many Requests", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 429");
        expect(retryFn(0, error)).toBe(false);
      });
    });

    describe("5xx server errors - should retry", () => {
      it("should retry on 500 Internal Server Error", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 500");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on 502 Bad Gateway", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 502");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on 503 Service Unavailable", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 503");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on 504 Gateway Timeout", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 504");
        // This is a key test - 504 should retry but currently broken
        expect(retryFn(0, error)).toBe(true);
      });
    });

    describe("network errors - should retry", () => {
      it("should retry on network failure", () => {
        const retryFn = getRetryFn();
        const error = new Error("Failed to fetch");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on timeout", () => {
        const retryFn = getRetryFn();
        const error = new Error("The operation was aborted");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on DNS failure", () => {
        const retryFn = getRetryFn();
        const error = new Error("DNS resolution failed");
        expect(retryFn(0, error)).toBe(true);
      });
    });

    describe("retry count limits", () => {
      it("should retry on first failure (failureCount=0)", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 500");
        expect(retryFn(0, error)).toBe(true);
      });

      it("should retry on second failure (failureCount=1)", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 500");
        expect(retryFn(1, error)).toBe(true);
      });

      it("should NOT retry after 2 failures (failureCount=2)", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 500");
        expect(retryFn(2, error)).toBe(false);
      });

      it("should NOT retry after 3 failures (failureCount=3)", () => {
        const retryFn = getRetryFn();
        const error = new Error("Request failed with status 500");
        expect(retryFn(3, error)).toBe(false);
      });
    });

    describe("edge cases", () => {
      it("should handle non-Error objects gracefully", () => {
        const retryFn = getRetryFn();
        // Some libraries throw non-Error objects
        const error = { message: "Request failed with status 500" } as Error;
        // Should default to retry since we can't determine status
        expect(retryFn(0, error)).toBe(true);
      });

      it("should handle errors without message", () => {
        const retryFn = getRetryFn();
        const error = new Error();
        // Should default to retry
        expect(retryFn(0, error)).toBe(true);
      });
    });
  });

  describe("mutation options", () => {
    it("should not retry mutations by default", () => {
      const options = queryClient.getDefaultOptions().mutations;
      expect(options?.retry).toBe(false);
    });

    it("should use online network mode for mutations", () => {
      const options = queryClient.getDefaultOptions().mutations;
      expect(options?.networkMode).toBe("online");
    });
  });
});
