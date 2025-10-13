# Frontend Development Guide

Comprehensive guide for React frontend development with TanStack Router, Vite, and Docker.

## Table of Contents

- [Quick Start](#quick-start)
- [Technology Stack](#technology-stack)
- [TanStack Router Setup](#tanstack-router-setup)
- [File-Based Routing](#file-based-routing)
- [Docker Development](#docker-development)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- Node.js 20+ (LTS)
- Docker and Docker Compose (recommended)
- npm or yarn

### Local Development

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# App runs at http://localhost:5173
```

### Docker Development (Recommended)

```bash
# From project root
docker compose up frontend

# App runs at http://localhost:5173
```

---

## Technology Stack

### Core Libraries

- **[Vite](https://vitejs.dev/)** - Lightning-fast development server and build tool
- **[React 18](https://react.dev/)** - UI library with concurrent features
- **[TypeScript](https://www.typescriptlang.org/)** - Type-safe development
- **[TanStack Router](https://tanstack.com/router)** - Type-safe, file-based routing with data loading
- **[TanStack Query](https://tanstack.com/query)** - Powerful server state management with caching
- **[Zustand](https://zustand.pm/)** - Lightweight client state management

### UI & Styling

- **[TailwindCSS](https://tailwindcss.com/)** - Utility-first CSS framework
- **[ShadCN UI](https://ui.shadcn.com/)** - Beautiful, accessible component library
- **[Lucide Icons](https://lucide.dev/)** - Modern icon library

### Testing

- **[Vitest](https://vitest.dev/)** - Fast unit testing framework
- **[React Testing Library](https://testing-library.com/react)** - Component testing utilities
- **[Happy DOM](https://github.com/capricorn86/happy-dom)** - Lightweight DOM implementation

---

## TanStack Router Setup

### Overview

TanStack Router is configured to work automatically in both development and production. Route generation happens automatically via the Vite plugin.

### How It Works

The TanStack Router plugin is configured in [vite.config.ts](../frontend/vite.config.ts):

```typescript
tanstackRouter({
  target: "react",
  autoCodeSplitting: true,
  routesDirectory: "./app/routes",
  generatedRouteTree: "./app/routeTree.gen.ts",
  routeFileIgnorePrefix: "-",
  quoteStyle: "single",
});
```

**The plugin automatically generates `app/routeTree.gen.ts` whenever:**

- You start the dev server (`npm run dev`)
- You build for production (`npm run build`)
- You save changes to any route file in `app/routes/`

### Important Notes

- ❌ **Don't** manually edit `routeTree.gen.ts`
- ✅ **Do** commit `routeTree.gen.ts` to git (ensures reliable builds)
- ✅ **Do** let the Vite plugin handle generation automatically
- ❌ **Don't** try to use the TanStack Router CLI (`tsr`)

### Router Configuration

The router is set up in [app/router.tsx](../frontend/app/router.tsx) with:

- SSR Query integration for server-side rendering support
- Intent-based preloading (routes preload on hover)
- Default pending component for loading states
- Error boundary for graceful error handling

---

## File-Based Routing

### Route File Naming Conventions

| File                | Route            | Description                      |
| ------------------- | ---------------- | -------------------------------- |
| `__root.tsx`        | `/`              | Root layout (required)           |
| `index.tsx`         | `/`              | Home page                        |
| `blog.index.tsx`    | `/blog`          | Blog index                       |
| `blog.$postId.tsx`  | `/blog/:postId`  | Dynamic route param              |
| `users.$userId.tsx` | `/users/:userId` | Dynamic route param              |
| `(auth)/login.tsx`  | `/login`         | Route group (doesn't affect URL) |
| `_protected.tsx`    | -                | Layout route (no URL)            |
| `-component.tsx`    | -                | Ignored (utility file)           |

### Creating a New Route

1. **Create the route file:**

```bash
touch app/routes/pricing.tsx
```

2. **Add route component:**

```typescript
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/pricing")({
  component: PricingPage,
});

function PricingPage() {
  return (
    <div className='container mx-auto p-6'>
      <h1 className='text-3xl font-bold'>Pricing</h1>
      <p>Check out our pricing plans!</p>
    </div>
  );
}
```

3. **The route is automatically detected!** Visit `http://localhost:5173/pricing`

### Current Routes

Your app already includes:

- `/` - Home page
- `/about` - About page
- `/blog` - Blog page
- `/demo` - Demo page
- `/search` - Search page with query params
- `/login` - Login page
- `/register` - Register page
- `/profile` - Profile page (protected)
- `/dashboard` - Dashboard page (protected)
- `/users` - Users list
- `/users/:userId` - User detail page
- `/analytics` - Analytics page
- `/analytics/overview` - Analytics overview
- `/settings` - Settings page
- `/layout-demo` - Layout demo

### Protected Routes

Protected routes check authentication before rendering:

```typescript
import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/dashboard")({
  beforeLoad: async ({ context }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: DashboardPage,
});
```

### Data Loading

Routes can load data before rendering:

```typescript
export const Route = createFileRoute("/users/$userId")({
  loader: async ({ params }) => {
    const response = await fetch(`/api/users/${params.userId}`);
    return response.json();
  },
  component: UserDetailPage,
});

function UserDetailPage() {
  const user = Route.useLoaderData();
  return <div>Hello {user.name}!</div>;
}
```

---

## Docker Development

### Docker Configuration

The frontend uses a multi-stage Dockerfile with:

- **Development stage**: Full Node.js environment with Vite dev server
- **Production stage**: Nginx Alpine serving static files (~30MB)

### Volume Mounts

```yaml
frontend:
  volumes:
    - ./frontend:/app:cached # Source code
    - /app/node_modules # Exclude node_modules
    - vite_cache:/app/.vite:cached # Vite build cache
    # .tanstack/ is part of main mount (NOT separate volume)
```

### Why No Separate .tanstack Volume?

TanStack Router generates routes atomically:

1. Writes file to `.tanstack/tmp/...`
2. Renames it to `app/routeTree.gen.ts`

If `.tanstack` is on a **different volume**, you'll get this error:

```
Error: EXDEV: cross-device link not permitted, rename '/app/.tanstack/tmp/...' -> '/app/app/routeTree.gen.ts'
```

**Solution:** Keep `.tanstack` as part of the main bind mount (not a separate volume).

### Starting Development

```bash
# Start frontend only
docker compose up frontend

# Start with rebuild
docker compose up --build frontend

# View logs
docker compose logs -f frontend

# Restart container
docker compose restart frontend
```

### How Docker Startup Works

1. Container starts → Runs [start-dev.sh](../frontend/start-dev.sh)
2. Executes `npx vite --host 0.0.0.0 --port 5173`
3. Vite plugin automatically generates `app/routeTree.gen.ts`
4. Generated file appears in your local filesystem (via volume mount)
5. Changes to routes trigger automatic regeneration (hot reload works!)

### Production Build

```bash
# Build production image
docker compose -f docker-compose.prod.yml build frontend

# Start production
docker compose -f docker-compose.prod.yml up frontend
```

**What happens during production build:**

1. Dependencies installed (`npm ci`)
2. Source code copied
3. `npm run build` executes `vite build`
4. Vite plugin generates routes automatically
5. Static files served by Nginx

---

## Testing

### Test Configuration

The frontend uses **Vitest** with **Happy DOM** for fast, reliable testing.

### Running Tests

```bash
cd frontend

# Run tests once (CI mode)
npm run test:fast

# Run tests in watch mode (development)
npm test

# Run tests with coverage
npm run test:coverage

# Run tests with web UI (opens browser)
npm run test:ui
```

### Test Environment Features

- ✅ **Happy DOM** - Fast, lightweight DOM implementation
- ✅ **Global test functions** - No need to import describe/it/expect
- ✅ **Hot reload** - Tests rerun automatically on file changes
- ✅ **Coverage reporting** - Built-in coverage with HTML reports
- ✅ **Web UI** - Visual test runner with detailed results

### Writing Tests

```typescript
// components/Button.test.tsx
import { render, screen } from "@testing-library/react";
import { Button } from "./Button";

describe("Button", () => {
  it("renders with text", () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });

  it("calls onClick when clicked", () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);
    screen.getByText("Click").click();
    expect(handleClick).toHaveBeenCalledOnce();
  });
});
```

---

## Troubleshooting

### Route Generation Issues

#### Problem: `routeTree.gen.ts` not generated

**Solution 1: Check Vite Plugin**

```bash
npm list @tanstack/router-plugin
```

**Solution 2: Restart Dev Server**

```bash
# Local
npm run dev

# Docker
docker compose restart frontend
```

**Solution 3: Manual Trigger**

```bash
npm run build
```

#### Problem: Routes not updating

**Solution: Clear Cache**

```bash
# Local
rm -rf node_modules/.vite app/.tanstack

# Docker
docker compose down
docker volume rm react-golang-starter-kit_vite_cache
docker compose up frontend
```

### Docker Issues

#### Problem: "cross-device link" error

This happens when `.tanstack` is on a separate Docker volume.

**Solution:** Ensure your `docker-compose.yml` does NOT have a separate volume for `.tanstack`:

```yaml
# ❌ WRONG - causes error
volumes:
  - ./frontend:/app
  - tanstack_cache:/app/.tanstack  # Don't do this!

# ✅ CORRECT - .tanstack is part of main mount
volumes:
  - ./frontend:/app
  - /app/node_modules
```

#### Problem: Frontend build fails in Docker

**Solution: Clean rebuild**

```bash
docker compose down -v
docker compose build --no-cache frontend
docker compose up frontend
```

#### Problem: Changes not reflecting in Docker

**Solution 1: Check volume mounts**

```bash
docker compose config | grep -A 10 frontend
```

**Solution 2: Restart with rebuild**

```bash
docker compose up --build frontend
```

### TanStack Query Issues

#### Problem: "No QueryClient set" error

This happens when TanStack Query hooks are used outside `QueryClientProvider`.

**Solution:** Ensure `QueryClientProvider` wraps your app in [main.tsx](../frontend/app/main.tsx):

```typescript
<QueryClientProvider client={queryClient}>
  <RouterProvider router={router} />
</QueryClientProvider>
```

#### Problem: Mutations during SSR

**Solution:** Guard mutations for client-side only:

```typescript
const mutation = useMutation({
  mutationFn: apiCall,
});

// Only call on client-side
if (typeof window !== "undefined") {
  mutation.mutate(data);
}
```

### Build Issues

#### Problem: TypeScript errors

**Solution: Run type check**

```bash
npm run typecheck
```

#### Problem: Vite build fails

**Solution 1: Clear Vite cache**

```bash
rm -rf node_modules/.vite
npm run build
```

**Solution 2: Check for circular dependencies**

```bash
npm run build -- --debug
```

### Performance Issues

#### Problem: Slow dev server startup

**Solution: Use Docker volumes for caches**

```yaml
volumes:
  - vite_cache:/app/.vite:cached
```

#### Problem: Hot reload slow

**Solution: Reduce watched files in vite.config.ts**

```typescript
server: {
  watch: {
    ignored: ['**/node_modules/**', '**/.git/**'],
  },
}
```

---

## Best Practices

### File Organization

```
frontend/
├── app/
│   ├── routes/              # Route files
│   │   ├── __root.tsx      # Root layout
│   │   ├── index.tsx       # Home page
│   │   └── (auth)/         # Route groups
│   ├── components/         # React components
│   ├── lib/                # Utilities
│   ├── hooks/              # Custom hooks
│   ├── router.tsx          # Router setup
│   └── main.tsx            # App entry
├── public/                 # Static assets
└── vite.config.ts          # Vite configuration
```

### Routing Best Practices

1. ✅ **Commit `app/routeTree.gen.ts`** to version control
2. ✅ **Use file-based routing** - create files in `app/routes/`
3. ✅ **Follow naming conventions** for dynamic routes
4. ✅ **Use route groups `(name)`** to organize without affecting URLs
5. ✅ **Use layouts `_layout.tsx`** for shared UI
6. ❌ **Don't** manually edit `routeTree.gen.ts`
7. ❌ **Don't** ignore generated routes in `.gitignore`

### Component Best Practices

1. ✅ **Use ShadCN UI components** for consistent design
2. ✅ **Leverage TanStack Query** for server state
3. ✅ **Use Zustand** for client-only state
4. ✅ **Write tests** for critical components
5. ✅ **Follow accessibility guidelines** (ARIA labels, keyboard nav)

### Performance Best Practices

1. ✅ **Lazy load routes** (enabled by default via `autoCodeSplitting`)
2. ✅ **Use route-based code splitting**
3. ✅ **Optimize images** (use WebP, lazy loading)
4. ✅ **Minimize bundle size** (check with `npm run build -- --analyze`)
5. ✅ **Use React.memo** for expensive components

---

## Additional Resources

- [TanStack Router Docs](https://tanstack.com/router/latest)
- [TanStack Query Docs](https://tanstack.com/query/latest)
- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://react.dev/)
- [TailwindCSS Docs](https://tailwindcss.com/docs)
- [ShadCN UI Components](https://ui.shadcn.com/)
- [Vitest Documentation](https://vitest.dev/)

---

## Summary

Your frontend setup is **production-ready**:

1. ✅ Vite plugin configured correctly
2. ✅ Docker volumes set up properly
3. ✅ File-based routing with automatic generation
4. ✅ Hot reload works in development
5. ✅ Production builds work reliably
6. ✅ Testing infrastructure complete

**Just start your dev server and code!** Route generation is fully automated.
