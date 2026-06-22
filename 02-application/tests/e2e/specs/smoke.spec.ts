import { test, expect } from '@playwright/test';

test('app responds and login page renders', async ({ page }) => {
  const response = await page.goto('/login');
  // Ensure the server did not return a 5xx error
  expect(response?.status()).toBeLessThan(500);
  // Sanity check: a heading should be present on the login page
  await expect(page.getByRole('heading').first()).toBeVisible({ timeout: 10000 });
});
