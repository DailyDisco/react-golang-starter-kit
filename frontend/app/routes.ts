import {
  type RouteConfig,
  index,
  layout,
  route,
} from '@react-router/dev/routes';

export default [
  // Authentication routes (no layout)
  route('login', './routes/login.tsx'),
  route('register', './routes/register.tsx'),

  // Regular routes with standard navbar/footer layout
  layout('./layouts/StandardLayout.tsx', [
    index('./routes/home.tsx'),
    route('demo', './routes/demo.tsx'),
    route('users/:userId', './routes/users.tsx'),
    route('profile', './routes/profile.tsx'),
  ]),

  // Custom layout routes - completely separate layout system
  layout('./layouts/CustomDemoLayout.tsx', [
    route('layout-demo', './routes/custom-layout-demo.tsx'),
  ]),

  // Catch-all route
  route('*', './routes/404.tsx'),
] satisfies RouteConfig;
