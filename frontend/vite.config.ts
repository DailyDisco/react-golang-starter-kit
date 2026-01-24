import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import tsconfigPaths from 'vite-tsconfig-paths';
import { visualizer } from 'rollup-plugin-visualizer';
import path from 'path';

export default defineConfig({
  plugins: [
    // Please make sure that '@tanstack/router-plugin' is passed before '@vitejs/plugin-
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
      routesDirectory: './app/routes',
      generatedRouteTree: './app/routeTree.gen.ts',
      routeFileIgnorePrefix: '-',
      quoteStyle: 'single',
    }),
    react(),
    tailwindcss(),
    tsconfigPaths(),
    visualizer({
      filename: 'dist/stats.html',
      open: false,
      gzipSize: true,
      brotliSize: true,
    }),
  ],
  server: {
    host: '0.0.0.0',
    port: Number(process.env.FRONTEND_PORT) || 5193,
    allowedHosts: true, // Allow all hosts for remote development
    hmr: {
      port: Number(process.env.FRONTEND_HMR_PORT) || 4193,
    },
    proxy: {
      '/api': {
        target: 'http://backend:8080',
        changeOrigin: true,
        // Don't rewrite - backend expects /api prefix
      },
      '/swagger': {
        target: 'http://backend:8080',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './app'),
    },
  },
  test: {
    globals: true,
    environment: 'happy-dom', // Faster than jsdom
    setupFiles: ['./app/test/setup.tsx', './app/test/vitest.setup.ts'],
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/e2e/**', // Exclude Playwright E2E tests
      '**/*.e2e.ts',
      '**/*.e2e.tsx',
    ],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'app/test/',
        'app/routeTree.gen.ts',
        '**/*.d.ts',
        '**/*.config.*',
        '**/types/**',
        '**/e2e/**',
      ],
      thresholds: {
        // Coverage thresholds - realistic for current codebase
        lines: 70,
        functions: 68,
        branches: 63,
        statements: 70,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Core React
          vendor: ['react', 'react-dom'],

          // Routing & Data Fetching
          router: ['@tanstack/react-router'],
          query: ['@tanstack/react-query'],

          // Radix UI primitives (heavy)
          'radix-core': [
            '@radix-ui/react-dialog',
            '@radix-ui/react-dropdown-menu',
            '@radix-ui/react-select',
            '@radix-ui/react-popover',
            '@radix-ui/react-tooltip',
            '@radix-ui/react-tabs',
            '@radix-ui/react-switch',
            '@radix-ui/react-checkbox',
            '@radix-ui/react-label',
            '@radix-ui/react-slot',
          ],
          'radix-extended': [
            '@radix-ui/react-alert-dialog',
            '@radix-ui/react-avatar',
            '@radix-ui/react-progress',
            '@radix-ui/react-scroll-area',
            '@radix-ui/react-separator',
            '@radix-ui/react-slider',
          ],

          // Charts (heavy - ~150KB)
          charts: ['recharts'],

          // Animation (heavy - ~100KB)
          animation: ['framer-motion'],

          // Forms
          forms: ['react-hook-form', '@hookform/resolvers', 'zod'],

          // Utilities
          utils: ['clsx', 'tailwind-merge', 'date-fns', 'lucide-react'],

          // i18n
          i18n: ['i18next', 'react-i18next', 'i18next-browser-languagedetector'],

          // State management
          state: ['zustand'],
        },
      },
    },
    sourcemap: true,
    reportCompressedSize: true,
  },
});
