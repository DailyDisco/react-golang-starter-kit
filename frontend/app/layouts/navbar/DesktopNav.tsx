import { NavLink } from "./NavLink";
import { navigation, type NavItem } from "./types";

interface DesktopNavProps {
  pathname: string;
}

export function DesktopNav({ pathname }: DesktopNavProps) {
  return (
    <div className="hidden md:ml-6 md:flex md:space-x-1">
      {navigation.map((item) => (
        <NavLink
          key={item.name}
          item={item}
          pathname={pathname}
          variant="desktop"
        />
      ))}
    </div>
  );
}
