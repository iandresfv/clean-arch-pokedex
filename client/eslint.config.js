import js from '@eslint/js';
import globals from 'globals';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import tseslint from 'typescript-eslint';
import prettier from 'eslint-config-prettier';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
import { defineConfig } from 'eslint/config';

export default defineConfig([
  // Global ignores
  {
    ignores: ['dist', 'node_modules', 'coverage', 'build'],
  },

  // Base configuration for all files
  js.configs.recommended,

  // TypeScript and React files
  {
    files: ['**/*.{ts,tsx}'],
    extends: [...tseslint.configs.strictTypeChecked, ...tseslint.configs.stylisticTypeChecked],
    languageOptions: {
      ecmaVersion: 2024,
      globals: globals.browser,
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json', './tsconfig.test.json'],
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
      'simple-import-sort': simpleImportSort,
    },
    rules: {
      // React Hooks rules
      ...reactHooks.configs.recommended.rules,

      // React Refresh
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],

      // Import sorting
      'simple-import-sort/imports': [
        'error',
        {
          groups: [
            // 1. React first
            ['^react$', '^react/'],
            // 2. Packages (node_modules)
            ['^@?\\w'],
            // 3. Parent imports (..)
            ['^\\.\\.'],
            // 4. Sibling imports (.)
            ['^\\./(?=.*/)(?!/?$)', '^\\.(?!/?$)', '^\\./?$'],
            // 5. Style imports (side effect CSS at the end)
            ['^.+\\.css$'],
            // 6. Side effect imports (without binding)
            ['^\\u0000'],
          ],
        },
      ],
      'simple-import-sort/exports': 'error',

      // TypeScript strict rules
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/consistent-type-imports': ['error', { prefer: 'type-imports' }],
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',

      // Allow empty interfaces (useful for extending types)
      '@typescript-eslint/no-empty-object-type': 'off',
    },
  },

  // Node.js config files
  {
    files: ['**/*.config.{js,ts}', 'vite.config.ts'],
    languageOptions: {
      globals: globals.node,
    },
  },

  // Prettier config (must be last to override conflicting rules)
  prettier,
]);
