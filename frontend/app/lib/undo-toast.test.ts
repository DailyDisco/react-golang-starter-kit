import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { showUndoableBulkDelete, showUndoableDelete } from "./undo-toast";

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: Object.assign(
    vi.fn(() => "toast-id-123"),
    {
      success: vi.fn(),
      error: vi.fn(),
    }
  ),
}));

// Mock QueryClient
const createMockQueryClient = () => ({
  getQueryData: vi.fn(),
  setQueryData: vi.fn(),
  invalidateQueries: vi.fn(),
});

describe("showUndoableDelete", () => {
  const mockUser = { id: 1, name: "John Doe", email: "john@example.com" };
  const queryKey = ["users", "list"];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("optimistically removes item from cache", () => {
    const queryClient = createMockQueryClient();
    queryClient.getQueryData.mockReturnValue([mockUser, { id: 2, name: "Jane" }]);

    showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    // Should snapshot previous data
    expect(queryClient.getQueryData).toHaveBeenCalledWith(queryKey);

    // Should set data with filter function
    expect(queryClient.setQueryData).toHaveBeenCalledWith(queryKey, expect.any(Function));

    // Verify the filter function removes the item
    const setDataCall = queryClient.setQueryData.mock.calls[0];
    const filterFn = setDataCall[1];
    const result = filterFn([mockUser, { id: 2, name: "Jane" }]);
    expect(result).toEqual([{ id: 2, name: "Jane" }]);
  });

  it("shows toast with undo action", () => {
    const queryClient = createMockQueryClient();

    showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    expect(toast).toHaveBeenCalledWith(
      "John Doe deleted",
      expect.objectContaining({
        duration: 5000,
        action: expect.objectContaining({
          label: "Undo",
        }),
      })
    );
  });

  it("restores item when undo is called", () => {
    const queryClient = createMockQueryClient();
    const previousData = [mockUser, { id: 2, name: "Jane" }];
    queryClient.getQueryData.mockReturnValue(previousData);
    const onUndo = vi.fn();

    const result = showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm: vi.fn().mockResolvedValue(undefined),
      onUndo,
    });

    // Call the cancel function
    result.cancel();

    // Should restore previous data
    expect(queryClient.setQueryData).toHaveBeenLastCalledWith(queryKey, previousData);

    // Should call onUndo callback
    expect(onUndo).toHaveBeenCalled();

    // Should show success toast
    expect(toast.success).toHaveBeenCalledWith("John Doe restored");
  });

  it("calls onConfirm when toast auto-closes", async () => {
    const queryClient = createMockQueryClient();
    const onConfirm = vi.fn().mockResolvedValue(undefined);

    showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm,
    });

    // Get the onAutoClose callback from toast call
    const toastCall = (toast as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
    const options = toastCall[1];

    // Simulate auto-close
    await options.onAutoClose();

    expect(onConfirm).toHaveBeenCalled();
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({ queryKey });
  });

  it("rolls back on confirm error", async () => {
    const queryClient = createMockQueryClient();
    const previousData = [mockUser];
    queryClient.getQueryData.mockReturnValue(previousData);
    const onConfirm = vi.fn().mockRejectedValue(new Error("Delete failed"));

    showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm,
    });

    // Get the onAutoClose callback
    const toastCall = (toast as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
    const options = toastCall[1];

    // Simulate auto-close
    await options.onAutoClose();

    // Should rollback
    expect(queryClient.setQueryData).toHaveBeenLastCalledWith(queryKey, previousData);

    // Should show error toast
    expect(toast.error).toHaveBeenCalledWith(
      "Failed to delete John Doe",
      expect.objectContaining({
        description: "Delete failed",
      })
    );
  });

  it("uses custom timeout", () => {
    const queryClient = createMockQueryClient();

    showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm: vi.fn().mockResolvedValue(undefined),
      timeout: 8000,
    });

    expect(toast).toHaveBeenCalledWith(
      "John Doe deleted",
      expect.objectContaining({
        duration: 8000,
      })
    );
  });

  it("returns toast ID and cancel function", () => {
    const queryClient = createMockQueryClient();

    const result = showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    expect(result.toastId).toBe("toast-id-123");
    expect(typeof result.cancel).toBe("function");
  });

  it("does not call onConfirm if already undone", async () => {
    const queryClient = createMockQueryClient();
    const onConfirm = vi.fn().mockResolvedValue(undefined);

    const result = showUndoableDelete({
      queryClient: queryClient as any,
      queryKey,
      item: mockUser,
      itemLabel: "John Doe",
      onConfirm,
    });

    // Undo first
    result.cancel();

    // Then simulate auto-close
    const toastCall = (toast as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
    const options = toastCall[1];
    await options.onAutoClose();

    // onConfirm should not be called because we already undid
    expect(onConfirm).not.toHaveBeenCalled();
  });
});

describe("showUndoableBulkDelete", () => {
  const mockUsers = [
    { id: 1, name: "John" },
    { id: 2, name: "Jane" },
  ];
  const queryKey = ["users", "list"];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("optimistically removes all items", () => {
    const queryClient = createMockQueryClient();
    queryClient.getQueryData.mockReturnValue([...mockUsers, { id: 3, name: "Bob" }]);

    showUndoableBulkDelete({
      queryClient: queryClient as any,
      queryKey,
      items: mockUsers,
      itemsLabel: "2 users",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    // Verify the filter function removes both items
    const setDataCall = queryClient.setQueryData.mock.calls[0];
    const filterFn = setDataCall[1];
    const result = filterFn([...mockUsers, { id: 3, name: "Bob" }]);
    expect(result).toEqual([{ id: 3, name: "Bob" }]);
  });

  it("shows toast with plural label", () => {
    const queryClient = createMockQueryClient();

    showUndoableBulkDelete({
      queryClient: queryClient as any,
      queryKey,
      items: mockUsers,
      itemsLabel: "2 users",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    expect(toast).toHaveBeenCalledWith(
      "2 users deleted",
      expect.objectContaining({
        action: expect.objectContaining({ label: "Undo" }),
      })
    );
  });

  it("restores all items when undone", () => {
    const queryClient = createMockQueryClient();
    const previousData = [...mockUsers, { id: 3, name: "Bob" }];
    queryClient.getQueryData.mockReturnValue(previousData);

    const result = showUndoableBulkDelete({
      queryClient: queryClient as any,
      queryKey,
      items: mockUsers,
      itemsLabel: "2 users",
      onConfirm: vi.fn().mockResolvedValue(undefined),
    });

    result.cancel();

    expect(queryClient.setQueryData).toHaveBeenLastCalledWith(queryKey, previousData);
    expect(toast.success).toHaveBeenCalledWith("2 users restored");
  });
});
