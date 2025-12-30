import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { OrganizationService } from "@/services/organizations/organizationService";
import { useCurrentOrg, useOrganizations, useOrgStore } from "@/stores/org-store";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Building2, Check, ChevronsUpDown, Plus, Settings } from "lucide-react";
import { useTranslation } from "react-i18next";

interface OrgSwitcherProps {
  className?: string;
  showCreateButton?: boolean;
}

export function OrgSwitcher({ className, showCreateButton = true }: OrgSwitcherProps) {
  const { t } = useTranslation("common");
  const navigate = useNavigate();
  const currentOrg = useCurrentOrg();
  const organizations = useOrganizations();
  const { setCurrentOrg, setOrganizations, setLoading } = useOrgStore();

  // Fetch organizations
  const { data: orgs, isLoading } = useQuery({
    queryKey: ["organizations"],
    queryFn: () => OrganizationService.listOrganizations(),
  });

  // Sync organizations to store
  useEffect(() => {
    if (orgs) {
      setOrganizations(orgs);
      setLoading(false);
    }
  }, [orgs, setOrganizations, setLoading]);

  // Set loading state
  useEffect(() => {
    setLoading(isLoading);
  }, [isLoading, setLoading]);

  const handleSelectOrg = (org: (typeof organizations)[0]) => {
    setCurrentOrg(org);
    // Optionally navigate to org dashboard
    // navigate({ to: `/org/${org.slug}` });
  };

  const handleCreateOrg = () => {
    navigate({ to: "/settings" }); // Navigate to settings where org creation would be
  };

  const handleOrgSettings = () => {
    if (currentOrg) {
      navigate({ to: `/org/${currentOrg.slug}/settings` });
    }
  };

  if (organizations.length === 0 && !isLoading) {
    return (
      <Button
        variant="outline"
        size="sm"
        className={cn("gap-2", className)}
        onClick={handleCreateOrg}
      >
        <Plus className="h-4 w-4" />
        {t("organization.create")}
      </Button>
    );
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-label={t("organization.select")}
          className={cn("justify-between gap-2", className)}
          disabled={isLoading}
        >
          <div className="flex items-center gap-2 truncate">
            <Building2 className="h-4 w-4 shrink-0" />
            <span className="truncate">{currentOrg?.name || t("organization.select")}</span>
          </div>
          <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[250px] p-0"
        align="start"
      >
        <Command>
          <CommandInput placeholder={t("organization.search")} />
          <CommandList>
            <CommandEmpty>{t("organization.notFound")}</CommandEmpty>
            <CommandGroup heading={t("organization.organizations")}>
              {organizations.map((org) => (
                <CommandItem
                  key={org.slug}
                  onSelect={() => handleSelectOrg(org)}
                  className="cursor-pointer"
                >
                  <Building2 className="mr-2 h-4 w-4" />
                  <div className="flex flex-1 flex-col">
                    <span className="truncate">{org.name}</span>
                    <span className="text-muted-foreground text-xs capitalize">{org.role}</span>
                  </div>
                  {currentOrg?.slug === org.slug && <Check className="ml-auto h-4 w-4" />}
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator />
            <CommandGroup>
              {currentOrg && (
                <CommandItem
                  onSelect={handleOrgSettings}
                  className="cursor-pointer"
                >
                  <Settings className="mr-2 h-4 w-4" />
                  {t("organization.settings")}
                </CommandItem>
              )}
              {showCreateButton && (
                <CommandItem
                  onSelect={handleCreateOrg}
                  className="cursor-pointer"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  {t("organization.create")}
                </CommandItem>
              )}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

// Compact version for sidebar
export function OrgSwitcherCompact({ className }: { className?: string }) {
  return (
    <OrgSwitcher
      className={cn("w-full", className)}
      showCreateButton={false}
    />
  );
}
