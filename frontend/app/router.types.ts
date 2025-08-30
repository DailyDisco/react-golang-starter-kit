import '@tanstack/react-router';

// Comprehensive router type overrides to bypass TanStack Router type issues
declare module '@tanstack/react-router' {
  interface Register {
    router: any;
  }

  // Override createFileRoute to accept any path
  export function createFileRoute<T extends string>(path: T): any;

  // Override navigation types
  export function useNavigate(): any;
  export function useLocation(): any;
  export function useParams(): any;
  export function Navigate(props: any): any;
}
