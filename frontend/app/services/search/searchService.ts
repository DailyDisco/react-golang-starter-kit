/**
 * Global search service for searching across different entities
 */

export interface SearchResult {
  id: string;
  type: "page" | "user" | "setting" | "action" | "help";
  title: string;
  description?: string;
  url?: string;
  icon?: string;
  keywords?: string[];
  metadata?: Record<string, unknown>;
}

export interface SearchOptions {
  /** Types to search */
  types?: SearchResult["type"][];
  /** Maximum results per type */
  limit?: number;
  /** Include metadata in results */
  includeMetadata?: boolean;
}

/**
 * Static pages that can be navigated to
 */
const staticPages: SearchResult[] = [
  // Main
  {
    id: "dashboard",
    type: "page",
    title: "Dashboard",
    description: "Your main dashboard",
    url: "/dashboard",
    keywords: ["home", "main"],
  },
  {
    id: "billing",
    type: "page",
    title: "Billing",
    description: "Manage your subscription",
    url: "/billing",
    keywords: ["payment", "subscription", "plan", "pricing"],
  },

  // Settings
  {
    id: "profile",
    type: "setting",
    title: "Profile Settings",
    description: "Edit your profile information",
    url: "/settings/profile",
    keywords: ["name", "email", "avatar"],
  },
  {
    id: "security",
    type: "setting",
    title: "Security Settings",
    description: "Password and 2FA",
    url: "/settings/security",
    keywords: ["password", "2fa", "authentication", "mfa"],
  },
  {
    id: "preferences",
    type: "setting",
    title: "Preferences",
    description: "Theme and language",
    url: "/settings/preferences",
    keywords: ["theme", "dark mode", "language"],
  },
  {
    id: "notifications",
    type: "setting",
    title: "Notification Settings",
    description: "Email and push notifications",
    url: "/settings/notifications",
    keywords: ["email", "alerts", "push"],
  },
  {
    id: "privacy",
    type: "setting",
    title: "Privacy Settings",
    description: "Data and privacy controls",
    url: "/settings/privacy",
    keywords: ["data", "gdpr", "export"],
  },
  {
    id: "login-history",
    type: "setting",
    title: "Login History",
    description: "View your login sessions",
    url: "/settings/login-history",
    keywords: ["sessions", "activity", "devices"],
  },
  {
    id: "connected-accounts",
    type: "setting",
    title: "Connected Accounts",
    description: "Social login connections",
    url: "/settings/connected-accounts",
    keywords: ["oauth", "google", "github", "social"],
  },

  // Admin
  {
    id: "admin",
    type: "page",
    title: "Admin Dashboard",
    description: "Administration overview",
    url: "/admin",
    keywords: ["admin", "management"],
  },
  {
    id: "admin-users",
    type: "page",
    title: "User Management",
    description: "Manage all users",
    url: "/admin/users",
    keywords: ["admin", "users", "accounts"],
  },
  {
    id: "admin-audit",
    type: "page",
    title: "Audit Logs",
    description: "View system audit logs",
    url: "/admin/audit-logs",
    keywords: ["admin", "logs", "history", "events"],
  },
  {
    id: "admin-flags",
    type: "page",
    title: "Feature Flags",
    description: "Toggle feature flags",
    url: "/admin/feature-flags",
    keywords: ["admin", "features", "toggles", "flags"],
  },
  {
    id: "admin-health",
    type: "page",
    title: "System Health",
    description: "Monitor system status",
    url: "/admin/health",
    keywords: ["admin", "status", "monitoring", "health"],
  },
  {
    id: "admin-announcements",
    type: "page",
    title: "Announcements",
    description: "System announcements",
    url: "/admin/announcements",
    keywords: ["admin", "banners", "messages"],
  },
  {
    id: "admin-email",
    type: "page",
    title: "Email Templates",
    description: "Manage email templates",
    url: "/admin/email-templates",
    keywords: ["admin", "email", "templates"],
  },
  {
    id: "admin-settings",
    type: "page",
    title: "Admin Settings",
    description: "System configuration",
    url: "/admin/settings",
    keywords: ["admin", "config", "system"],
  },

  // Public pages
  {
    id: "home",
    type: "page",
    title: "Home Page",
    description: "Landing page",
    url: "/",
    keywords: ["home", "landing"],
  },
  {
    id: "pricing",
    type: "page",
    title: "Pricing",
    description: "View pricing plans",
    url: "/pricing",
    keywords: ["plans", "cost", "subscription"],
  },
  { id: "demo", type: "page", title: "Demo", description: "Interactive demo", url: "/demo", keywords: ["try", "test"] },
];

