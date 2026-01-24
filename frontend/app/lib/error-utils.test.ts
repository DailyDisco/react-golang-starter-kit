import { describe, expect, it } from "vitest";

import {
  categorizeError,
  isAuthError,
  isNetworkError,
  isRetryableError,
  isServerError,
  isValidationError,
  type CategorizedError,
} from "./error-utils";

// Mock ApiError class for testing
class MockApiError extends Error {
  code: string;
  statusCode: number;

  constructor(message: string, code: string, statusCode: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.statusCode = statusCode;
  }
}

describe("categorizeError", () => {
  describe("auth errors", () => {
    it("categorizes 401 as auth error", () => {
      const error = new MockApiError("Unauthorized", "UNAUTHORIZED", 401);
      const result = categorizeError(error);

      expect(result.category).toBe("auth");
      expect(result.message).toBe("Please log in to continue");
      expect(result.retryable).toBe(false);
      expect(result.statusCode).toBe(401);
    });

    it("categorizes 403 as auth error with permission message", () => {
      const error = new MockApiError("Forbidden", "FORBIDDEN", 403);
      const result = categorizeError(error);

      expect(result.category).toBe("auth");
      expect(result.message).toBe("You don't have permission for this action");
      expect(result.retryable).toBe(false);
      expect(result.statusCode).toBe(403);
    });

    it("preserves original message in details for auth errors", () => {
      const error = new MockApiError("Token expired", "TOKEN_EXPIRED", 401);
      const result = categorizeError(error);

      expect(result.details).toBe("Token expired");
      expect(result.code).toBe("TOKEN_EXPIRED");
    });
  });

  describe("validation errors", () => {
    it("categorizes 400 as validation error", () => {
      const error = new MockApiError("Invalid email format", "VALIDATION_ERROR", 400);
      const result = categorizeError(error);

      expect(result.category).toBe("validation");
      expect(result.message).toBe("Invalid email format");
      expect(result.retryable).toBe(false);
      expect(result.statusCode).toBe(400);
    });

    it("categorizes 422 as validation error", () => {
      const error = new MockApiError("Email already exists", "DUPLICATE_EMAIL", 422);
      const result = categorizeError(error);

      expect(result.category).toBe("validation");
      expect(result.message).toBe("Email already exists");
      expect(result.retryable).toBe(false);
      expect(result.statusCode).toBe(422);
    });

    it("provides fallback message for empty validation errors", () => {
      const error = new MockApiError("", "VALIDATION_ERROR", 400);
      const result = categorizeError(error);

      expect(result.category).toBe("validation");
      expect(result.message).toBe("Please check your input");
    });
  });

  describe("server errors", () => {
    it("categorizes 500 as server error with retryable true", () => {
      const error = new MockApiError("Internal Server Error", "INTERNAL_ERROR", 500);
      const result = categorizeError(error);

      expect(result.category).toBe("server");
      expect(result.message).toBe("Something went wrong on our end");
      expect(result.retryable).toBe(true);
      expect(result.statusCode).toBe(500);
    });

    it("categorizes 502 as server error", () => {
      const error = new MockApiError("Bad Gateway", "BAD_GATEWAY", 502);
      const result = categorizeError(error);

      expect(result.category).toBe("server");
      expect(result.retryable).toBe(true);
    });

    it("categorizes 503 as server error", () => {
      const error = new MockApiError("Service Unavailable", "SERVICE_UNAVAILABLE", 503);
      const result = categorizeError(error);

      expect(result.category).toBe("server");
      expect(result.retryable).toBe(true);
    });

    it("categorizes 429 rate limit as server error", () => {
      const error = new MockApiError("Rate limit exceeded", "RATE_LIMITED", 429);
      const result = categorizeError(error);

      expect(result.category).toBe("server");
      expect(result.message).toBe("Too many requests. Please wait and try again.");
      expect(result.retryable).toBe(true);
    });
  });

  describe("network errors", () => {
    it("categorizes 'Failed to fetch' as network error", () => {
      const error = new Error("Failed to fetch");
      const result = categorizeError(error);

      expect(result.category).toBe("network");
      expect(result.message).toBe("Unable to connect. Check your internet connection.");
      expect(result.retryable).toBe(true);
    });

    it("categorizes 'NetworkError' as network error", () => {
      const error = new Error("NetworkError when attempting to fetch resource");
      const result = categorizeError(error);

      expect(result.category).toBe("network");
      expect(result.retryable).toBe(true);
    });

    it("categorizes timeout errors as network errors", () => {
      const error = new Error("Request timeout");
      const result = categorizeError(error);

      expect(result.category).toBe("network");
      expect(result.retryable).toBe(true);
    });

    it("categorizes connection refused as network error", () => {
      const error = new Error("ECONNREFUSED");
      const result = categorizeError(error);

      expect(result.category).toBe("network");
      expect(result.retryable).toBe(true);
    });
  });

  describe("unknown errors", () => {
    it("categorizes generic errors as unknown", () => {
      const error = new Error("Something weird happened");
      const result = categorizeError(error);

      expect(result.category).toBe("unknown");
      expect(result.message).toBe("Something weird happened");
      expect(result.retryable).toBe(true);
    });

    it("handles null input", () => {
      const result = categorizeError(null);

      expect(result.category).toBe("unknown");
      expect(result.retryable).toBe(true);
    });

    it("handles undefined input", () => {
      const result = categorizeError(undefined);

      expect(result.category).toBe("unknown");
      expect(result.retryable).toBe(true);
    });

    it("handles string input", () => {
      const result = categorizeError("some error string" as unknown);

      expect(result.category).toBe("unknown");
      expect(result.message).toBe("some error string");
    });

    it("handles empty error message", () => {
      const error = new Error("");
      const result = categorizeError(error);

      expect(result.category).toBe("unknown");
      expect(result.message).toBe("An unexpected error occurred");
    });
  });

  describe("other 4xx errors", () => {
    it("categorizes 404 as validation error", () => {
      const error = new MockApiError("Not found", "NOT_FOUND", 404);
      const result = categorizeError(error);

      expect(result.category).toBe("validation");
      expect(result.retryable).toBe(false);
    });

    it("categorizes 409 conflict as validation error", () => {
      const error = new MockApiError("Resource conflict", "CONFLICT", 409);
      const result = categorizeError(error);

      expect(result.category).toBe("validation");
      expect(result.retryable).toBe(false);
    });
  });
});

