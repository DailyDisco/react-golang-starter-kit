import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface FileState {
  // Client state (UI state)
  selectedFile: File | null;
  isDragOver: boolean;

  // Actions
  setSelectedFile: (file: File | null) => void;
  setIsDragOver: (isDragOver: boolean) => void;
  resetFileSelection: () => void;
}

export const useFileStore = create<FileState>()(
  devtools(
    set => ({
      selectedFile: null,
      isDragOver: false,

      setSelectedFile: file => set({ selectedFile: file }),
      setIsDragOver: isDragOver => set({ isDragOver }),
      resetFileSelection: () => set({ selectedFile: null, isDragOver: false }),
    }),
    { name: 'file-store' }
  )
);