/**
 * Help articles and documentation
 */
const helpArticles: SearchResult[] = [
  {
    id: "help-start",
    type: "help",
    title: "Getting Started",
    description: "Learn the basics",
    keywords: ["start", "begin", "tutorial", "guide"],
  },
  {
    id: "help-keyboard",
    type: "help",
    title: "Keyboard Shortcuts",
    description: "Available shortcuts",
    keywords: ["keys", "shortcuts", "hotkeys", "cmd"],
  },
  {
    id: "help-security",
    type: "help",
    title: "Security Best Practices",
    description: "Keep your account secure",
    keywords: ["secure", "password", "2fa"],
  },
  {
    id: "help-billing",
    type: "help",
    title: "Billing FAQ",
    description: "Common billing questions",
    keywords: ["payment", "invoice", "refund"],
  },
];

/**
 * Search all static content
 */
function searchStatic(query: string, types: SearchResult["type"][], limit: number): SearchResult[] {
  const normalizedQuery = query.toLowerCase().trim();
  if (!normalizedQuery) return [];

  const allItems = [...staticPages, ...helpArticles];

  const results = allItems
    .filter((item) => {
      // Filter by type
      if (!types.includes(item.type)) return false;

      // Search in title, description, and keywords
      const searchText = [item.title, item.description, ...(item.keywords || [])].join(" ").toLowerCase();

      return searchText.includes(normalizedQuery);
    })
    .map((item) => ({
      ...item,
      // Calculate relevance score
      score: calculateRelevance(item, normalizedQuery),
    }))
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
    .map(({ score, ...item }) => item); // Remove score from result

  return results;
}

/**
 * Calculate relevance score for a search result
 */
function calculateRelevance(item: SearchResult, query: string): number {
  let score = 0;

  // Exact title match
  if (item.title.toLowerCase() === query) score += 100;

  // Title starts with query
  if (item.title.toLowerCase().startsWith(query)) score += 50;

  // Title contains query
  if (item.title.toLowerCase().includes(query)) score += 25;

  // Description contains query
  if (item.description?.toLowerCase().includes(query)) score += 10;

  // Keywords contain exact match
  if (item.keywords?.some((k) => k.toLowerCase() === query)) score += 30;

  // Keywords contain partial match
  if (item.keywords?.some((k) => k.toLowerCase().includes(query))) score += 15;

  return score;
}

/**
 * Global search function
 */
export async function globalSearch(query: string, options: SearchOptions = {}): Promise<SearchResult[]> {
  const { types = ["page", "setting", "action", "help"], limit = 10 } = options;

  // For now, just search static content
  // In a real app, this would also search users via API
  const staticResults = searchStatic(query, types, limit);

  return staticResults;
}

/**
 * Get recent searches from localStorage
 */
export function getRecentSearches(): string[] {
  if (typeof window === "undefined") return [];
  const stored = localStorage.getItem("recent-searches");
  return stored ? JSON.parse(stored) : [];
}

/**
 * Add a search to recent searches
 */
export function addRecentSearch(query: string): void {
  if (typeof window === "undefined" || !query.trim()) return;

  const recent = getRecentSearches();
  const updated = [query, ...recent.filter((q) => q !== query)].slice(0, 5);
  localStorage.setItem("recent-searches", JSON.stringify(updated));
}

/**
 * Clear recent searches
 */
export function clearRecentSearches(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem("recent-searches");
}
