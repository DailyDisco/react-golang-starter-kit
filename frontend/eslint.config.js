import js from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import reactPlugin from 'eslint-plugin-react';
import reactHooksPlugin from 'eslint-plugin-react-hooks';
import jsxA11yPlugin from 'eslint-plugin-jsx-a11y';
import tanstackQueryPlugin from '@tanstack/eslint-plugin-query';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';
// Removed: simple-import-sort conflicts with Prettier's @ianvs/prettier-plugin-sort-imports
import importPlugin from 'eslint-plugin-import';
import unicornPlugin from 'eslint-plugin-unicorn';
import securityPlugin from 'eslint-plugin-security';
import promisePlugin from 'eslint-plugin-promise';
import vitestPlugin from 'eslint-plugin-vitest';
import reactRefreshPlugin from 'eslint-plugin-react-refresh';

export default [
  // Base ESLint recommended rules
  js.configs.recommended,

  // Browser environment globals
  {
    languageOptions: {
      globals: {
        window: 'readonly',
        document: 'readonly',
        localStorage: 'readonly',
        sessionStorage: 'readonly',
        console: 'readonly',
        fetch: 'readonly',
        setTimeout: 'readonly',
        clearTimeout: 'readonly',
        setInterval: 'readonly',
        clearInterval: 'readonly',
        React: 'readonly',
        navigator: 'readonly',
        crypto: 'readonly',
        global: 'readonly',
        FormData: 'readonly',
        Blob: 'readonly',
        File: 'readonly',
        URL: 'readonly',
        URLSearchParams: 'readonly',
        AbortController: 'readonly',
        Headers: 'readonly',
        Request: 'readonly',
        Response: 'readonly',
        confirm: 'readonly',
        alert: 'readonly',
        prompt: 'readonly',
        atob: 'readonly',
        btoa: 'readonly',
        performance: 'readonly',
        process: 'readonly', // For Vite/Node.js compatibility
      },
    },
  },

  // TypeScript ESLint rules
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        project: './tsconfig.json',
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      ...tseslint.configs['recommended-requiring-type-checking'].rules,

      // Custom TypeScript rules - unused vars as warning to allow incremental fixing
      '@typescript-eslint/no-unused-vars': [
        'warn',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          ignoreRestSiblings: true,
        },
      ],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/no-explicit-any': 'warn', // Warn for starter project
      '@typescript-eslint/no-non-null-assertion': 'warn',
      '@typescript-eslint/prefer-nullish-coalescing': 'warn', // Warn instead of error
      '@typescript-eslint/prefer-optional-chain': 'warn', // Warn instead of error
      '@typescript-eslint/no-unsafe-assignment': 'warn',
      '@typescript-eslint/no-unsafe-member-access': 'warn',
      '@typescript-eslint/no-unsafe-call': 'warn',
      '@typescript-eslint/no-unsafe-return': 'warn',
      '@typescript-eslint/no-unsafe-argument': 'warn',
      '@typescript-eslint/no-floating-promises': 'warn',
      '@typescript-eslint/require-await': 'warn',
      '@typescript-eslint/unbound-method': 'warn', // Often false positive in React hooks
      '@typescript-eslint/no-misused-promises': 'warn', // Common in React onClick handlers
      '@typescript-eslint/only-throw-error': 'warn', // Sometimes throw strings for simplicity
    },
  },

  // React rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      react: reactPlugin,
      'react-hooks': reactHooksPlugin,
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
    rules: {
      ...reactPlugin.configs.recommended.rules,
      ...reactPlugin.configs['jsx-runtime'].rules,
      ...reactHooksPlugin.configs.recommended.rules,

      // Custom React rules
      'react/prop-types': 'off', // Using TypeScript for prop validation
      'react/jsx-uses-react': 'off', // Not needed with React 17+
      'react/react-in-jsx-scope': 'off', // Not needed with React 17+
      'react/jsx-filename-extension': [1, { extensions: ['.tsx', '.jsx'] }],
      'react/function-component-definition': 'off', // Relaxed for starter project
      'react/no-unescaped-entities': 'warn', // Warn for unescaped entities
      'react/display-name': 'warn', // Downgrade to warning
      'react-hooks/set-state-in-effect': 'warn', // Allow setState in effects for data fetching patterns
    },
  },

  // Accessibility rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      'jsx-a11y': jsxA11yPlugin,
    },
    rules: {
      ...jsxA11yPlugin.configs.recommended.rules,
      'jsx-a11y/no-autofocus': 'warn', // Sometimes needed for UX
      'jsx-a11y/no-redundant-roles': 'warn', // Low priority
      'jsx-a11y/anchor-has-content': 'warn', // Check manually
    },
  },

  // TanStack Query rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      '@tanstack/query': tanstackQueryPlugin,
    },
    rules: {
      ...tanstackQueryPlugin.configs.recommended.rules,
    },
  },

  // Import validation (sorting handled by Prettier's @ianvs/prettier-plugin-sort-imports)
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      import: importPlugin,
    },
    rules: {
      // Disabled: simple-import-sort conflicts with Prettier's import sorting plugin
      // causing circular fixes. Prettier handles import sorting via @ianvs/prettier-plugin-sort-imports
      'import/first': 'warn',
      'import/newline-after-import': 'warn',
      'import/no-duplicates': 'warn',
      'import/no-unused-modules': 'warn',
    },
    settings: {
      'import/resolver': {
        typescript: {
          alwaysTryTypes: true,
          project: './tsconfig.json',
        },
        node: {
          extensions: ['.js', '.jsx', '.ts', '.tsx'],
        },
      },
    },
  },

  // Modern JavaScript best practices
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      unicorn: unicornPlugin,
      promise: promisePlugin,
    },
    rules: {
      'unicorn/prefer-node-protocol': 'error',
      'unicorn/no-array-for-each': 'error',
      'unicorn/prefer-dom-node-text-content': 'error',
      'unicorn/prefer-modern-dom-apis': 'error',
      'promise/no-return-wrap': 'error',
      'promise/param-names': 'error',
      'promise/no-nesting': 'warn',
    },
  },

  // Security rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      security: securityPlugin,
    },
    rules: {
      'security/detect-object-injection': 'warn',
      'security/detect-non-literal-fs-filename': 'warn',
      'security/detect-possible-timing-attacks': 'warn',
    },
  },

  // React Fast Refresh compatibility
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      'react-refresh': reactRefreshPlugin,
    },
    rules: {
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
    },
  },

  // Testing rules
  {
    files: ['**/*.test.*', '**/*.spec.*', '**/__tests__/**'],
    languageOptions: {
      globals: {
        describe: 'readonly',
        it: 'readonly',
        expect: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly',
        beforeAll: 'readonly',
        afterAll: 'readonly',
        vi: 'readonly',
        test: 'readonly',
        jest: 'readonly',
      },
    },
    plugins: {
      vitest: vitestPlugin,
    },
    rules: {
      ...vitestPlugin.configs.recommended.rules,
      'vitest/no-focused-tests': 'error',
      'vitest/no-disabled-tests': 'warn',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',
      '@typescript-eslint/no-unsafe-return': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/unbound-method': 'off',
      '@typescript-eslint/only-throw-error': 'off', // Tests often throw strings for simplicity
      '@typescript-eslint/no-unused-vars': 'warn', // Test files often have unused imports for types
      '@typescript-eslint/no-unused-expressions': 'off', // Test assertions
      'react/display-name': 'off',
      'prefer-const': 'warn', // Less strict in tests
    },
  },

  // TanStack Router specific rules
  {
    files: ['**/routes/**/*.{ts,tsx}'],
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'warn', // Routes often export unused functions for TanStack Router
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          args: 'after-used',
          ignoreRestSiblings: true,
        },
      ],
      'react/display-name': 'warn', // Route components may not have display names
      // Add custom rules for route validation
    },
  },

  // Performance rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    rules: {
      'react/jsx-no-bind': 'warn',
      'react/jsx-no-constructed-context-values': 'warn',
      '@typescript-eslint/no-unnecessary-condition': 'warn',
      'unicorn/prefer-spread': 'warn',
    },
  },

  // Prettier integration (must be last)
  // Note: Prettier options are defined in .prettierrc - do not duplicate here
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      prettier: prettierPlugin,
    },
    rules: {
      ...prettierConfig.rules,
      'prettier/prettier': 'error',
    },
  },

  // General code quality rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    rules: {
      'no-console': 'warn',
      'no-debugger': 'error',
      'prefer-const': 'warn', // Downgrade to warning
      'no-var': 'error',
      'no-control-regex': 'warn', // Sometimes needed for sanitization
      '@typescript-eslint/no-redundant-type-constituents': 'warn',
      '@typescript-eslint/no-unused-expressions': 'warn',
    },
  },

  // Global ignores (replaces .eslintignore)
  {
    ignores: [
      'node_modules/',
      'dist/',
      'build/',
      '.vite/',
      '.tanstack/',
      'coverage/',
      '.vscode/',
      '*.config.js',
      '*.config.ts',
      '*.config.mjs',
      'vite.config.*',
      'tailwind.config.*',
      'postcss.config.*',
      'routeTree.gen.ts',
      'package.json',
      'package-lock.json',
      'tsconfig.json',
      'components.json',
      'vercel.json',
      '*.config.json',
      '.prettierrc',
      '.prettierignore',
      'nginx.conf',
      '.dockerignore',
      '.gitignore',
      '.commitlintrc.json',
      'app.css', // CSS files that ESLint can't parse
      '**/*.css', // All CSS files
      '**/*.json', // JSON files
      '**/*.md', // Markdown files
      '**/*.yml', // YAML files
      '**/*.yaml', // YAML files
      'app/test/setup.ts', // Test setup may not be in tsconfig project
      'app/test/setup.tsx', // Test setup may not be in tsconfig project
    ],
  },

  // Configuration files override
  {
    files: [
      '*.config.js',
      '*.config.ts',
      '*.config.mjs',
      'vite.config.*',
      'tailwind.config.*',
      'postcss.config.*',
    ],
    rules: {
      '@typescript-eslint/no-var-requires': 'off',
      '@typescript-eslint/no-require-imports': 'off',
    },
  },

  // Test files override
  {
    files: ['**/*.test.*', '**/*.spec.*', '**/__tests__/**'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'react/display-name': 'off',
    },
  },

  // Test setup files - disable type-checked rules that require tsconfig inclusion
  {
    files: ['**/test/setup.*', '**/test/*.ts', '**/test/*.tsx'],
    rules: {
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',
      '@typescript-eslint/no-unsafe-return': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/unbound-method': 'off',
    },
  },

  // E2E tests (Playwright) - Node.js environment
  {
    files: ['e2e/**/*.ts', 'e2e/**/*.js'],
    languageOptions: {
      globals: {
        process: 'readonly',
        Buffer: 'readonly',
        __dirname: 'readonly',
        __filename: 'readonly',
        module: 'readonly',
        require: 'readonly',
        exports: 'readonly',
        global: 'readonly',
      },
    },
    rules: {
      // E2E tests have different patterns
      'no-console': 'off',
      '@typescript-eslint/no-floating-promises': 'off', // Playwright handles async differently
      '@typescript-eslint/no-unused-vars': 'warn', // E2E tests may have unused vars
      'react-hooks/rules-of-hooks': 'off', // Not React components
      'security/detect-non-literal-fs-filename': 'off', // E2E tests use dynamic paths
    },
  },
];
