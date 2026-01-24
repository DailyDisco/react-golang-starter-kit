import { lazy, Suspense } from "react";

import { createFileRoute } from "@tanstack/react-router";

const Demo = lazy(() => import("../../components/demo/demo").then((m) => ({ default: m.Demo })));

export const Route = createFileRoute("/(public)/demo")({
  component: DemoRoute,
});

function DemoRoute() {
  return (
    <Suspense fallback={<div className="flex h-screen items-center justify-center">Loading demo...</div>}>
      <Demo />
    </Suspense>
  );
}
