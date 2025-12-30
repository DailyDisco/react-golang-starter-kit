import { createFileRoute } from "@tanstack/react-router";

import { Demo } from "../../components/demo/demo";

export const Route = createFileRoute("/(public)/demo")({
  component: DemoRoute,
});

function DemoRoute() {
  return <Demo />;
}
