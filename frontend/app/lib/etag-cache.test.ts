import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  clearAllETags,
  clearETag,
  createETagQueryFn,
  fetchWithETag,
  getStoredETag,
  isNotModifiedError,
  NotModifiedError,
  queryKeyToETagKey,
  storeETag,
} from "./etag-cache";

function mockResponse(status: number, headers: Record<string, string> = {}) {
  return {
    status,
    headers: new Headers(headers),
    json: () => Promise.resolve({ ok: true }),
  } as unknown as Response;
}

describe("etag cache storage", () => {
  beforeEach(() => clearAllETags());

  it("stores and retrieves an ETag by key", () => {
    expect(getStoredETag("k1")).toBeUndefined();
    storeETag("k1", 'W/"abc"');
    expect(getStoredETag("k1")).toBe('W/"abc"');
  });

  it("removes a single ETag", () => {
    storeETag("k1", '"a"');
    storeETag("k2", '"b"');
    clearETag("k1");
    expect(getStoredETag("k1")).toBeUndefined();
    expect(getStoredETag("k2")).toBe('"b"');
  });

  it("clears all ETags", () => {
    storeETag("k1", '"a"');
    storeETag("k2", '"b"');
    clearAllETags();
    expect(getStoredETag("k1")).toBeUndefined();
    expect(getStoredETag("k2")).toBeUndefined();
  });

  it("serializes query keys deterministically", () => {
    expect(queryKeyToETagKey(["users", { page: 2 }])).toBe('["users",{"page":2}]');
    storeETag(queryKeyToETagKey(["users", { page: 2 }]), '"q"');
    expect(getStoredETag('["users",{"page":2}]')).toBe('"q"');
  });
});

describe("NotModifiedError", () => {
  it("identifies NotModifiedError instances only", () => {
    expect(isNotModifiedError(new NotModifiedError())).toBe(true);
    expect(isNotModifiedError(new Error("Not Modified"))).toBe(false);
    expect(isNotModifiedError(undefined)).toBe(false);
    const err = new NotModifiedError();
    expect(err.name).toBe("NotModifiedError");
    expect(err.message).toBe("Not Modified");
  });
});

describe("fetchWithETag", () => {
  beforeEach(() => {
    clearAllETags();
    vi.restoreAllMocks();
  });

  it("sends If-None-Match when an ETag is cached and stores new ETags", async () => {
    storeETag("key", '"v1"');
    const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => {
      void _url;
      void init;
      return mockResponse(200, { ETag: '"v2"' });
    });
    vi.stubGlobal("fetch", fetchMock);

    const res = await fetchWithETag("/api/data", "key");

    expect(res.status).toBe(200);
    const call = fetchMock.mock.calls[0];
    expect(new Headers(call[1]?.headers).get("If-None-Match")).toBe('"v1"');
    expect(getStoredETag("key")).toBe('"v2"');
  });

  it("throws NotModifiedError on 304", async () => {
    storeETag("key", '"v1"');
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(304))
    );

    await expect(fetchWithETag("/api/data", "key")).rejects.toBeInstanceOf(NotModifiedError);
  });

  it("leaves the cache untouched when the response has no ETag", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(200))
    );

    await fetchWithETag("/api/data", "missing");

    expect(getStoredETag("missing")).toBeUndefined();
  });
});

describe("createETagQueryFn", () => {
  beforeEach(() => clearAllETags());

  it("returns parsed JSON and stores the ETag", async () => {
    const fetchFn = vi.fn(async () => mockResponse(200, { ETag: '"v9"' }));
    const queryFn = createETagQueryFn(fetchFn, "qkey", () => undefined);

    await expect(queryFn()).resolves.toEqual({ ok: true });
    expect(getStoredETag("qkey")).toBe('"v9"');
  });

  it("returns previous data when the fetcher throws NotModifiedError", async () => {
    const fetchFn = vi.fn(async () => {
      throw new NotModifiedError();
    });
    const queryFn = createETagQueryFn(fetchFn, "qkey", () => ({ cached: true }));

    await expect(queryFn()).resolves.toEqual({ cached: true });
  });

  it("rethrows other errors", async () => {
    const boom = new Error("boom");
    const queryFn = createETagQueryFn(
      async () => {
        throw boom;
      },
      "qkey",
      () => ({ cached: true })
    );

    await expect(queryFn()).rejects.toBe(boom);
  });

  it("rethrows NotModifiedError when no previous data exists", async () => {
    const queryFn = createETagQueryFn(
      async () => {
        throw new NotModifiedError();
      },
      "qkey",
      () => undefined
    );

    await expect(queryFn()).rejects.toBeInstanceOf(NotModifiedError);
  });
});
