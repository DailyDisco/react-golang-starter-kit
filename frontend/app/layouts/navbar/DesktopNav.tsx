import { NavLink } from "./NavLink";
import { navigation } from "./types";

interface DesktopNavProps {
  pathname: string;
}

export function DesktopNav({ pathname }: DesktopNavProps) {
  return (
    <div className="hidden md:ml-6 md:flex md:items-center md:space-x-1">
      {navigation.map((item) => (
        <div
          key={item.name}
          className="flex items-center"
        >
          <NavLink
            item={item}
            pathname={pathname}
            variant="desktop"
          />
          {item.separator && <div className="ml-2 h-5 w-px bg-gray-300 dark:bg-gray-600" />}
        </div>
      ))}
    </div>
  );
}
