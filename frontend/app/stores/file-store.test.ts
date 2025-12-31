import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { useFileStore } from "./file-store";

describe("useFileStore", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    act(() => {
      useFileStore.getState().resetFileSelection();
    });
  });

  describe("initial state", () => {
    it("has null selectedFile initially", () => {
      const state = useFileStore.getState();
      expect(state.selectedFile).toBeNull();
    });

    it("has isDragOver set to false initially", () => {
      const state = useFileStore.getState();
      expect(state.isDragOver).toBe(false);
    });
  });

  describe("setSelectedFile", () => {
    it("sets the selected file", () => {
      const mockFile = new File(["test content"], "test.txt", { type: "text/plain" });

      act(() => {
        useFileStore.getState().setSelectedFile(mockFile);
      });

      const state = useFileStore.getState();
      expect(state.selectedFile).toBe(mockFile);
      expect(state.selectedFile?.name).toBe("test.txt");
    });

    it("can set selectedFile to null", () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

      act(() => {
        useFileStore.getState().setSelectedFile(mockFile);
      });

      expect(useFileStore.getState().selectedFile).toBe(mockFile);

      act(() => {
        useFileStore.getState().setSelectedFile(null);
      });

      expect(useFileStore.getState().selectedFile).toBeNull();
    });
  });

  describe("setIsDragOver", () => {
    it("sets isDragOver to true", () => {
      act(() => {
        useFileStore.getState().setIsDragOver(true);
      });

      expect(useFileStore.getState().isDragOver).toBe(true);
    });

    it("sets isDragOver to false", () => {
      act(() => {
        useFileStore.getState().setIsDragOver(true);
      });

      expect(useFileStore.getState().isDragOver).toBe(true);

      act(() => {
        useFileStore.getState().setIsDragOver(false);
      });

      expect(useFileStore.getState().isDragOver).toBe(false);
    });
  });

  describe("resetFileSelection", () => {
    it("resets selectedFile to null", () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

      act(() => {
        useFileStore.getState().setSelectedFile(mockFile);
      });

      expect(useFileStore.getState().selectedFile).toBe(mockFile);

      act(() => {
        useFileStore.getState().resetFileSelection();
      });

      expect(useFileStore.getState().selectedFile).toBeNull();
    });

    it("resets isDragOver to false", () => {
      act(() => {
        useFileStore.getState().setIsDragOver(true);
      });

      expect(useFileStore.getState().isDragOver).toBe(true);

      act(() => {
        useFileStore.getState().resetFileSelection();
      });

      expect(useFileStore.getState().isDragOver).toBe(false);
    });

    it("resets both selectedFile and isDragOver together", () => {
      const mockFile = new File(["test"], "test.txt", { type: "text/plain" });

      act(() => {
        useFileStore.getState().setSelectedFile(mockFile);
        useFileStore.getState().setIsDragOver(true);
      });

      expect(useFileStore.getState().selectedFile).toBe(mockFile);
      expect(useFileStore.getState().isDragOver).toBe(true);

      act(() => {
        useFileStore.getState().resetFileSelection();
      });

      const state = useFileStore.getState();
      expect(state.selectedFile).toBeNull();
      expect(state.isDragOver).toBe(false);
    });
  });
});
