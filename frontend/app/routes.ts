import {
  type RouteConfig,
  index,
  layout,
  route,
} from "@react-router/dev/routes";

export default [
  // Regular routes with standard navbar/footer layout
  layout("./layouts/StandardLayout.tsx", [
    index("./routes/home.tsx"),
    route("demo", "./routes/demo.tsx"),
  ]),

  // Custom layout routes - completely separate layout system
  layout("./layouts/CustomDemoLayout.tsx", [
    route("layout-demo", "./routes/custom-layout-demo.tsx"),
  ]),

  // Catch-all route
  route("*", "./routes/404.tsx"),
] satisfies RouteConfig;
