import { test, expect } from '@playwright/test';
import testUser from '../fixtures/test-user.json';

test.describe("Login Page", () => {
  test.use({ storageState: { cookies: [], origins: [] } });
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('renders login form', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /welcome back/i })).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in|signin|login/i })).toBeVisible();
  });

  test('shows validation errors for empty form', async ({ page }) => {
    const submitButton = page.getByRole('button', { name: /sign in|signin|login/i });
    await submitButton.click();

    // Email validation: an error message appears
    await expect(page.getByText(/invalid email|required/i).first()).toBeVisible({ timeout: 5000 });
  });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.getByLabel(/email/i).fill('wrong@example.com');
    await page.getByLabel(/password/i).fill('wrongpassword');
    await page.getByRole('button', { name: /sign in|signin|login/i }).click();

    await expect(page.getByText(/invalid|error|failed/i).first()).toBeVisible({ timeout: 10000 });
  });

  test('redirects to dashboard on successful login', async ({ page }) => {
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.getByLabel(/password/i).fill(testUser.password);
    await page.getByRole('button', { name: /sign in|signin|login/i }).click();

    await expect(page).toHaveURL(/\/hosts|\/dashboard|\/mfa/, { timeout: 10000 });
  });

  test('has link to forgot password', async ({ page }) => {
    const forgotLink = page.getByRole('link', { name: /forgot|reset password/i });
    if (await forgotLink.isVisible()) {
      await forgotLink.click();
      await expect(page).toHaveURL(/forgot|reset/, { timeout: 10000 });
    }
  });
});
