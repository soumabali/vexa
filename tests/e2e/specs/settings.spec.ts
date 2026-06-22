import { test, expect, Page } from "@playwright/test";
import testUser from "../fixtures/test-user.json";

async function login(page: Page) {
  await page.goto("/login");
  await page.getByLabel(/email/i).fill(testUser.email);
  await page.getByLabel(/password/i).fill(testUser.password);
  await page.getByRole("button", { name: /sign in|signin|login/i }).click();
  await page.waitForURL(/\/hosts|\/dashboard|\/mfa/, { timeout: 10000 });
}

test.describe("Settings Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/settings");
  });

  test("renders settings overview page", async ({ page }) => {
    await expect(page).toHaveURL("/settings");
    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
  });

  test("navigates to profile settings", async ({ page }) => {
    await page.locator('a[href="/settings/profile"]').click();
    await expect(page).toHaveURL("/settings/profile");
    await expect(page.locator('h1:has-text("Profile")')).toBeVisible();
  });

  test("navigates to security settings", async ({ page }) => {
    await page.goto("/settings");
    await page.locator('a[href="/settings/security"]').click();
    await expect(page).toHaveURL("/settings/security");
    await expect(page.locator('h1:has-text("Security")')).toBeVisible();
  });

  test("navigates to webauthn settings", async ({ page }) => {
    await page.goto("/settings");
    await page.locator('a[href="/settings/webauthn"]').click();
    await expect(page).toHaveURL("/settings/webauthn");
    await expect(page.locator('h1:has-text("Passkeys")')).toBeVisible();
  });
});
