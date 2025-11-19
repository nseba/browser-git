/**
 * Playwright test fixtures for browser-git tests
 * Provides shared setup and teardown for browser tests
 */

import { test as base, Page } from '@playwright/test';
import { clearAllStorage, createTestPage } from './helpers';

/**
 * Extended test fixtures with custom setup
 */
export type TestFixtures = {
  /**
   * Page with clean storage (all storage cleared before test)
   */
  cleanPage: Page;

  /**
   * Page with test HTML structure loaded
   */
  testPage: Page;
};

/**
 * Extended test with custom fixtures
 */
export const test = base.extend<TestFixtures>({
  cleanPage: async ({ page }, use) => {
    // Set up test page first to ensure proper context
    await createTestPage(page);

    // Clear all storage before test
    await clearAllStorage(page);

    // Use the page
    await use(page);

    // Clean up after test
    try {
      await clearAllStorage(page);
    } catch (e) {
      // Page might be closed or navigated, ignore cleanup errors
    }
  },

  testPage: async ({ page }, use) => {
    // Set up test page and clear storage
    await createTestPage(page);
    await clearAllStorage(page);

    // Use the page
    await use(page);

    // Clean up after test
    try {
      await clearAllStorage(page);
    } catch (e) {
      // Page might be closed or navigated, ignore cleanup errors
    }
  },
});

/**
 * Export expect from @playwright/test
 */
export { expect } from '@playwright/test';
