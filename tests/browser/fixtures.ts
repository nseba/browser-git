/**
 * Playwright test fixtures for browser-git tests
 * Provides shared setup and teardown for browser tests
 */

import { test as base, Page } from "@playwright/test";
import { clearAllStorage, createTestPage } from "./helpers";

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
    // Clear all storage before test
    await clearAllStorage(page);

    // Use the page
    await use(page);

    // Clean up after test
    await clearAllStorage(page);
  },

  testPage: async ({ page }, use) => {
    // Clear storage and set up test page
    await clearAllStorage(page);
    await createTestPage(page);

    // Use the page
    await use(page);

    // Clean up after test
    await clearAllStorage(page);
  },
});

/**
 * Export expect from @playwright/test
 */
export { expect } from "@playwright/test";
