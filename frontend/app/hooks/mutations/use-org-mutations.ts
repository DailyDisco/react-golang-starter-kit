import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { logger } from "../../lib/logger";
import { showMutationError, showMutationSuccess } from "../../lib/mutation-toast";
import { queryKeys } from "../../lib/query-keys";
import {
  OrganizationService,
  type OrganizationInvitation,
  type OrganizationMember,
  type UpdateOrganizationRequest,
} from "../../services/organizations/organizationService";
import { useOrgStore } from "../../stores/org-store";

// ============================================================================
// Organization Management Mutations
// ============================================================================

export function useUpdateOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const { updateOrganization } = useOrgStore();

  const mutation = useMutation({
    mutationFn: (data: UpdateOrganizationRequest) => OrganizationService.updateOrganization(orgSlug, data),

    onSuccess: (updatedOrg) => {
      updateOrganization(orgSlug, { name: updatedOrg.name });
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      showMutationSuccess({ message: "Organization updated" });
    },

    onError: (error: Error, variables) => {
      logger.error("Organization update error", error);
      showMutationError({
        error,
        onRetry: () => mutation.mutate(variables),
      });
    },
  });

  return mutation;
}

export function useDeleteOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { removeOrganization } = useOrgStore();

  const mutation = useMutation({
    mutationFn: () => OrganizationService.deleteOrganization(orgSlug),

    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      showMutationSuccess({ message: "Organization deleted" });
      navigate({ to: "/" });
    },

    onError: (error: Error) => {
      logger.error("Organization deletion error", error);
      showMutationError({
        error,
        onRetry: () => mutation.mutate(),
      });
    },
  });

  return mutation;
}

export function useLeaveOrganization(orgSlug: string) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { removeOrganization } = useOrgStore();

  const mutation = useMutation({
    mutationFn: () => OrganizationService.leaveOrganization(orgSlug),

    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      showMutationSuccess({ message: "Left organization" });
      navigate({ to: "/" });
    },

    onError: (error: Error) => {
      logger.error("Leave organization error", error);
      showMutationError({
        error,
        onRetry: () => mutation.mutate(),
      });
    },
  });

  return mutation;
}

// ============================================================================
// Member Management Mutations
// ============================================================================

export function useInviteMember(orgSlug: string) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: { email: string; role: "admin" | "member" }) => OrganizationService.inviteMember(orgSlug, data),

    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.invitations(orgSlug),
      });
      showMutationSuccess({ message: "Invitation sent" });
    },

    onError: (error: Error, variables) => {
      logger.error("Invite member error", error);
      showMutationError({
        error,
        onRetry: () => mutation.mutate(variables),
      });
    },
  });

  return mutation;
}

export function useUpdateMemberRole(orgSlug: string) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: "owner" | "admin" | "member" }) =>
      OrganizationService.updateMemberRole(orgSlug, userId, { role }),

    // Optimistic update
    onMutate: async ({ userId, role }) => {
      const queryKey = queryKeys.organizations.members(orgSlug);
      await queryClient.cancelQueries({ queryKey });

      const previousMembers = queryClient.getQueryData<OrganizationMember[]>(queryKey);

      queryClient.setQueryData<OrganizationMember[]>(queryKey, (old) =>
        old?.map((member) => (member.user_id === userId ? { ...member, role } : member))
      );

      return { previousMembers };
    },

    onSuccess: () => {
      showMutationSuccess({ message: "Member role updated" });
    },

    onError: (error: Error, variables, context) => {
      logger.error("Update member role error", error);

      if (context?.previousMembers) {
        queryClient.setQueryData(queryKeys.organizations.members(orgSlug), context.previousMembers);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(variables),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.members(orgSlug),
      });
    },
  });

  return mutation;
}

export function useRemoveMember(orgSlug: string) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (userId: number) => OrganizationService.removeMember(orgSlug, userId),

    // Optimistic delete
    onMutate: async (userId) => {
      const queryKey = queryKeys.organizations.members(orgSlug);
      await queryClient.cancelQueries({ queryKey });

      const previousMembers = queryClient.getQueryData<OrganizationMember[]>(queryKey);

      queryClient.setQueryData<OrganizationMember[]>(queryKey, (old) =>
        old?.filter((member) => member.user_id !== userId)
      );

      return { previousMembers };
    },

    onSuccess: () => {
      showMutationSuccess({ message: "Member removed" });
    },

    onError: (error: Error, userId, context) => {
      logger.error("Remove member error", error);

      if (context?.previousMembers) {
        queryClient.setQueryData(queryKeys.organizations.members(orgSlug), context.previousMembers);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(userId),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.members(orgSlug),
      });
    },
  });

  return mutation;
}

export function useCancelInvitation(orgSlug: string) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (invitationId: number) => OrganizationService.cancelInvitation(orgSlug, invitationId),

    // Optimistic delete
    onMutate: async (invitationId) => {
      const queryKey = queryKeys.organizations.invitations(orgSlug);
      await queryClient.cancelQueries({ queryKey });

      const previousInvitations = queryClient.getQueryData<OrganizationInvitation[]>(queryKey);

      queryClient.setQueryData<OrganizationInvitation[]>(queryKey, (old) =>
        old?.filter((inv) => inv.id !== invitationId)
      );

      return { previousInvitations };
    },

    onSuccess: () => {
      showMutationSuccess({ message: "Invitation cancelled" });
    },

    onError: (error: Error, invitationId, context) => {
      logger.error("Cancel invitation error", error);

      if (context?.previousInvitations) {
        queryClient.setQueryData(queryKeys.organizations.invitations(orgSlug), context.previousInvitations);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(invitationId),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.invitations(orgSlug),
      });
    },
  });

  return mutation;
}