describe("error type helpers", () => {
  it("isAuthError returns true for 401", () => {
    const error = new MockApiError("Unauthorized", "UNAUTHORIZED", 401);
    expect(isAuthError(error)).toBe(true);
    expect(isValidationError(error)).toBe(false);
  });

  it("isValidationError returns true for 400", () => {
    const error = new MockApiError("Bad request", "BAD_REQUEST", 400);
    expect(isValidationError(error)).toBe(true);
    expect(isAuthError(error)).toBe(false);
  });

  it("isNetworkError returns true for fetch failures", () => {
    const error = new Error("Failed to fetch");
    expect(isNetworkError(error)).toBe(true);
    expect(isServerError(error)).toBe(false);
  });

  it("isServerError returns true for 500", () => {
    const error = new MockApiError("Server error", "INTERNAL_ERROR", 500);
    expect(isServerError(error)).toBe(true);
  });

  it("isRetryableError returns true for server errors", () => {
    const error = new MockApiError("Server error", "INTERNAL_ERROR", 500);
    expect(isRetryableError(error)).toBe(true);
  });

  it("isRetryableError returns false for validation errors", () => {
    const error = new MockApiError("Invalid input", "VALIDATION_ERROR", 400);
    expect(isRetryableError(error)).toBe(false);
  });
});
