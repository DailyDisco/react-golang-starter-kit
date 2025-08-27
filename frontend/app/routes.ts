import { type RouteConfig, index } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  {
    path: "/demo",
    file: "routes/demo.tsx",
  },
] satisfies RouteConfig;
