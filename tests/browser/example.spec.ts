import { test, expect } from '@playwright/test';
import { createTestPage } from './helpers';

/**
 * Example Playwright test to verify setup
 * This test will be replaced with actual browser-git tests
 */
test.describe('Playwright Setup', () => {
  test('should verify browser environment', async ({ page }) => {
    // Set up test page
    await createTestPage(page);

    // Verify the page loaded
    await expect(page.locator('h1')).toHaveText('Browser Git Test');
  });

  test('should have IndexedDB support', async ({ page }) => {
    // Set up test page
    await createTestPage(page);

    // Check that IndexedDB is available
    const hasIndexedDB = await page.evaluate(() => {
      return typeof indexedDB !== 'undefined';
    });

    expect(hasIndexedDB).toBe(true);
  });

  test('should have localStorage API available', async ({ page }) => {
    // Set up test page
    await createTestPage(page);

    // Check that localStorage API exists
    const hasLocalStorageAPI = await page.evaluate(() => {
      return typeof Storage !== 'undefined' && typeof localStorage !== 'undefined';
    });

    expect(hasLocalStorageAPI).toBe(true);
  });
});
