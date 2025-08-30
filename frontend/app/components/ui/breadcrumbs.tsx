import { Link, useLocation } from '@tanstack/react-router';
import { Home } from 'lucide-react';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from './breadcrumb';
import { cn } from '@/lib/utils';

interface BreadcrumbItemType {
    label: string;
    href?: string;
}

interface BreadcrumbsProps {
    items?: BreadcrumbItemType[];
    className?: string;
}

export function Breadcrumbs({ items, className }: BreadcrumbsProps) {
    const location = useLocation();

    // Auto-generate breadcrumbs from current path if not provided
    const generateBreadcrumbs = (): BreadcrumbItemType[] => {
        if (items) return items;

        const pathSegments = location.pathname.split('/').filter(Boolean);
        const breadcrumbs: BreadcrumbItemType[] = [{ label: 'Home', href: '/' }];

        let currentPath = '';
        pathSegments.forEach((segment: string, index: number) => {
            currentPath += `/${segment}`;

            // Skip route parameters (segments starting with $)
            if (segment.startsWith('$')) return;

            // Convert route segments to readable labels
            let label = segment
                .replace(/^\(|\)$/g, '') // Remove parentheses from route groups
                .replace(/-/g, ' ') // Replace hyphens with spaces
                .replace(/\b\w/g, (l: string) => l.toUpperCase()); // Capitalize words

            // Handle special cases
            const specialLabels: Record<string, string> = {
                'users': 'Users',
                'analytics': 'Analytics',
                'settings': 'Settings',
                'profile': 'Profile',
                'login': 'Sign In',
                'register': 'Sign Up',
                'about': 'About',
                'blog': 'Blog',
                'search': 'Search',
                'demo': 'Demo',
                'layout-demo': 'Layout Demo',
            };

            if (specialLabels[segment.toLowerCase()]) {
                label = specialLabels[segment.toLowerCase()];
            }

            // Handle user IDs in routes (e.g., /users/123)
            if (segment.match(/^\d+$/) && index > 0 && pathSegments[index - 1] === 'users') {
                label = `User ${segment}`;
            }

            breadcrumbs.push({
                label,
                href: index === pathSegments.length - 1 ? undefined : currentPath
            });
        });

        return breadcrumbs;
    };

    const breadcrumbs = generateBreadcrumbs();

    if (breadcrumbs.length <= 1) return null;

    return (
        <Breadcrumb className={cn('', className)}>
            <BreadcrumbList>
                {breadcrumbs.map((item, index) => (
                    <div key={index} className="flex items-center">
                        {index > 0 && <BreadcrumbSeparator />}
                        <BreadcrumbItem>
                            {item.href ? (
                                <BreadcrumbLink asChild>
                                    <Link
                                        to={item.href}
                                        search={{}}
                                        className="flex items-center gap-1"
                                    >
                                        {index === 0 && <Home className="w-4 h-4" />}
                                        {item.label}
                                    </Link>
                                </BreadcrumbLink>
                            ) : (
                                <BreadcrumbPage className="flex items-center gap-1">
                                    {index === 0 && <Home className="w-4 h-4" />}
                                    {item.label}
                                </BreadcrumbPage>
                            )}
                        </BreadcrumbItem>
                    </div>
                ))}
            </BreadcrumbList>
        </Breadcrumb>
    );
}
