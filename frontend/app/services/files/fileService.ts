import { API_BASE_URL, authenticatedFetch, authenticatedFetchWithParsing, parseErrorResponse } from "../api/client";
import type { File, FileResponse, StorageStatus } from "../types";

export class FileService {
  /**
   * Upload a file
   */
  static async uploadFile(file: File, token: string): Promise<FileResponse> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${API_BASE_URL}/api/files/upload`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
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
    } catch (parseError) {
      console.error("Failed to parse upload response:", parseError);
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
          console.warn("Authentication required for files endpoint, returning empty array");
          return [];
        }
        const errorMessage = await parseErrorResponse(response, "Failed to fetch files");
        throw new Error(errorMessage);
      }

      const responseData = await response.json();
      console.log("Files API response:", responseData); // Debug log

      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data || [];
      }
      // Fallback for old format
      return responseData || [];
    } catch (parseError) {
      console.error("Failed to parse files response:", parseError);
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
    } catch (parseError) {
      console.error("Failed to parse file URL response:", parseError);
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Delete a file
   */
  static async deleteFile(fileId: number, token: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/files/${fileId}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
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
    } catch (parseError) {
      console.error("Failed to parse storage status response:", parseError);
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
