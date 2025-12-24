import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { logger } from "../../lib/logger";
import { FileService, type FileResponse, type StorageStatus } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

export function useFiles(limit?: number, offset?: number) {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: ["files", limit, offset],
    queryFn: () => FileService.fetchFiles(limit, offset),
    enabled: isAuthenticated, // Only run query if authenticated
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
    onError: (error: Error) => {
      logger.error("Files fetch error", error);
      toast.error("Failed to load files", {
        description: "Please try again later",
      });
    },
  });
}

export function useFileUrl(fileId: number) {
  return useQuery({
    queryKey: ["file-url", fileId],
    queryFn: () => FileService.getFileUrl(fileId),
    enabled: !!fileId && fileId > 0,
    staleTime: 10 * 60 * 1000, // 10 minutes
    retry: 1,
    onError: (error: Error) => {
      logger.error("File URL fetch error", error);
    },
  });
}

export function useStorageStatus() {
  return useQuery({
    queryKey: ["storage-status"],
    queryFn: () => FileService.getStorageStatus(),
    staleTime: 30 * 60 * 1000, // 30 minutes
    retry: 1,
    onError: (error: Error) => {
      logger.error("Storage status fetch error", error);
    },
  });
}

// Hook for downloading files
export function useFileDownload() {
  return {
    downloadFile: async (fileId: number, fileName: string) => {
      try {
        const blob = await FileService.downloadFile(fileId);

        // Create download link
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);

        toast.success("File downloaded successfully!");
      } catch (error) {
        logger.error("File download error", error);
        toast.error("Download failed", {
          description: "Please try again later",
        });
      }
    },
  };
}
