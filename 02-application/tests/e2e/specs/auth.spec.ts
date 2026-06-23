import { test, expect } from "@playwright/test";

test.describe("MFA TOTP flow", () => {
  test("settings security shows TOTP setup dialog on click", async ({ page }) => {
    await page.goto("/settings/security");
    await expect(page.locator('h1:has-text("Security")')).toBeVisible();

    const setupButton = page.getByRole("button", { name: /set up authenticator/i });
    await expect(setupButton).toBeVisible();
    await setupButton.click();

    // The setup dialog should render the TwoFASetup card with the continue/setup button
    await expect(page.getByRole("heading", { name: /setup two-factor authentication/i })).toBeVisible();
  });
});
