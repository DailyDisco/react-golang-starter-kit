import React from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { FileService, type FileResponse } from "../../services";
import { useAuthStore } from "../../stores/auth-store";
// Import the hooks to test
import { useFileDelete, useFileUpload } from "./use-file-mutations";

// Mock the FileService
vi.mock("../../services", () => ({
  FileService: {
    uploadFile: vi.fn(),
    deleteFile: vi.fn(),
  },
}));

// Mock the logger
vi.mock("../../lib/logger", () => ({
  logger: {
    error: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
  },
}));

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock the auth store
vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

// Helper to create wrapper with QueryClientProvider
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe("useFileUpload", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock auth store with valid token
    vi.mocked(useAuthStore).mockReturnValue({
      accessToken: "test-token",
    } as any);
  });

  it("should upload file successfully", async () => {
    const mockFile = new File(["test content"], "test.txt", { type: "text/plain" });
    const mockResponse: FileResponse = {
      id: 1,
      file_name: "test.txt",
      content_type: "text/plain",
      file_size: 12,
      location: "/uploads/test.txt",
      storage_type: "s3",
      created_at: "2024-01-01",
      updated_at: "2024-01-01",
    };
    vi.mocked(FileService.uploadFile).mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockFile);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(FileService.uploadFile).toHaveBeenCalledWith(mockFile, "test-token");
  });

  it("should throw error when no auth token", async () => {
    // Mock auth store without token
    vi.mocked(useAuthStore).mockReturnValue({
      accessToken: null,
    } as any);

    const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockFile);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error?.message).toBe("Authentication token not found.");
  });

  it("should handle upload failure", async () => {
    vi.mocked(FileService.uploadFile).mockRejectedValueOnce(new Error("Upload failed"));

    const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockFile);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should handle file too large error", async () => {
    vi.mocked(FileService.uploadFile).mockRejectedValueOnce(new Error("File size exceeds maximum limit"));

    const mockFile = new File(["test"], "large.txt", { type: "text/plain" });

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockFile);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should handle authentication error", async () => {
    vi.mocked(FileService.uploadFile).mockRejectedValueOnce(new Error("User not found"));

    const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockFile);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should handle async upload flow correctly", async () => {
    const mockResponse: FileResponse = {
      id: 1,
      file_name: "test.txt",
      content_type: "text/plain",
      file_size: 12,
      location: "/uploads/test.txt",
      storage_type: "s3",
      created_at: "2024-01-01",
      updated_at: "2024-01-01",
    };
    vi.mocked(FileService.uploadFile).mockResolvedValueOnce(mockResponse);

    const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    // Initially should not be pending
    expect(result.current.isPending).toBe(false);
    expect(result.current.isSuccess).toBe(false);

    act(() => {
      result.current.mutate(mockFile);
    });

    // Wait for success and verify the flow completed
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(FileService.uploadFile).toHaveBeenCalledWith(mockFile, "test-token");
  });
});

describe("useFileDelete", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock auth store with valid token
    vi.mocked(useAuthStore).mockReturnValue({
      accessToken: "test-token",
    } as any);
  });

  it("should delete file successfully", async () => {
    vi.mocked(FileService.deleteFile).mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useFileDelete(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(FileService.deleteFile).toHaveBeenCalledWith(1, "test-token");
  });

  it("should throw error when no auth token", async () => {
    // Mock auth store without token
    vi.mocked(useAuthStore).mockReturnValue({
      accessToken: null,
    } as any);

    const { result } = renderHook(() => useFileDelete(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error?.message).toBe("Authentication token not found.");
  });

  it("should handle delete failure", async () => {
    vi.mocked(FileService.deleteFile).mockRejectedValueOnce(new Error("File not found"));

    const { result } = renderHook(() => useFileDelete(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(999);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error?.message).toBe("File not found");
  });

  it("should pass the correct file ID to the service", async () => {
    vi.mocked(FileService.deleteFile).mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useFileDelete(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(42);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(FileService.deleteFile).toHaveBeenCalledWith(42, "test-token");
  });
});

describe("Mutation hooks mock compatibility", () => {
  beforeEach(() => {
    vi.mocked(useAuthStore).mockReturnValue({
      accessToken: "test-token",
    } as any);
  });

  it("useFileUpload returns expected interface", () => {
    const { result } = renderHook(() => useFileUpload(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(result.current.mutateAsync).toBeDefined();
    expect(typeof result.current.isPending).toBe("boolean");
    expect(typeof result.current.isError).toBe("boolean");
    expect(typeof result.current.isSuccess).toBe("boolean");
  });

  it("useFileDelete returns expected interface", () => {
    const { result } = renderHook(() => useFileDelete(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(result.current.mutateAsync).toBeDefined();
    expect(typeof result.current.isPending).toBe("boolean");
    expect(typeof result.current.isError).toBe("boolean");
    expect(typeof result.current.isSuccess).toBe("boolean");
  });
});
