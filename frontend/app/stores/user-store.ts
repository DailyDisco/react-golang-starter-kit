import { create } from "zustand";
import { devtools } from "zustand/middleware";

export interface UserFilters {
  search?: string;
  role?: string;
  isActive?: boolean;
}

interface UserState {
  // Client state (UI state)
  selectedUserId: number | null;
  filters: UserFilters;
  editMode: boolean;
  formData: { name: string; email: string; password: string };

  // Actions
  setSelectedUser: (id: number | null) => void;
  setFilters: (filters: Partial<UserFilters>) => void;
  setEditMode: (mode: boolean) => void;
  setFormData: (data: Partial<UserState["formData"]>) => void;
  resetForm: () => void;
}

export const useUserStore = create<UserState>()(
  devtools(
    (set, get) => ({
      selectedUserId: null,
      filters: { search: "", role: "", isActive: true },
      editMode: false,
      formData: { name: "", email: "", password: "" },

      setSelectedUser: (id) => set({ selectedUserId: id }),
      setFilters: (filters) =>
        set((state) => ({
          filters: { ...state.filters, ...filters },
        })),
      setEditMode: (mode) => set({ editMode: mode }),
      setFormData: (data) =>
        set((state) => ({
          formData: { ...state.formData, ...data },
        })),
      resetForm: () => set({ formData: { name: "", email: "", password: "" } }),
    }),
    { name: "user-store" }
  )
);
