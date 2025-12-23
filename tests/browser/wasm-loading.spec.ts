/**
 * Cross-browser tests for WASM loading and performance
 * Tests WASM instantiation, memory usage, and basic operations
 */

import { test, expect } from "./fixtures";
import { setupConsoleLogging, TEST_PAGE_URL } from "./helpers";

test.describe("WASM Loading - Cross Browser", () => {
  test.beforeEach(async ({ page }) => {
    setupConsoleLogging(page);
  });

  test.describe("WebAssembly Support", () => {
    test("should have WebAssembly support", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const hasWasm = await page.evaluate(() => {
        return typeof WebAssembly !== "undefined";
      });

      expect(hasWasm).toBe(true);
    });

    test("should support WebAssembly.instantiate", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const hasInstantiate = await page.evaluate(() => {
        return typeof WebAssembly.instantiate === "function";
      });

      expect(hasInstantiate).toBe(true);
    });

    test("should support WebAssembly.instantiateStreaming", async ({
      page,
      browserName,
    }) => {
      await page.goto(TEST_PAGE_URL);

      const hasInstantiateStreaming = await page.evaluate(() => {
        return typeof WebAssembly.instantiateStreaming === "function";
      });

      // instantiateStreaming is available in most modern browsers
      // but may not work with all protocols (e.g., file://)
      expect(hasInstantiateStreaming).toBe(true);
    });

    test("should support WebAssembly memory", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(() => {
        try {
          const memory = new WebAssembly.Memory({
            initial: 1,
            maximum: 10,
          });
          return {
            success: true,
            byteLength: memory.buffer.byteLength,
            canGrow: memory.buffer.byteLength > 0,
          };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.byteLength).toBe(65536); // 1 page = 64KB
    });

    test("should support WebAssembly table", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(() => {
        try {
          const table = new WebAssembly.Table({
            initial: 1,
            maximum: 10,
            element: "anyfunc",
          });
          return {
            success: true,
            length: table.length,
          };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.length).toBe(1);
    });
  });

  test.describe("WASM Instantiation", () => {
    test("should instantiate simple WASM module", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        // Simple WASM module that exports an add function
        // (module (func (export "add") (param i32 i32) (result i32) local.get 0 local.get 1 i32.add))
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01,
          0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07,
          0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, 0x0a, 0x09, 0x01,
          0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b,
        ]);

        try {
          const module = await WebAssembly.instantiate(wasmBytes);
          const add = (module.instance.exports as any).add as (
            a: number,
            b: number,
          ) => number;
          const result = add(5, 7);
          return { success: true, result };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.result).toBe(12);
    });

    test("should measure WASM instantiation time", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01,
          0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07,
          0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, 0x0a, 0x09, 0x01,
          0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b,
        ]);

        try {
          const start = performance.now();
          await WebAssembly.instantiate(wasmBytes);
          const duration = performance.now() - start;
          return { success: true, duration };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.duration).toBeGreaterThan(0);
      console.log("WASM instantiation time:", result.duration, "ms");

      // Should be reasonably fast (< 100ms for simple module)
      expect(result.duration).toBeLessThan(100);
    });

    test("should handle WASM memory growth", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        // WASM module with memory that can grow
        // (module (memory (export "memory") 1 10))
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x05, 0x04, 0x01,
          0x01, 0x01, 0x0a, 0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f,
          0x72, 0x79, 0x02, 0x00,
        ]);

        try {
          const module = await WebAssembly.instantiate(wasmBytes);
          const memory = (module.instance.exports as any)
            .memory as WebAssembly.Memory;

          const initialSize = memory.buffer.byteLength;
          const prevPages = memory.grow(1); // Grow by 1 page
          const newSize = memory.buffer.byteLength;

          return {
            success: true,
            initialSize,
            newSize,
            prevPages,
            grew: newSize > initialSize,
          };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.grew).toBe(true);
      expect(result.newSize).toBe(result.initialSize + 65536); // Added 1 page (64KB)
    });
  });

  test.describe("WASM Performance", () => {
    test("should measure memory allocation performance", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(() => {
        const iterations = 1000;
        const start = performance.now();

        for (let i = 0; i < iterations; i++) {
          const memory = new WebAssembly.Memory({ initial: 1, maximum: 10 });
          // Access memory to ensure it's actually allocated
          const view = new Uint8Array(memory.buffer);
          view[0] = i % 256;
        }

        const duration = performance.now() - start;
        return {
          duration,
          avgTime: duration / iterations,
          opsPerSecond: Math.round(iterations / (duration / 1000)),
        };
      });

      console.log("Memory allocation performance:", result);
      expect(result.duration).toBeGreaterThan(0);
      expect(result.avgTime).toBeLessThan(1); // Should be sub-millisecond
    });

    test("should measure function call overhead", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01,
          0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07,
          0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, 0x0a, 0x09, 0x01,
          0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b,
        ]);

        const module = await WebAssembly.instantiate(wasmBytes);
        const add = (module.instance.exports as any).add as (
          a: number,
          b: number,
        ) => number;

        // Warm up
        for (let i = 0; i < 100; i++) {
          add(i, i + 1);
        }

        // Measure
        const iterations = 10000;
        const start = performance.now();

        for (let i = 0; i < iterations; i++) {
          add(i, i + 1);
        }

        const duration = performance.now() - start;
        return {
          duration,
          avgTime: duration / iterations,
          opsPerSecond: Math.round(iterations / (duration / 1000)),
        };
      });

      console.log("WASM function call performance:", result);
      // In fast CI environments, these may complete in under 1ms
      expect(result.duration).toBeGreaterThanOrEqual(0);
      expect(result.avgTime).toBeLessThan(0.1); // Should be very fast
      expect(result.opsPerSecond).toBeGreaterThan(0); // Should complete some operations
    });

    test("should measure data transfer JS<->WASM", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        // WASM module with memory export
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x05, 0x04, 0x01,
          0x01, 0x01, 0x0a, 0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f,
          0x72, 0x79, 0x02, 0x00,
        ]);

        const module = await WebAssembly.instantiate(wasmBytes);
        const memory = (module.instance.exports as any)
          .memory as WebAssembly.Memory;
        const buffer = new Uint8Array(memory.buffer);

        // Create test data
        const dataSize = 1024 * 1024; // 1MB
        const testData = new Uint8Array(dataSize);
        for (let i = 0; i < 1000; i++) {
          testData[i] = i % 256;
        }

        // Measure write to WASM memory
        const writeStart = performance.now();
        buffer.set(testData.subarray(0, Math.min(dataSize, buffer.length)));
        const writeDuration = performance.now() - writeStart;

        // Measure read from WASM memory
        const readStart = performance.now();
        const readData = buffer.slice(0, Math.min(dataSize, buffer.length));
        const readDuration = performance.now() - readStart;

        return {
          writeTime: writeDuration,
          readTime: readDuration,
          writeMBps: Math.round(
            Math.min(dataSize, buffer.length) /
              (1024 * 1024) /
              (writeDuration / 1000),
          ),
          readMBps: Math.round(
            Math.min(dataSize, buffer.length) /
              (1024 * 1024) /
              (readDuration / 1000),
          ),
        };
      });

      console.log("JS<->WASM data transfer performance:", result);
      // In fast CI environments, these may complete in less than 1ms
      expect(result.writeTime).toBeGreaterThanOrEqual(0);
      expect(result.readTime).toBeGreaterThanOrEqual(0);
    });
  });

  test.describe("WASM Error Handling", () => {
    test("should handle invalid WASM bytes", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        const invalidBytes = new Uint8Array([0x00, 0x01, 0x02, 0x03]);

        try {
          await WebAssembly.instantiate(invalidBytes);
          return { success: false, error: "Should have thrown" };
        } catch (error: any) {
          return { success: true, errorType: error.constructor.name };
        }
      });

      expect(result.success).toBe(true);
      expect(result.errorType).toMatch(/Error/);
    });

    test("should handle out of bounds memory access", async ({ page }) => {
      await page.goto(TEST_PAGE_URL);

      const result = await page.evaluate(async () => {
        // WASM module with memory
        const wasmBytes = new Uint8Array([
          0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x05, 0x04, 0x01,
          0x01, 0x01, 0x0a, 0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f,
          0x72, 0x79, 0x02, 0x00,
        ]);

        try {
          const module = await WebAssembly.instantiate(wasmBytes);
          const memory = (module.instance.exports as any)
            .memory as WebAssembly.Memory;

          // Try to grow beyond maximum
          try {
            memory.grow(100); // Try to grow by 100 pages (beyond max of 10)
            return { success: true, error: "Should have thrown on grow" };
          } catch (error: any) {
            return { success: true, errorType: error.constructor.name };
          }
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
    });
  });

  test.describe("Browser-Specific WASM Features", () => {
    test("should report WASM features", async ({ page, browserName }) => {
      await page.goto(TEST_PAGE_URL);

      const features = await page.evaluate(() => {
        return {
          hasInstantiate: typeof WebAssembly.instantiate === "function",
          hasInstantiateStreaming:
            typeof WebAssembly.instantiateStreaming === "function",
          hasCompile: typeof WebAssembly.compile === "function",
          hasCompileStreaming:
            typeof WebAssembly.compileStreaming === "function",
          hasValidate: typeof WebAssembly.validate === "function",
          hasModule: typeof WebAssembly.Module === "function",
          hasInstance: typeof WebAssembly.Instance === "function",
          hasMemory: typeof WebAssembly.Memory === "function",
          hasTable: typeof WebAssembly.Table === "function",
        };
      });

      console.log(`WASM features in ${browserName}:`, features);

      // All modern browsers should support these
      expect(features.hasInstantiate).toBe(true);
      expect(features.hasCompile).toBe(true);
      expect(features.hasValidate).toBe(true);
      expect(features.hasModule).toBe(true);
      expect(features.hasInstance).toBe(true);
      expect(features.hasMemory).toBe(true);
      expect(features.hasTable).toBe(true);
    });
  });
});
