import js from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import reactPlugin from 'eslint-plugin-react';
import reactHooksPlugin from 'eslint-plugin-react-hooks';
import jsxA11yPlugin from 'eslint-plugin-jsx-a11y';
import tanstackQueryPlugin from '@tanstack/eslint-plugin-query';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
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

      // Custom TypeScript rules
      '@typescript-eslint/no-unused-vars': [
        'error',
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

  // Import sorting and validation
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    plugins: {
      'simple-import-sort': simpleImportSort,
      import: importPlugin,
    },
    rules: {
      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',
      'import/first': 'error',
      'import/newline-after-import': 'error',
      'import/no-duplicates': 'error',
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
    plugins: {
      vitest: vitestPlugin,
    },
    rules: {
      ...vitestPlugin.configs.recommended.rules,
      'vitest/no-focused-tests': 'error',
      'vitest/no-disabled-tests': 'warn',
      '@typescript-eslint/no-explicit-any': 'off',
      'react/display-name': 'off',
    },
  },

  // TanStack Router specific rules
  {
    files: ['**/routes/**/*.{ts,tsx}'],
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          args: 'after-used',
          ignoreRestSiblings: true,
        },
      ],
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
  {
    files: ['**/*.{ts,tsx,js,jsx,json,css,md}'],
    plugins: {
      prettier: prettierPlugin,
    },
    rules: {
      ...prettierConfig.rules,
      'prettier/prettier': [
        'error',
        {
          semi: true,
          trailingComma: 'es5',
          singleQuote: true,
          printWidth: 80,
          tabWidth: 2,
          useTabs: false,
          bracketSpacing: true,
          bracketSameLine: false,
          arrowParens: 'avoid',
          endOfLine: 'lf',
          quoteProps: 'as-needed',
          jsxSingleQuote: true,
        },
      ],
    },
  },

  // General code quality rules
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    rules: {
      'no-console': 'warn',
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',
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
];
