import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { logger } from "../../lib/logger";
import { FileService, type FileResponse } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

export function useFileUpload() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async (file: File) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return FileService.uploadFile(file);
    },
    onSuccess: (uploadedFile: FileResponse) => {
      queryClient.invalidateQueries({ queryKey: ["files"] });
      toast.success("File uploaded successfully!", {
        description: `${uploadedFile.file_name} (${(uploadedFile.file_size / 1024).toFixed(2)} KB)`,
      });
    },
    onError: (error: Error) => {
      logger.error("File upload error", error);

      if (error.message.includes("size exceeds")) {
        toast.error("File too large", {
          description: "Please select a smaller file (max 10MB)",
        });
      } else if (error.message.includes("not found") || error.message.includes("Authentication")) {
        toast.error("Authentication required", {
          description: "Please log in to upload files",
        });
      } else {
        toast.error("Upload failed", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
  });
}

export function useFileDelete() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async (fileId: number) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return FileService.deleteFile(fileId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["files"] });
      toast.success("File deleted successfully!");
    },
    onError: (error: Error) => {
      logger.error("File deletion error", error);
      toast.error("Failed to delete file", {
        description: error.message || "An unexpected error occurred",
      });
    },
  });
}
