// Import CSS
import './app.css';
// Import router types first
import './router.types';

import { RouterProvider } from '@tanstack/react-router';
import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';

// Import the router with SSR Query integration
import { createAppRouter } from './router';

// Router types are registered in router.types.ts

// Create a new router instance with SSR Query integration
const router = createAppRouter();

// Render the app
const rootElement = document.getElementById('root')!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>
  );
}
