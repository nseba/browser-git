/**
 * Cross-browser tests for storage adapters
 * Tests IndexedDB, OPFS, localStorage, and memory adapters across all browsers
 */

import { test, expect } from './fixtures';
import { checkIndexedDB, checkOPFS, getStorageQuota, setupConsoleLogging } from './helpers';

test.describe('Storage Adapters - Cross Browser', () => {
  test.beforeEach(async ({ page }) => {
    setupConsoleLogging(page);
  });

  test.describe('Feature Detection', () => {
    test('should detect IndexedDB availability', async ({ page }) => {
      await page.goto('about:blank');
      const hasIndexedDB = await checkIndexedDB(page);

      // IndexedDB should be available in all modern browsers
      expect(hasIndexedDB).toBe(true);
    });

    test('should report OPFS availability', async ({ page, browserName }) => {
      await page.goto('about:blank');
      const hasOPFS = await checkOPFS(page);

      // OPFS may not be available in all browsers
      // Chrome/Edge: available
      // Firefox: available in recent versions
      // Safari: not available
      console.log(`OPFS available in ${browserName}:`, hasOPFS);
    });

    test('should have localStorage API', async ({ page }) => {
      await page.goto('about:blank');

      const hasLocalStorage = await page.evaluate(() => {
        return typeof localStorage !== 'undefined';
      });

      expect(hasLocalStorage).toBe(true);
    });

    test('should report storage quota', async ({ page }) => {
      await page.goto('about:blank');
      const quota = await getStorageQuota(page);

      if (quota) {
        expect(quota.quota).toBeGreaterThan(0);
        expect(quota.usage).toBeGreaterThanOrEqual(0);
        expect(quota.available).toBeGreaterThan(0);
        console.log('Storage quota:', quota);
      }
    });
  });

  test.describe('IndexedDB Adapter', () => {
    test('should create and open database', async ({ cleanPage }) => {
      // cleanPage already navigates to the test page via fixture, no need to goto again

      const result = await cleanPage.evaluate(async () => {
        return new Promise<{ success: boolean; error?: string }>((resolve) => {
          const dbName = 'test-browser-git';
          const request = indexedDB.open(dbName, 1);

          request.onupgradeneeded = (event: any) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('objects')) {
              db.createObjectStore('objects', { keyPath: 'key' });
            }
          };

          request.onsuccess = () => {
            request.result.close();
            indexedDB.deleteDatabase(dbName);
            resolve({ success: true });
          };

          request.onerror = () => {
            resolve({ success: false, error: request.error?.message });
          };
        });
      });

      expect(result.success).toBe(true);
    });

    test('should write and read data', async ({ cleanPage }) => {
      // cleanPage already navigates to the test page via fixture, no need to goto again

      const result = await cleanPage.evaluate(async () => {
        return new Promise<{ success: boolean; data?: string }>((resolve) => {
          const dbName = 'test-rw';
          const request = indexedDB.open(dbName, 1);

          request.onupgradeneeded = (event: any) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('data')) {
              db.createObjectStore('data', { keyPath: 'key' });
            }
          };

          request.onsuccess = () => {
            const db = request.result;
            const tx = db.transaction('data', 'readwrite');
            const store = tx.objectStore('data');

            // Write data
            store.put({ key: 'test-key', value: 'test-value' });

            tx.oncomplete = () => {
              // Read data
              const readTx = db.transaction('data', 'readonly');
              const readStore = readTx.objectStore('data');
              const getRequest = readStore.get('test-key');

              getRequest.onsuccess = () => {
                db.close();
                indexedDB.deleteDatabase(dbName);
                resolve({ success: true, data: getRequest.result?.value });
              };

              getRequest.onerror = () => {
                db.close();
                resolve({ success: false });
              };
            };

            tx.onerror = () => {
              db.close();
              resolve({ success: false });
            };
          };

          request.onerror = () => {
            resolve({ success: false });
          };
        });
      });

      expect(result.success).toBe(true);
      expect(result.data).toBe('test-value');
    });

    test('should handle large binary data', async ({ cleanPage }) => {
      // cleanPage already navigates to the test page via fixture, no need to goto again

      const result = await cleanPage.evaluate(async () => {
        return new Promise<{ success: boolean; size?: number }>((resolve) => {
          const dbName = 'test-binary';
          const request = indexedDB.open(dbName, 1);

          request.onupgradeneeded = (event: any) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('binary')) {
              db.createObjectStore('binary', { keyPath: 'key' });
            }
          };

          request.onsuccess = () => {
            const db = request.result;

            // Create 1MB of binary data
            const size = 1024 * 1024;
            const buffer = new ArrayBuffer(size);
            const view = new Uint8Array(buffer);
            for (let i = 0; i < 100; i++) {
              view[i] = i % 256;
            }

            const tx = db.transaction('binary', 'readwrite');
            const store = tx.objectStore('binary');
            store.put({ key: 'binary-data', value: buffer });

            tx.oncomplete = () => {
              const readTx = db.transaction('binary', 'readonly');
              const readStore = readTx.objectStore('binary');
              const getRequest = readStore.get('binary-data');

              getRequest.onsuccess = () => {
                const data = getRequest.result?.value;
                db.close();
                indexedDB.deleteDatabase(dbName);
                resolve({
                  success: true,
                  size: data instanceof ArrayBuffer ? data.byteLength : 0
                });
              };
            };
          };

          request.onerror = () => {
            resolve({ success: false });
          };
        });
      });

      expect(result.success).toBe(true);
      expect(result.size).toBe(1024 * 1024);
    });

    test('should handle concurrent transactions', async ({ cleanPage }) => {
      // cleanPage already navigates to the test page via fixture, no need to goto again

      const result = await cleanPage.evaluate(async () => {
        return new Promise<{ success: boolean; count?: number }>((resolve) => {
          const dbName = 'test-concurrent';
          const request = indexedDB.open(dbName, 1);

          request.onupgradeneeded = (event: any) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('items')) {
              db.createObjectStore('items', { keyPath: 'id' });
            }
          };

          request.onsuccess = () => {
            const db = request.result;
            const promises: Promise<void>[] = [];

            // Write 10 items concurrently
            for (let i = 0; i < 10; i++) {
              const promise = new Promise<void>((res) => {
                const tx = db.transaction('items', 'readwrite');
                const store = tx.objectStore('items');
                store.put({ id: `item-${i}`, value: `value-${i}` });
                tx.oncomplete = () => res();
              });
              promises.push(promise);
            }

            Promise.all(promises).then(() => {
              // Count items
              const tx = db.transaction('items', 'readonly');
              const store = tx.objectStore('items');
              const countRequest = store.count();

              countRequest.onsuccess = () => {
                db.close();
                indexedDB.deleteDatabase(dbName);
                resolve({ success: true, count: countRequest.result });
              };
            });
          };

          request.onerror = () => {
            resolve({ success: false });
          };
        });
      });

      expect(result.success).toBe(true);
      expect(result.count).toBe(10);
    });
  });

  test.describe('localStorage Adapter', () => {
    test('should write and read data', async ({ cleanPage }) => {
      // cleanPage fixture already sets up the test page, no need to navigate
      const result = await cleanPage.evaluate(() => {
        try {
          localStorage.setItem('test-key', 'test-value');
          const value = localStorage.getItem('test-key');
          localStorage.removeItem('test-key');
          return { success: true, value };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.value).toBe('test-value');
    });

    test('should handle size limits gracefully', async ({ cleanPage }) => {
      // cleanPage fixture already sets up the test page, no need to navigate
      const result = await cleanPage.evaluate(() => {
        try {
          // Try to store 10MB (likely to exceed quota in most browsers)
          const largeData = 'x'.repeat(10 * 1024 * 1024);
          localStorage.setItem('large-key', largeData);
          return { success: true, exceeded: false };
        } catch (error: any) {
          // QuotaExceededError is expected
          return {
            success: true,
            exceeded: true,
            errorName: error.name
          };
        }
      });

      expect(result.success).toBe(true);
      // Most browsers will throw QuotaExceededError
      if (result.exceeded) {
        expect(result.errorName).toMatch(/quota/i);
      }
    });

    test('should handle JSON serialization', async ({ cleanPage }) => {
      // cleanPage fixture already sets up the test page, no need to navigate
      const result = await cleanPage.evaluate(() => {
        try {
          const obj = { foo: 'bar', num: 42, arr: [1, 2, 3] };
          localStorage.setItem('json-key', JSON.stringify(obj));
          const retrieved = JSON.parse(localStorage.getItem('json-key') || '{}');
          localStorage.removeItem('json-key');
          return { success: true, data: retrieved };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      expect(result.success).toBe(true);
      expect(result.data).toEqual({ foo: 'bar', num: 42, arr: [1, 2, 3] });
    });
  });

  test.describe('OPFS Adapter', () => {
    test.skip(({ browserName }) => browserName === 'webkit', 'OPFS not supported in WebKit');

    test('should create and write files', async ({ cleanPage, browserName }) => {
      test.skip(browserName === 'webkit', 'OPFS not supported in WebKit');

      // cleanPage already navigates to the test page via fixture, no need to goto again
      const hasOPFS = await checkOPFS(cleanPage);

      test.skip(!hasOPFS, 'OPFS not available');

      const result = await cleanPage.evaluate(async () => {
        try {
          const root = await (navigator.storage as any).getDirectory();
          const fileHandle = await root.getFileHandle('test.txt', { create: true });
          const writable = await fileHandle.createWritable();
          await writable.write('Hello OPFS');
          await writable.close();

          // Read back
          const file = await fileHandle.getFile();
          const text = await file.text();

          // Cleanup
          await root.removeEntry('test.txt');

          return { success: true, content: text };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      if (hasOPFS) {
        expect(result.success).toBe(true);
        expect(result.content).toBe('Hello OPFS');
      }
    });

    test('should create directory hierarchy', async ({ cleanPage, browserName }) => {
      test.skip(browserName === 'webkit', 'OPFS not supported in WebKit');

      // cleanPage already navigates to the test page via fixture, no need to goto again
      const hasOPFS = await checkOPFS(cleanPage);

      test.skip(!hasOPFS, 'OPFS not available');

      const result = await cleanPage.evaluate(async () => {
        try {
          const root = await (navigator.storage as any).getDirectory();
          const dirHandle = await root.getDirectoryHandle('test-dir', { create: true });
          const subDirHandle = await dirHandle.getDirectoryHandle('sub-dir', { create: true });
          const fileHandle = await subDirHandle.getFileHandle('file.txt', { create: true });

          const writable = await fileHandle.createWritable();
          await writable.write('nested file');
          await writable.close();

          // Cleanup
          await root.removeEntry('test-dir', { recursive: true });

          return { success: true };
        } catch (error: any) {
          return { success: false, error: error.message };
        }
      });

      if (hasOPFS) {
        expect(result.success).toBe(true);
      }
    });
  });

  test.describe('Performance Comparison', () => {
    test('should measure IndexedDB write performance', async ({ cleanPage }) => {
      // cleanPage already navigates to the test page via fixture, no need to goto again

      const result = await cleanPage.evaluate(async () => {
        const dbName = 'perf-test';
        const request = indexedDB.open(dbName, 1);

        return new Promise<{ duration: number; opsPerSecond: number }>((resolve) => {
          request.onupgradeneeded = (event: any) => {
            const db = event.target.result;
            db.createObjectStore('data', { keyPath: 'key' });
          };

          request.onsuccess = () => {
            const db = request.result;
            const start = performance.now();
            const tx = db.transaction('data', 'readwrite');
            const store = tx.objectStore('data');

            // Write 100 items
            for (let i = 0; i < 100; i++) {
              store.put({ key: `key-${i}`, value: `value-${i}` });
            }

            tx.oncomplete = () => {
              const duration = performance.now() - start;
              db.close();
              indexedDB.deleteDatabase(dbName);
              resolve({
                duration,
                opsPerSecond: Math.round(100 / (duration / 1000))
              });
            };
          };
        });
      });

      console.log('IndexedDB write performance:', result);
      expect(result.duration).toBeGreaterThan(0);
      expect(result.opsPerSecond).toBeGreaterThan(0);
    });

    test('should measure localStorage write performance', async ({ cleanPage }) => {
      // cleanPage fixture already sets up the test page, no need to navigate
      const result = await cleanPage.evaluate(() => {
        const start = performance.now();

        // Write 100 items
        for (let i = 0; i < 100; i++) {
          localStorage.setItem(`key-${i}`, `value-${i}`);
        }

        // Cleanup
        for (let i = 0; i < 100; i++) {
          localStorage.removeItem(`key-${i}`);
        }

        const duration = performance.now() - start;
        return {
          duration,
          opsPerSecond: Math.round(100 / (duration / 1000))
        };
      });

      console.log('localStorage write performance:', result);
      expect(result.duration).toBeGreaterThan(0);
      expect(result.opsPerSecond).toBeGreaterThan(0);
    });
  });
});
