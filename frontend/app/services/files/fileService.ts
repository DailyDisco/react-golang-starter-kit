import { API_BASE_URL, authenticatedFetch, generateRequestId, parseErrorResponse } from "../api/client";
import type { FileResponse, StorageStatus } from "../types";

// Get CSRF token from cookie for file uploads
const getCSRFToken = (): string | null => {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp("(^| )csrf_token=([^;]+)"));
  return match ? decodeURIComponent(match[2]) : null;
};

export class FileService {
  /**
   * Upload a file
   * Authentication is handled via httpOnly cookies
   */
  static async uploadFile(file: File): Promise<FileResponse> {
    const formData = new FormData();
    formData.append("file", file);

    // Build headers without Content-Type (browser sets it with boundary for FormData)
    const headers: Record<string, string> = {
      "X-Request-ID": generateRequestId(),
    };

    // Add CSRF token for state-changing request
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }

    const response = await fetch(`${API_BASE_URL}/api/files/upload`, {
      method: "POST",
      credentials: "include", // Include httpOnly cookies for authentication
      headers,
      body: formData,
    });

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, "Failed to upload file");
      throw new Error(errorMessage);
    }

    try {
      const responseData = await response.json();
      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data;
      }
      // Fallback for old format
      return responseData;
    } catch {
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Fetch all files
   */
  static async fetchFiles(limit?: number, offset?: number): Promise<FileResponse[]> {
    const params = new URLSearchParams();
    if (limit !== undefined) params.append("limit", limit.toString());
    if (offset !== undefined) params.append("offset", offset.toString());

    const url = `${API_BASE_URL}/api/files${params.toString() ? "?" + params.toString() : ""}`;

    try {
      const response = await authenticatedFetch(url);
      if (!response.ok) {
        // If authentication fails, return empty array instead of throwing
        if (response.status === 401 || response.status === 403) {
          return [];
        }
        const errorMessage = await parseErrorResponse(response, "Failed to fetch files");
        throw new Error(errorMessage);
      }

      const responseData = await response.json();

      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data || [];
      }
      // Fallback for old format
      return responseData || [];
    } catch {
      // Return empty array on parse errors
      return [];
    }
  }

  /**
   * Get file URL for download
   */
  static async getFileUrl(fileId: number): Promise<string> {
    const response = await fetch(`${API_BASE_URL}/api/files/${fileId}/url`);
    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, "Failed to get file URL");
      throw new Error(errorMessage);
    }

    try {
      const responseData = await response.json();
      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data;
      }
      // Fallback for old format
      return responseData;
    } catch {
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Delete a file
   * Authentication is handled via httpOnly cookies
   */
  static async deleteFile(fileId: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/files/${fileId}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, "Failed to delete file");
      throw new Error(errorMessage);
    }
  }

  /**
   * Get storage status
   */
  static async getStorageStatus(): Promise<StorageStatus> {
    const response = await fetch(`${API_BASE_URL}/api/files/storage/status`);
    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, "Failed to get storage status");
      throw new Error(errorMessage);
    }

    try {
      const responseData = await response.json();
      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data;
      }
      // Fallback for old format
      return responseData;
    } catch {
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Download file directly (for database-stored files)
   */
  static async downloadFile(fileId: number): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/api/files/${fileId}/download`);
    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, "Failed to download file");
      throw new Error(errorMessage);
    }

    return response.blob();
  }
}
