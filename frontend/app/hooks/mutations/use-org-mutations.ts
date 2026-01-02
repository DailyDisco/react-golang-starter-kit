import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import { queryKeys } from "../../lib/query-keys";
import { OrganizationService, type UpdateOrganizationRequest } from "../../services/organizations/organizationService";
import { useOrgStore } from "../../stores/org-store";

// ============================================================================
// Organization Management Mutations
// ============================================================================

export function useUpdateOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const { updateOrganization } = useOrgStore();

  return useMutation({
    mutationFn: (data: UpdateOrganizationRequest) => OrganizationService.updateOrganization(orgSlug, data),
    onSuccess: (updatedOrg) => {
      updateOrganization(orgSlug, { name: updatedOrg.name });
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useDeleteOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { removeOrganization } = useOrgStore();

  return useMutation({
    mutationFn: () => OrganizationService.deleteOrganization(orgSlug),
    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      navigate({ to: "/" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useLeaveOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { removeOrganization } = useOrgStore();

  return useMutation({
    mutationFn: () => OrganizationService.leaveOrganization(orgSlug),
    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      navigate({ to: "/" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

// ============================================================================
// Member Management Mutations
// ============================================================================

export function useInviteMember(orgSlug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { email: string; role: "admin" | "member" }) => OrganizationService.inviteMember(orgSlug, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.invitations(orgSlug),
      });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useUpdateMemberRole(orgSlug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: "owner" | "admin" | "member" }) =>
      OrganizationService.updateMemberRole(orgSlug, userId, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.members(orgSlug),
      });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useRemoveMember(orgSlug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: number) => OrganizationService.removeMember(orgSlug, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.members(orgSlug),
      });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useCancelInvitation(orgSlug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (invitationId: number) => OrganizationService.cancelInvitation(orgSlug, invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.invitations(orgSlug),
      });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}
