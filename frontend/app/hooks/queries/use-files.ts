import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { CACHE_TIMES } from "../../lib/cache-config";
import { logger } from "../../lib/logger";
import { queryKeys } from "../../lib/query-keys";
import { FileService, type FileResponse, type StorageStatus } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

// Note: Error handling moved to components - use the `error` return value from these hooks

export function useFiles(limit?: number, offset?: number) {
  const { isAuthenticated } = useAuthStore();

  return useQuery<FileResponse[], Error>({
    queryKey: queryKeys.files.list(limit, offset),
    queryFn: () => FileService.fetchFiles(limit, offset),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.FILES,
    retry: 2,
  });
}

export function useFileUrl(fileId: number) {
  return useQuery<string, Error>({
    queryKey: queryKeys.files.url(fileId),
    queryFn: () => FileService.getFileUrl(fileId),
    enabled: !!fileId && fileId > 0,
    staleTime: CACHE_TIMES.FILE_URL,
    retry: 1,
  });
}

export function useStorageStatus() {
  return useQuery<StorageStatus, Error>({
    queryKey: queryKeys.files.storageStatus(),
    queryFn: () => FileService.getStorageStatus(),
    staleTime: CACHE_TIMES.STORAGE_STATUS,
    retry: 1,
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
