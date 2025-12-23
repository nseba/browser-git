import { defineConfig } from "vitest/config";

/**
 * Vitest configuration for benchmarks
 */
export default defineConfig({
  test: {
    // Use jsdom environment for browser-like APIs
    environment: "jsdom",

    // Enable globals like describe, it, expect
    globals: true,

    // Benchmark configuration
    benchmark: {
      include: ["benchmarks/**/*.bench.ts"],
      exclude: ["**/node_modules/**", "**/dist/**"],
    },

    // Include benchmark files
    include: ["benchmarks/**/*.bench.ts"],

    // Timeout for benchmarks (longer than tests)
    testTimeout: 60000,
    hookTimeout: 60000,
  },
});
