import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { FileResponse, StorageStatus } from "../../services";
import { useFileDownload, useFiles, useFileUrl, useStorageStatus } from "./use-files";

// Mock dependencies
vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("../../lib/logger", () => ({
  logger: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("../../services", () => ({
  FileService: {
    fetchFiles: vi.fn(),
    getFileUrl: vi.fn(),
    getStorageStatus: vi.fn(),
    downloadFile: vi.fn(),
  },
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

describe("useFiles", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useFiles).toBeDefined();
    expect(typeof useFiles).toBe("function");
  });

  it("can be mocked to return files data", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    const mockFiles: FileResponse[] = [
      {
        id: 1,
        file_name: "test.pdf",
        file_size: 1024,
        content_type: "application/pdf",
        location: "/files/1",
        storage_type: "s3",
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      },
      {
        id: 2,
        file_name: "image.png",
        file_size: 2048,
        content_type: "image/png",
        location: "/files/2",
        storage_type: "s3",
        created_at: "2024-01-02",
        updated_at: "2024-01-02",
      },
    ];

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: mockFiles,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<FileResponse[], Error>);

    const result = useFiles();
    expect(result.data).toEqual(mockFiles);
    expect(result.isSuccess).toBe(true);
  });

  it("is enabled when user is authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: [],
      isLoading: false,
    } as unknown as UseQueryResult<FileResponse[], Error>);

    useFiles();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: true,
      })
    );
  });

  it("is disabled when user is not authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: false });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as UseQueryResult<FileResponse[], Error>);

    useFiles();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });

  it("passes limit and offset parameters", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: [],
      isLoading: false,
    } as unknown as UseQueryResult<FileResponse[], Error>);

    useFiles(10, 20);

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["files", { limit: 10, offset: 20 }],
      })
    );
  });
});

describe("useFileUrl", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useFileUrl).toBeDefined();
    expect(typeof useFileUrl).toBe("function");
  });

  it("returns file URL when loaded", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockUrl = "https://example.com/files/1";
    vi.mocked(useQuery).mockReturnValue({
      data: mockUrl,
      isLoading: false,
      isSuccess: true,
    } as unknown as UseQueryResult<string, Error>);

    const result = useFileUrl(1);
    expect(result.data).toBe(mockUrl);
  });

  it("is enabled when fileId is valid", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<string, Error>);

    useFileUrl(5);

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: true,
      })
    );
  });

  it("is disabled when fileId is 0 or invalid", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as UseQueryResult<string, Error>);

    useFileUrl(0);

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });
});

describe("useStorageStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useStorageStatus).toBeDefined();
    expect(typeof useStorageStatus).toBe("function");
  });

  it("returns storage status data", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockStatus: StorageStatus = {
      storage_type: "s3",
      message: "Storage is available",
    };

    vi.mocked(useQuery).mockReturnValue({
      data: mockStatus,
      isLoading: false,
      isSuccess: true,
    } as unknown as UseQueryResult<StorageStatus, Error>);

    const result = useStorageStatus();
    expect(result.data).toEqual(mockStatus);
  });

  it("has correct stale time configuration", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<StorageStatus, Error>);

    useStorageStatus();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        staleTime: 30 * 60 * 1000, // 30 minutes
        retry: 1,
      })
    );
  });
});

describe("useFileDownload", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useFileDownload).toBeDefined();
    expect(typeof useFileDownload).toBe("function");
  });

  it("returns an object with downloadFile function", () => {
    const result = useFileDownload();
    expect(result).toHaveProperty("downloadFile");
    expect(typeof result.downloadFile).toBe("function");
  });

  it("downloadFile creates a download link and triggers download", async () => {
    const { FileService } = await import("../../services");
    const { toast } = await import("sonner");

    const mockBlob = new Blob(["test content"], { type: "application/pdf" });
    vi.mocked(FileService.downloadFile).mockResolvedValue(mockBlob);

    // Mock URL methods
    const mockCreateObjectURL = vi.fn(() => "blob:http://localhost/test");
    const mockRevokeObjectURL = vi.fn();
    globalThis.URL.createObjectURL = mockCreateObjectURL;
    globalThis.URL.revokeObjectURL = mockRevokeObjectURL;

    // Mock DOM methods
    const mockLink = {
      href: "",
      download: "",
      click: vi.fn(),
    };
    const mockCreateElement = vi.spyOn(document, "createElement").mockReturnValue(mockLink as unknown as HTMLElement);
    const mockAppendChild = vi
      .spyOn(document.body, "appendChild")
      .mockImplementation(() => mockLink as unknown as Node);
    const mockRemoveChild = vi
      .spyOn(document.body, "removeChild")
      .mockImplementation(() => mockLink as unknown as Node);

    const { downloadFile } = useFileDownload();
    await downloadFile(1, "test.pdf");

    expect(FileService.downloadFile).toHaveBeenCalledWith(1);
    expect(mockCreateObjectURL).toHaveBeenCalledWith(mockBlob);
    expect(mockLink.download).toBe("test.pdf");
    expect(mockLink.click).toHaveBeenCalled();
    expect(mockRevokeObjectURL).toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("File downloaded successfully!");

    // Cleanup
    mockCreateElement.mockRestore();
    mockAppendChild.mockRestore();
    mockRemoveChild.mockRestore();
  });

  it("handles download errors gracefully", async () => {
    const { FileService } = await import("../../services");
    const { toast } = await import("sonner");
    const { logger } = await import("../../lib/logger");

    const mockError = new Error("Download failed");
    vi.mocked(FileService.downloadFile).mockRejectedValue(mockError);

    const { downloadFile } = useFileDownload();
    await downloadFile(1, "test.pdf");

    expect(logger.error).toHaveBeenCalledWith("File download error", mockError);
    expect(toast.error).toHaveBeenCalledWith("Download failed", {
      description: "Please try again later",
    });
  });
});
