import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { useUserStore } from "./user-store";

describe("useUserStore", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    act(() => {
      useUserStore.setState({
        selectedUserId: null,
        filters: { search: "", role: "", isActive: true },
        editMode: false,
        formData: { name: "", email: "", password: "" },
      });
    });
  });

  describe("initial state", () => {
    it("has null selectedUserId initially", () => {
      const state = useUserStore.getState();
      expect(state.selectedUserId).toBeNull();
    });

    it("has default filters", () => {
      const state = useUserStore.getState();
      expect(state.filters).toEqual({
        search: "",
        role: "",
        isActive: true,
      });
    });

    it("has editMode set to false initially", () => {
      const state = useUserStore.getState();
      expect(state.editMode).toBe(false);
    });

    it("has empty formData initially", () => {
      const state = useUserStore.getState();
      expect(state.formData).toEqual({
        name: "",
        email: "",
        password: "",
      });
    });
  });

  describe("setSelectedUser", () => {
    it("sets the selected user id", () => {
      act(() => {
        useUserStore.getState().setSelectedUser(5);
      });

      expect(useUserStore.getState().selectedUserId).toBe(5);
    });

    it("can set selectedUserId to null", () => {
      act(() => {
        useUserStore.getState().setSelectedUser(10);
      });

      expect(useUserStore.getState().selectedUserId).toBe(10);

      act(() => {
        useUserStore.getState().setSelectedUser(null);
      });

      expect(useUserStore.getState().selectedUserId).toBeNull();
    });
  });

  describe("setFilters", () => {
    it("updates search filter", () => {
      act(() => {
        useUserStore.getState().setFilters({ search: "john" });
      });

      const filters = useUserStore.getState().filters;
      expect(filters.search).toBe("john");
      expect(filters.role).toBe("");
      expect(filters.isActive).toBe(true);
    });

    it("updates role filter", () => {
      act(() => {
        useUserStore.getState().setFilters({ role: "admin" });
      });

      const filters = useUserStore.getState().filters;
      expect(filters.role).toBe("admin");
    });

    it("updates isActive filter", () => {
      act(() => {
        useUserStore.getState().setFilters({ isActive: false });
      });

      const filters = useUserStore.getState().filters;
      expect(filters.isActive).toBe(false);
    });

    it("updates multiple filters at once", () => {
      act(() => {
        useUserStore.getState().setFilters({
          search: "test",
          role: "user",
          isActive: false,
        });
      });

      const filters = useUserStore.getState().filters;
      expect(filters).toEqual({
        search: "test",
        role: "user",
        isActive: false,
      });
    });

    it("preserves existing filters when updating partially", () => {
      act(() => {
        useUserStore.getState().setFilters({ search: "initial", role: "admin" });
      });

      act(() => {
        useUserStore.getState().setFilters({ search: "updated" });
      });

      const filters = useUserStore.getState().filters;
      expect(filters.search).toBe("updated");
      expect(filters.role).toBe("admin");
    });
  });

  describe("setEditMode", () => {
    it("sets editMode to true", () => {
      act(() => {
        useUserStore.getState().setEditMode(true);
      });

      expect(useUserStore.getState().editMode).toBe(true);
    });

    it("sets editMode to false", () => {
      act(() => {
        useUserStore.getState().setEditMode(true);
      });

      expect(useUserStore.getState().editMode).toBe(true);

      act(() => {
        useUserStore.getState().setEditMode(false);
      });

      expect(useUserStore.getState().editMode).toBe(false);
    });
  });

  describe("setFormData", () => {
    it("updates name in formData", () => {
      act(() => {
        useUserStore.getState().setFormData({ name: "John Doe" });
      });

      const formData = useUserStore.getState().formData;
      expect(formData.name).toBe("John Doe");
      expect(formData.email).toBe("");
      expect(formData.password).toBe("");
    });

    it("updates email in formData", () => {
      act(() => {
        useUserStore.getState().setFormData({ email: "john@example.com" });
      });

      const formData = useUserStore.getState().formData;
      expect(formData.email).toBe("john@example.com");
    });

    it("updates password in formData", () => {
      act(() => {
        useUserStore.getState().setFormData({ password: "secret123" });
      });

      const formData = useUserStore.getState().formData;
      expect(formData.password).toBe("secret123");
    });

    it("updates multiple fields at once", () => {
      act(() => {
        useUserStore.getState().setFormData({
          name: "Jane Doe",
          email: "jane@example.com",
          password: "password123",
        });
      });

      const formData = useUserStore.getState().formData;
      expect(formData).toEqual({
        name: "Jane Doe",
        email: "jane@example.com",
        password: "password123",
      });
    });

    it("preserves existing formData when updating partially", () => {
      act(() => {
        useUserStore.getState().setFormData({
          name: "Initial Name",
          email: "initial@example.com",
        });
      });

      act(() => {
        useUserStore.getState().setFormData({ name: "Updated Name" });
      });

      const formData = useUserStore.getState().formData;
      expect(formData.name).toBe("Updated Name");
      expect(formData.email).toBe("initial@example.com");
    });
  });

  describe("resetForm", () => {
    it("resets formData to empty values", () => {
      act(() => {
        useUserStore.getState().setFormData({
          name: "John Doe",
          email: "john@example.com",
          password: "secret123",
        });
      });

      expect(useUserStore.getState().formData).toEqual({
        name: "John Doe",
        email: "john@example.com",
        password: "secret123",
      });

      act(() => {
        useUserStore.getState().resetForm();
      });

      expect(useUserStore.getState().formData).toEqual({
        name: "",
        email: "",
        password: "",
      });
    });

    it("only resets formData, not other state", () => {
      act(() => {
        useUserStore.getState().setSelectedUser(5);
        useUserStore.getState().setEditMode(true);
        useUserStore.getState().setFilters({ search: "test" });
        useUserStore.getState().setFormData({ name: "Test" });
      });

      act(() => {
        useUserStore.getState().resetForm();
      });

      const state = useUserStore.getState();
      expect(state.selectedUserId).toBe(5);
      expect(state.editMode).toBe(true);
      expect(state.filters.search).toBe("test");
      expect(state.formData).toEqual({ name: "", email: "", password: "" });
    });
  });
});
