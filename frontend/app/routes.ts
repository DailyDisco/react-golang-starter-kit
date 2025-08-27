import { type RouteConfig, index } from "@react-router/dev/routes";

export default [
    index("routes/home.tsx"),
    {
        path: "/test-golang",
        file: "routes/testGolang.tsx",
    }
] satisfies RouteConfig;
