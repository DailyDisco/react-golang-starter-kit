import { CACHE_TIMES } from "@/lib/cache-config";
import { queryKeys } from "@/lib/query-keys";
import { OrganizationService } from "@/services/organizations/organizationService";
import { useQuery } from "@tanstack/react-query";

/**
 * Fetch the list of organizations the current user belongs to
 *
 * @example
 * ```tsx
 * const { data: organizations, isLoading } = useOrganizations();
 *
 * return (
 *   <Select>
 *     {organizations?.map((org) => (
 *       <SelectItem key={org.slug} value={org.slug}>
 *         {org.name}
 *       </SelectItem>
 *     ))}
 *   </Select>
 * );
 * ```
 */
export function useOrganizations() {
  return useQuery({
    queryKey: queryKeys.organizations.list(),
    queryFn: () => OrganizationService.listOrganizations(),
    staleTime: CACHE_TIMES.USER_DATA,
  });
}

/**
 * Fetch a single organization by slug
 *
 * @example
 * ```tsx
 * const { data: org, isLoading } = useOrganization("acme-corp");
 * ```
 */
export function useOrganization(slug: string) {
  return useQuery({
    queryKey: queryKeys.organizations.detail(slug),
    queryFn: () => OrganizationService.getOrganization(slug),
    enabled: Boolean(slug),
    staleTime: CACHE_TIMES.USER_DATA,
  });
}
