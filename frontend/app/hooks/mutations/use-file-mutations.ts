import { useMutation, useQueryClient } from "@tanstack/react-query";

import { logger } from "../../lib/logger";
import { showMutationError, showMutationSuccess } from "../../lib/mutation-toast";
import { queryKeys } from "../../lib/query-keys";
import { FileService, type FileResponse } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

export function useFileUpload() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  const mutation = useMutation({
    mutationFn: async (file: File) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return FileService.uploadFile(file);
    },

    onSuccess: (uploadedFile: FileResponse) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
      showMutationSuccess({
        message: "File uploaded successfully!",
        description: `${uploadedFile.file_name} (${(uploadedFile.file_size / 1024).toFixed(2)} KB)`,
      });
    },

    onError: (error: Error, file) => {
      logger.error("File upload error", error);
      showMutationError({
        error,
        onRetry: () => mutation.mutate(file),
      });
    },
  });

  return mutation;
}

export function useFileDelete() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  const mutation = useMutation({
    mutationFn: async (fileId: number) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return FileService.deleteFile(fileId);
    },

    // Optimistic delete
    onMutate: async (fileId) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: queryKeys.files.all });

      // Snapshot previous value (we need to match the query key pattern)
      const previousFiles = queryClient.getQueryData<FileResponse[]>(queryKeys.files.list());

      // Optimistically remove from cache
      queryClient.setQueryData<FileResponse[]>(queryKeys.files.list(), (old) =>
        old?.filter((file) => file.id !== fileId)
      );

      return { previousFiles, deletedFileId: fileId };
    },

    onSuccess: () => {
      showMutationSuccess({ message: "File deleted successfully!" });
    },

    onError: (error: Error, fileId, context) => {
      logger.error("File deletion error", error);

      // Rollback optimistic update
      if (context?.previousFiles) {
        queryClient.setQueryData(queryKeys.files.list(), context.previousFiles);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(fileId),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.all });
    },
  });

  return mutation;
}
