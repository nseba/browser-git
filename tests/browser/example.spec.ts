import { test, expect } from '@playwright/test';

/**
 * Example Playwright test to verify setup
 * This test will be replaced with actual browser-git tests
 */
test.describe('Playwright Setup', () => {
  test('should verify browser environment', async ({ page }) => {
    // Navigate to a data URL with a simple HTML page
    await page.goto('data:text/html,<h1>Browser Git Test</h1>');

    // Verify the page loaded
    await expect(page.locator('h1')).toHaveText('Browser Git Test');
  });

  test('should have IndexedDB support', async ({ page }) => {
    await page.goto('data:text/html,<html><body></body></html>');

    // Check that IndexedDB is available
    const hasIndexedDB = await page.evaluate(() => {
      return typeof indexedDB !== 'undefined';
    });

    expect(hasIndexedDB).toBe(true);
  });

  test('should have localStorage API available', async ({ page }) => {
    await page.goto('data:text/html,<html><body></body></html>');

    // Check that localStorage API exists (even if disabled in data: URLs)
    const hasLocalStorageAPI = await page.evaluate(() => {
      return typeof Storage !== 'undefined';
    });

    expect(hasLocalStorageAPI).toBe(true);
  });
});
