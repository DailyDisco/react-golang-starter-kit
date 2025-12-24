// Import CSS
import "./app.css";

import { StrictMode } from "react";

import { RouterProvider } from "@tanstack/react-router";
import ReactDOM from "react-dom/client";

// Import Sentry and initialize as early as possible
import { initSentry } from "./lib/sentry";
// Import the router with SSR Query integration
import { createAppRouter } from "./router";
// Import auth store for initialization
import { useAuthStore } from "./stores/auth-store";

// Initialize Sentry before rendering
// This runs asynchronously but we don't wait for it
initSentry();

// Initialize auth store synchronously before router creation
// This loads cached user/token from localStorage for immediate availability
if (typeof window !== "undefined") {
  useAuthStore.getState().initialize();
}

// Router types are registered in router.types.ts

// Create a new router instance with SSR Query integration
const router = createAppRouter();

// Render the app
const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>
  );
}
