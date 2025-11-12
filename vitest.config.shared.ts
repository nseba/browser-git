import { defineConfig } from 'vitest/config';

/**
 * Shared Vitest configuration for all packages
 * Individual packages should extend this configuration
 */
export default defineConfig({
  test: {
    // Use jsdom environment for browser-like APIs
    environment: 'jsdom',

    // Enable globals like describe, it, expect
    globals: true,

    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        '**/node_modules/**',
        '**/dist/**',
        '**/*.config.*',
        '**/test/**',
        '**/*.test.*',
        '**/*.spec.*',
      ],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 80,
        statements: 80,
      },
    },

    // Test file patterns
    include: ['**/*.{test,spec}.{ts,tsx}'],

    // Timeout settings
    testTimeout: 10000,
    hookTimeout: 10000,
  },
});
