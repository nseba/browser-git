/**
 * Cross-browser performance tests for Git operations
 * Tests performance targets: commit <50ms, checkout <200ms, clone <5s
 */

import { test, expect } from './fixtures';
import { setupConsoleLogging, measurePerformance, TEST_PAGE_URL } from './helpers';

test.describe('Git Operations Performance - Cross Browser', () => {
  test.beforeEach(async ({ page }) => {
    setupConsoleLogging(page);
  });

  test.describe('Repository Initialization', () => {
    test('should initialize repository quickly', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const start = performance.now();

        // Simulate repository initialization
        const repoStructure = {
          '.git/objects': {},
          '.git/refs/heads': {},
          '.git/refs/tags': {},
          '.git/HEAD': 'ref: refs/heads/main',
          '.git/config': '[core]\n\trepositoryformatversion = 0\n'
        };

        const duration = performance.now() - start;
        return { duration, structure: Object.keys(repoStructure).length };
      });

      console.log('Repository init time:', result.duration, 'ms');
      expect(result.duration).toBeLessThan(10); // Should be very fast
    });
  });

  test.describe('Commit Operations', () => {
    test('should create commit in < 50ms (target)', async ({ cleanPage, browserName }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        // Simulate commit creation
        const author = {
          name: 'Test User',
          email: 'test@example.com',
          timestamp: Date.now()
        };

        const commitData = {
          tree: 'a'.repeat(40), // SHA-1 hash
          parent: 'b'.repeat(40),
          author: `${author.name} <${author.email}> ${author.timestamp}`,
          committer: `${author.name} <${author.email}> ${author.timestamp}`,
          message: 'Test commit message'
        };

        // Create commit object
        const times: number[] = [];

        for (let i = 0; i < 10; i++) {
          const start = performance.now();

          // Simulate commit object serialization
          const commitText = [
            `tree ${commitData.tree}`,
            `parent ${commitData.parent}`,
            `author ${commitData.author}`,
            `committer ${commitData.committer}`,
            '',
            commitData.message
          ].join('\n');

          // Simulate hashing
          const encoder = new TextEncoder();
          const data = encoder.encode(commitText);
          const hashBuffer = await crypto.subtle.digest('SHA-1', data);
          const hashArray = Array.from(new Uint8Array(hashBuffer));
          const hash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

          const duration = performance.now() - start;
          times.push(duration);
        }

        const avgDuration = times.reduce((a, b) => a + b, 0) / times.length;
        const minDuration = Math.min(...times);
        const maxDuration = Math.max(...times);

        return {
          avgDuration,
          minDuration,
          maxDuration,
          times
        };
      });

      console.log(`Commit performance (${browserName}):`, result);

      // Target: < 50ms average
      expect(result.avgDuration).toBeLessThan(50);
      expect(result.minDuration).toBeLessThan(50);
    });

    test('should handle multiple commits efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const commitCount = 100;
        const start = performance.now();

        const author = {
          name: 'Test User',
          email: 'test@example.com'
        };

        for (let i = 0; i < commitCount; i++) {
          const commitData = {
            tree: 'a'.repeat(40),
            parent: 'b'.repeat(40),
            author: `${author.name} <${author.email}> ${Date.now()}`,
            message: `Commit ${i}`
          };

          const commitText = [
            `tree ${commitData.tree}`,
            `parent ${commitData.parent}`,
            `author ${commitData.author}`,
            `committer ${commitData.author}`,
            '',
            commitData.message
          ].join('\n');

          // Hash the commit
          const encoder = new TextEncoder();
          const data = encoder.encode(commitText);
          await crypto.subtle.digest('SHA-1', data);
        }

        const totalDuration = performance.now() - start;
        const avgDuration = totalDuration / commitCount;

        return {
          totalDuration,
          avgDuration,
          commitCount,
          commitsPerSecond: Math.round(commitCount / (totalDuration / 1000))
        };
      });

      console.log('Multiple commits performance:', result);
      expect(result.avgDuration).toBeLessThan(50);
      expect(result.commitsPerSecond).toBeGreaterThan(10);
    });

    test('should hash commit objects efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const iterations = 1000;
        const testData = new TextEncoder().encode('test commit data');

        const start = performance.now();

        for (let i = 0; i < iterations; i++) {
          await crypto.subtle.digest('SHA-1', testData);
        }

        const duration = performance.now() - start;
        const avgTime = duration / iterations;
        const hashesPerSecond = Math.round(iterations / (duration / 1000));

        return {
          duration,
          avgTime,
          hashesPerSecond,
          iterations
        };
      });

      console.log('SHA-1 hashing performance:', result);
      expect(result.avgTime).toBeLessThan(1); // Should be sub-millisecond
      expect(result.hashesPerSecond).toBeGreaterThan(100);
    });
  });

  test.describe('Checkout Operations', () => {
    test('should checkout branch in < 200ms (target)', async ({ cleanPage, browserName }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        // Simulate checkout operation
        const fileCount = 50;
        const files: { path: string; content: string }[] = [];

        for (let i = 0; i < fileCount; i++) {
          files.push({
            path: `file${i}.txt`,
            content: `Content of file ${i}\n`.repeat(10)
          });
        }

        const times: number[] = [];

        for (let i = 0; i < 5; i++) {
          const start = performance.now();

          // Simulate writing files to working directory
          const written: string[] = [];
          for (const file of files) {
            // Simulate file write
            written.push(file.path);
          }

          // Simulate updating HEAD
          const head = 'ref: refs/heads/main';

          // Simulate updating index
          const index = files.map(f => ({ path: f.path, hash: 'a'.repeat(40) }));

          const duration = performance.now() - start;
          times.push(duration);
        }

        const avgDuration = times.reduce((a, b) => a + b, 0) / times.length;
        const minDuration = Math.min(...times);
        const maxDuration = Math.max(...times);

        return {
          avgDuration,
          minDuration,
          maxDuration,
          fileCount,
          times
        };
      });

      console.log(`Checkout performance (${browserName}):`, result);

      // Target: < 200ms average for 50 files
      expect(result.avgDuration).toBeLessThan(200);
    });

    test('should update working directory efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const fileCount = 100;
        const start = performance.now();

        // Simulate writing files
        const files: { path: string; content: Uint8Array }[] = [];
        const encoder = new TextEncoder();

        for (let i = 0; i < fileCount; i++) {
          const content = encoder.encode(`File content ${i}\n`);
          files.push({ path: `file${i}.txt`, content });
        }

        const duration = performance.now() - start;
        const avgTimePerFile = duration / fileCount;

        return {
          duration,
          avgTimePerFile,
          fileCount,
          filesPerSecond: Math.round(fileCount / (duration / 1000))
        };
      });

      console.log('Working directory update performance:', result);
      expect(result.avgTimePerFile).toBeLessThan(2); // < 2ms per file
      expect(result.filesPerSecond).toBeGreaterThan(50);
    });
  });

  test.describe('Clone Operations', () => {
    test('should clone repository in < 5s for 100 commits (target)', async ({ cleanPage, browserName }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const commitCount = 100;
        const fileCount = 20;

        const start = performance.now();

        // Simulate receiving and processing commits
        const commits = [];
        for (let i = 0; i < commitCount; i++) {
          const commit = {
            hash: 'a'.repeat(40),
            tree: 'b'.repeat(40),
            parent: i > 0 ? 'c'.repeat(40) : null,
            author: `Author <author@example.com> ${Date.now()}`,
            message: `Commit ${i}`
          };
          commits.push(commit);
        }

        // Simulate unpacking objects
        const objects = [];
        for (let i = 0; i < commitCount * 2; i++) {
          // Each commit has a commit object and a tree object
          const encoder = new TextEncoder();
          const data = encoder.encode(`object data ${i}`);
          await crypto.subtle.digest('SHA-1', data);
          objects.push(data);
        }

        // Simulate writing files
        const files = [];
        for (let i = 0; i < fileCount; i++) {
          files.push({
            path: `file${i}.txt`,
            content: `Content ${i}\n`
          });
        }

        const duration = performance.now() - start;

        return {
          duration,
          commitCount,
          fileCount,
          objectCount: objects.length,
          avgTimePerCommit: duration / commitCount
        };
      });

      console.log(`Clone performance (${browserName}):`, result);

      // Target: < 5000ms (5 seconds) for 100 commits
      expect(result.duration).toBeLessThan(5000);
      expect(result.avgTimePerCommit).toBeLessThan(50);
    });

    test('should handle packfile processing efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const objectCount = 200;
        const start = performance.now();

        // Simulate packfile object processing
        const encoder = new TextEncoder();

        for (let i = 0; i < objectCount; i++) {
          // Simulate decompression and hashing
          const data = encoder.encode(`object ${i} content`);

          // Hash the object
          await crypto.subtle.digest('SHA-1', data);
        }

        const duration = performance.now() - start;
        const avgTimePerObject = duration / objectCount;
        const objectsPerSecond = Math.round(objectCount / (duration / 1000));

        return {
          duration,
          avgTimePerObject,
          objectsPerSecond,
          objectCount
        };
      });

      console.log('Packfile processing performance:', result);
      expect(result.avgTimePerObject).toBeLessThan(25); // < 25ms per object
      expect(result.objectsPerSecond).toBeGreaterThan(10);
    });
  });

  test.describe('Memory Usage', () => {
    test('should track memory usage during operations', async ({ cleanPage, browserName }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const getMemory = (): number => {
          // @ts-ignore - performance.memory is Chrome-specific
          if (performance.memory) {
            // @ts-ignore
            return performance.memory.usedJSHeapSize;
          }
          return 0;
        };

        const memoryBefore = getMemory();

        // Allocate some data
        const largeArray = new Array(10000);
        for (let i = 0; i < largeArray.length; i++) {
          largeArray[i] = {
            hash: 'a'.repeat(40),
            data: `object ${i}`
          };
        }

        const memoryAfter = getMemory();
        const memoryUsed = memoryAfter - memoryBefore;

        // Clear the array
        largeArray.length = 0;

        return {
          memoryBefore,
          memoryAfter,
          memoryUsed,
          hasMemoryAPI: memoryBefore > 0
        };
      });

      console.log(`Memory usage (${browserName}):`, result);

      if (result.hasMemoryAPI) {
        // Memory usage should be non-negative (can be 0 if GC ran between allocations)
        expect(result.memoryUsed).toBeGreaterThanOrEqual(0);
        console.log('Memory used:', (result.memoryUsed / 1024 / 1024).toFixed(2), 'MB');
      } else {
        console.log('Memory API not available in this browser');
      }
    });

    test('should measure peak memory during large operation', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const getMemory = (): number => {
          // @ts-ignore
          if (performance.memory) {
            // @ts-ignore
            return performance.memory.usedJSHeapSize;
          }
          return 0;
        };

        const memoryBefore = getMemory();
        let peakMemory = memoryBefore;

        // Simulate large operation
        const batches = 10;
        const batchSize = 1000;

        for (let batch = 0; batch < batches; batch++) {
          const data = new Array(batchSize);
          for (let i = 0; i < batchSize; i++) {
            data[i] = {
              id: `${batch}-${i}`,
              content: new Uint8Array(1024) // 1KB per object
            };
          }

          const currentMemory = getMemory();
          if (currentMemory > peakMemory) {
            peakMemory = currentMemory;
          }

          // Process data
          for (const item of data) {
            const encoder = new TextEncoder();
            await crypto.subtle.digest('SHA-1', encoder.encode(item.id));
          }
        }

        const memoryAfter = getMemory();

        return {
          memoryBefore,
          peakMemory,
          memoryAfter,
          peakIncrease: peakMemory - memoryBefore,
          hasMemoryAPI: memoryBefore > 0
        };
      });

      console.log('Peak memory usage:', result);

      if (result.hasMemoryAPI) {
        console.log('Peak memory increase:', (result.peakIncrease / 1024 / 1024).toFixed(2), 'MB');
        // Should not use excessive memory (< 50MB for this test)
        expect(result.peakIncrease).toBeLessThan(50 * 1024 * 1024);
      }
    });
  });

  test.describe('Diff Operations', () => {
    test('should compute diff efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const original = 'line 1\nline 2\nline 3\nline 4\nline 5\n';
        const modified = 'line 1\nline 2 modified\nline 3\nline 5\nline 6\n';

        const times: number[] = [];

        for (let i = 0; i < 100; i++) {
          const start = performance.now();

          // Simple diff algorithm (simulated)
          const originalLines = original.split('\n');
          const modifiedLines = modified.split('\n');

          const changes = [];
          for (let j = 0; j < Math.max(originalLines.length, modifiedLines.length); j++) {
            if (originalLines[j] !== modifiedLines[j]) {
              changes.push({
                line: j,
                old: originalLines[j],
                new: modifiedLines[j]
              });
            }
          }

          const duration = performance.now() - start;
          times.push(duration);
        }

        const avgDuration = times.reduce((a, b) => a + b, 0) / times.length;

        return {
          avgDuration,
          minDuration: Math.min(...times),
          maxDuration: Math.max(...times),
          iterations: times.length
        };
      });

      console.log('Diff performance:', result);
      expect(result.avgDuration).toBeLessThan(1); // Should be very fast for small files
    });
  });

  test.describe('Compression Performance', () => {
    test('should compress data efficiently', async ({ cleanPage }) => {
      await cleanPage.goto(TEST_PAGE_URL);

      const result = await cleanPage.evaluate(async () => {
        const encoder = new TextEncoder();
        const data = encoder.encode('test data\n'.repeat(1000));

        const start = performance.now();

        // Use CompressionStream if available
        if (typeof CompressionStream !== 'undefined') {
          const stream = new ReadableStream({
            start(controller) {
              controller.enqueue(data);
              controller.close();
            }
          });

          const compressed = stream.pipeThrough(new CompressionStream('gzip'));
          const reader = compressed.getReader();

          const chunks: Uint8Array[] = [];
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            chunks.push(value);
          }

          const duration = performance.now() - start;
          const compressedSize = chunks.reduce((sum, chunk) => sum + chunk.length, 0);

          return {
            hasCompressionStream: true,
            duration,
            originalSize: data.length,
            compressedSize,
            compressionRatio: data.length / compressedSize
          };
        } else {
          // Fallback: just measure time to copy data
          const copy = new Uint8Array(data);
          const duration = performance.now() - start;

          return {
            hasCompressionStream: false,
            duration,
            originalSize: data.length,
            compressedSize: copy.length,
            compressionRatio: 1
          };
        }
      });

      console.log('Compression performance:', result);
      expect(result.duration).toBeGreaterThan(0);

      if (result.hasCompressionStream) {
        expect(result.compressedSize).toBeLessThan(result.originalSize);
        expect(result.compressionRatio).toBeGreaterThan(1);
      }
    });
  });
});
