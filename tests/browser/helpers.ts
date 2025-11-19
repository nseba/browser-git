/**
 * Playwright test helper utilities for browser-git e2e tests
 */

import { Page } from '@playwright/test';

/**
 * Creates a simple test page with browser-git loaded
 */
export async function createTestPage(page: Page): Promise<void> {
  // Navigate to the test page served via HTTP
  // This provides a secure context for Web Crypto API and storage APIs
  await page.goto('/test-page.html', { waitUntil: 'domcontentloaded' });
}

/**
 * Checks if IndexedDB is available and working
 */
export async function checkIndexedDB(page: Page): Promise<boolean> {
  return await page.evaluate(async () => {
    if (typeof indexedDB === 'undefined') {
      return false;
    }

    try {
      const dbName = '__test_db__';
      const request = indexedDB.open(dbName, 1);

      return new Promise<boolean>((resolve) => {
        request.onsuccess = () => {
          request.result.close();
          indexedDB.deleteDatabase(dbName);
          resolve(true);
        };

        request.onerror = () => {
          resolve(false);
        };
      });
    } catch {
      return false;
    }
  });
}

/**
 * Checks if OPFS (Origin Private File System) is available
 */
export async function checkOPFS(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    return (
      typeof navigator !== 'undefined' &&
      'storage' in navigator &&
      'getDirectory' in navigator.storage
    );
  });
}

/**
 * Gets browser storage quota information
 */
export async function getStorageQuota(page: Page): Promise<{
  usage: number;
  quota: number;
  available: number;
} | null> {
  return await page.evaluate(async () => {
    if (!navigator.storage || !navigator.storage.estimate) {
      return null;
    }

    try {
      const estimate = await navigator.storage.estimate();
      const usage = estimate.usage || 0;
      const quota = estimate.quota || 0;
      const available = quota - usage;

      return { usage, quota, available };
    } catch {
      return null;
    }
  });
}

/**
 * Clears all browser storage (IndexedDB, localStorage, sessionStorage)
 */
export async function clearAllStorage(page: Page): Promise<void> {
  await page.evaluate(async () => {
    // Clear localStorage
    try {
      if (typeof localStorage !== 'undefined') {
        localStorage.clear();
      }
    } catch (e) {
      // Ignore SecurityError and other errors
      console.log('localStorage clear error:', e);
    }

    // Clear sessionStorage
    try {
      if (typeof sessionStorage !== 'undefined') {
        sessionStorage.clear();
      }
    } catch (e) {
      // Ignore SecurityError and other errors
      console.log('sessionStorage clear error:', e);
    }

    // Clear IndexedDB
    try {
      if (typeof indexedDB !== 'undefined') {
        const databases = await indexedDB.databases();
        for (const db of databases) {
          if (db.name) {
            indexedDB.deleteDatabase(db.name);
          }
        }
      }
    } catch (e) {
      // Ignore SecurityError and other errors
      console.log('IndexedDB clear error:', e);
    }

    // Clear OPFS if available
    try {
      if (
        typeof navigator !== 'undefined' &&
        'storage' in navigator &&
        'getDirectory' in navigator.storage
      ) {
        const root = await (navigator.storage as any).getDirectory();
        const entries = [];
        for await (const entry of (root as any).values()) {
          entries.push(entry);
        }
        for (const entry of entries) {
          await (root as any).removeEntry(entry.name, { recursive: true });
        }
      }
    } catch (e) {
      // Ignore SecurityError and other errors
      console.log('OPFS clear error:', e);
    }
  });
}

/**
 * Waits for a condition to be true in the browser context
 */
export async function waitForCondition(
  page: Page,
  condition: () => boolean | Promise<boolean>,
  timeout: number = 5000
): Promise<void> {
  await page.waitForFunction(condition, { timeout });
}

/**
 * Logs console messages from the browser to the test output
 */
export function setupConsoleLogging(page: Page): void {
  page.on('console', (msg) => {
    const type = msg.type();
    const text = msg.text();
    console.log(`[Browser ${type}]:`, text);
  });
}

/**
 * Captures and returns console errors
 */
export function captureConsoleErrors(page: Page): string[] {
  const errors: string[] = [];

  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      errors.push(msg.text());
    }
  });

  page.on('pageerror', (error) => {
    errors.push(error.message);
  });

  return errors;
}

/**
 * Measures performance of a browser operation
 */
export async function measurePerformance<T>(
  page: Page,
  operation: (page: Page) => Promise<T>
): Promise<{ result: T; duration: number }> {
  const start = await page.evaluate(() => performance.now());
  const result = await operation(page);
  const end = await page.evaluate(() => performance.now());
  const duration = end - start;

  return { result, duration };
}

/**
 * Simulates a network condition (online/offline)
 */
export async function setNetworkCondition(page: Page, online: boolean): Promise<void> {
  await page.context().setOffline(!online);
}

/**
 * Takes a screenshot with a timestamp
 */
export async function takeTimestampedScreenshot(page: Page, name: string): Promise<void> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  await page.screenshot({
    path: `test-results/screenshot-${name}-${timestamp}.png`,
    fullPage: true,
  });
}

/**
 * Injects a script into the page
 */
export async function injectScript(page: Page, scriptPath: string): Promise<void> {
  await page.addScriptTag({ path: scriptPath });
}

/**
 * Creates a blob URL from content
 */
export async function createBlobURL(page: Page, content: string, mimeType: string): Promise<string> {
  return await page.evaluate(
    ({ content, mimeType }) => {
      const blob = new Blob([content], { type: mimeType });
      return URL.createObjectURL(blob);
    },
    { content, mimeType }
  );
}

/**
 * Simulates a file download
 */
export async function simulateDownload(
  page: Page,
  content: string,
  filename: string
): Promise<void> {
  await page.evaluate(
    ({ content, filename }) => {
      const blob = new Blob([content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
    },
    { content, filename }
  );
}
