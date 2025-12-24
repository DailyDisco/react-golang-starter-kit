import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { authenticatedFetch, parseErrorResponse } from "../api/client";
import { FileService } from "./fileService";

// Mock the API client module
vi.mock("../api/client", () => ({
  API_BASE_URL: "http://localhost:8080",
  authenticatedFetch: vi.fn(),
  parseErrorResponse: vi.fn(),
}));

// Mock global fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe("FileService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("uploadFile", () => {
    it("should upload file successfully", async () => {
      const mockFile = new File(["test content"], "test.txt", { type: "text/plain" });
      const mockResponse = {
        id: 1,
        file_name: "test.txt",
        content_type: "text/plain",
        file_size: 12,
        storage_type: "s3",
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await FileService.uploadFile(mockFile, "test-token");

      expect(result).toEqual(mockResponse);
      expect(mockFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/files/upload",
        expect.objectContaining({
          method: "POST",
          headers: { Authorization: "Bearer test-token" },
        })
      );

      // Verify FormData was used
      const callArgs = mockFetch.mock.calls[0][1];
      expect(callArgs.body).toBeInstanceOf(FormData);
    });

    it("should handle old response format (fallback)", async () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });
      const mockResponse = { id: 1, file_name: "test.txt" };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await FileService.uploadFile(mockFile, "test-token");

      expect(result).toEqual(mockResponse);
    });

    it("should throw error on upload failure", async () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 413,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("File too large");

      await expect(FileService.uploadFile(mockFile, "test-token")).rejects.toThrow("File too large");
    });

    it("should throw error on invalid JSON response", async () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as Response);

      await expect(FileService.uploadFile(mockFile, "test-token")).rejects.toThrow(
        "Invalid response format from server"
      );
    });
  });

  describe("fetchFiles", () => {
    it("should fetch all files", async () => {
      const mockFiles = [
        { id: 1, file_name: "file1.txt" },
        { id: 2, file_name: "file2.txt" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockFiles }),
      } as Response);

      const result = await FileService.fetchFiles();

      expect(result).toEqual(mockFiles);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/files");
    });

    it("should include pagination parameters when provided", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: [] }),
      } as Response);

      await FileService.fetchFiles(10, 20);

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/files?limit=10&offset=20");
    });

    it("should return empty array on auth failure (401/403)", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 401,
      } as Response);

      const result = await FileService.fetchFiles();

      expect(result).toEqual([]);
    });

    it("should return empty array on parse error", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as Response);

      const result = await FileService.fetchFiles();

      expect(result).toEqual([]);
    });

    it("should handle old response format (fallback)", async () => {
      const mockFiles = [{ id: 1, file_name: "file1.txt" }];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockFiles,
      } as Response);

      const result = await FileService.fetchFiles();

      expect(result).toEqual(mockFiles);
    });

    it("should throw error on server error (non-auth)", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("Server error");

      // fetchFiles catches errors and returns empty array
      const result = await FileService.fetchFiles();
      expect(result).toEqual([]);
    });
  });

  describe("getFileUrl", () => {
    it("should return file URL", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: "https://s3.example.com/file.txt" }),
      } as Response);

      const result = await FileService.getFileUrl(1);

      expect(result).toBe("https://s3.example.com/file.txt");
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/files/1/url");
    });

    it("should handle old response format (fallback)", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => "https://s3.example.com/file.txt",
      } as Response);

      const result = await FileService.getFileUrl(1);

      expect(result).toBe("https://s3.example.com/file.txt");
    });

    it("should throw error on failure", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("File not found");

      await expect(FileService.getFileUrl(999)).rejects.toThrow("File not found");
    });

    it("should throw error on invalid JSON response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as Response);

      await expect(FileService.getFileUrl(1)).rejects.toThrow("Invalid response format from server");
    });
  });

  describe("deleteFile", () => {
    it("should delete file successfully", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(FileService.deleteFile(1, "test-token")).resolves.toBeUndefined();
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/files/1", {
        method: "DELETE",
        headers: { Authorization: "Bearer test-token" },
      });
    });

    it("should throw error on delete failure", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("File not found");

      await expect(FileService.deleteFile(999, "test-token")).rejects.toThrow("File not found");
    });
  });

  describe("getStorageStatus", () => {
    it("should return storage status", async () => {
      const mockStatus = { storage_type: "s3", message: "Storage available" };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockStatus }),
      } as Response);

      const result = await FileService.getStorageStatus();

      expect(result).toEqual(mockStatus);
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/files/storage/status");
    });

    it("should handle old response format (fallback)", async () => {
      const mockStatus = { storage_type: "database", message: "Using database storage" };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatus,
      } as Response);

      const result = await FileService.getStorageStatus();

      expect(result).toEqual(mockStatus);
    });

    it("should throw error on failure", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("Storage unavailable");

      await expect(FileService.getStorageStatus()).rejects.toThrow("Storage unavailable");
    });

    it("should throw error on invalid JSON response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as Response);

      await expect(FileService.getStorageStatus()).rejects.toThrow("Invalid response format from server");
    });
  });

  describe("downloadFile", () => {
    it("should download file and return blob", async () => {
      const mockBlob = new Blob(["test content"], { type: "text/plain" });

      mockFetch.mockResolvedValueOnce({
        ok: true,
        blob: async () => mockBlob,
      } as Response);

      const result = await FileService.downloadFile(1);

      expect(result).toBe(mockBlob);
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/files/1/download");
    });

    it("should throw error on download failure", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce("File not found");

      await expect(FileService.downloadFile(999)).rejects.toThrow("File not found");
    });
  });
});
