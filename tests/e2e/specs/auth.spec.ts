import { test, expect, Page, Response as PWResponse } from "@playwright/test";
import { authenticator } from "otplib";
import testUser from "../fixtures/test-user.json";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Log in with credentials. Returns true if the inline MFA step appeared
 * (i.e. MFA is currently enabled), false if we landed straight on /hosts.
 */
async function loginStep1(page: Page): Promise<boolean> {
  await page.goto("/login");
  await page.getByLabel(/email/i).fill(testUser.email);
  await page.locator("input#password").fill(testUser.password);
  await page.getByRole("button", { name: /sign in/i }).click();

  try {
    await page.waitForURL(/\/hosts/, { timeout: 8_000 });
    return false; // no MFA step
  } catch {
    await expect(page.getByPlaceholder("000000").first()).toBeVisible({
      timeout: 8_000,
    });
    return true; // inline MFA step is showing
  }
}

async function logout(page: Page) {
  await page.context().clearCookies();
  await page.evaluate(() => {
    try {
      localStorage.clear();
      sessionStorage.clear();
    } catch {}
  });
}

/**
 * Open the MFA setup dialog ("Configure" → "Set up TOTP"), trigger the setup
 * API, and return the TOTP secret (no OCR of the QR code needed).
 */
async function enableMFA(page: Page): Promise<string> {
  await page.goto("/settings/security");
  await expect(
    page.locator("h1").filter({ hasText: /security/i }).first()
  ).toBeVisible();

  await page.getByRole("button", { name: /configure/i }).click();
  await expect(page.getByRole("heading", { name: /set up totp/i })).toBeVisible();

  const responsePromise = page.waitForResponse(
    (r: PWResponse) =>
      r.url().includes("/api/v1/auth/mfa/setup") &&
      r.request().method() === "POST"
  );
  await page.getByRole("button", { name: /continue to setup/i }).click();
  const response = await responsePromise;
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  const secret = body.secret as string;

  await expect(page.locator('img[alt="TOTP QR Code"]')).toBeVisible();
  await expect(
    page.locator("code").filter({ hasText: /[A-Z2-7]{16,}/ })
  ).toBeVisible();

  const totpCode = authenticator.generate(secret);
  await page.locator("input#code").fill(totpCode);
  await page.getByRole("button", { name: /enable 2fa/i }).click();

  await expect(
    page.getByText(/2fa enabled|is active/i).first()
  ).toBeVisible({ timeout: 10_000 });

  await page.getByRole("button", { name: /saved the backup codes/i }).click();
  return secret;
}

/**
 * Disable MFA from the security page using a known TOTP secret.
 */
async function disableMFA(page: Page, secret: string) {
  await page.goto("/settings/security");
  await expect(
    page.locator("h1").filter({ hasText: /security/i }).first()
  ).toBeVisible();

  await page.getByRole("button", { name: /^disable$/i }).first().click();
  await expect(page.getByRole("heading", { name: /disable mfa/i })).toBeVisible();

  const code = authenticator.generate(secret);
  await page.locator("input#disable-totp-code").fill(code);
  await page.getByRole("button", { name: /disable mfa/i }).last().click();

  await expect(page.getByText(/mfa disabled/i)).toBeVisible({ timeout: 10_000 });
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

test.describe("MFA TOTP end-to-end", () => {
  test("user can enable MFA, logout, login with TOTP", async ({ page }) => {
    // Precondition: MFA must be off. If it's on, disable it first via a fresh
    // setup capture is impossible — instead bail with a clear skip.
    const alreadyMfa = await loginStep1(page);
    if (alreadyMfa) {
      test.skip(true, "MFA already enabled — run 'disable MFA' test first");
    }

    const secret = await enableMFA(page);

    // Logout and log back in — inline MFA step must appear
    await logout(page);
    await page.goto("/login");
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.locator("input#password").fill(testUser.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page.getByPlaceholder("000000").first()).toBeVisible({
      timeout: 10_000,
    });

    const loginTotp = authenticator.generate(secret);
    await page.getByPlaceholder("000000").first().fill(loginTotp);
    await page.getByRole("button", { name: /verify/i }).click();
    await page.waitForURL(/\/hosts/, { timeout: 15_000 });

    // Cleanup: leave the test user in a disabled state for future runs.
    await disableMFA(page, secret);
  });

  test("user can disable MFA from security page", async ({ page }) => {
    // Ensure MFA is on (enable it if not), capturing the secret so we can
    // produce the confirmation code deterministically.
    const alreadyMfa = await loginStep1(page);
    const secret = alreadyMfa
      ? ""
      : await enableMFA(page);

    if (alreadyMfa) {
      // MFA was already enabled by a prior run; without the secret we cannot
      // produce a disable code deterministically.
      test.skip(true, "MFA pre-enabled and secret unknown — skip");
    }

    // Now disable it.
    await page.goto("/settings/security");
    await expect(
      page.locator("h1").filter({ hasText: /security/i }).first()
    ).toBeVisible();

    await page.getByRole("button", { name: /^disable$/i }).first().click();
    await expect(page.getByRole("heading", { name: /disable mfa/i })).toBeVisible();

    const code = authenticator.generate(secret);
    await page.locator("input#disable-totp-code").fill(code);
    await page.getByRole("button", { name: /disable mfa/i }).last().click();

    await expect(page.getByText(/mfa disabled/i)).toBeVisible({ timeout: 10_000 });
  });

  test("invalid TOTP code is rejected with error", async ({ page }) => {
    // Ensure MFA is on so the inline step appears.
    const alreadyMfa = await loginStep1(page);
    let secret = "";
    if (!alreadyMfa) {
      secret = await enableMFA(page);
    }

    // Logout and log back in to reach the inline MFA step.
    await logout(page);
    await page.goto("/login");
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.locator("input#password").fill(testUser.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page.getByPlaceholder("000000").first()).toBeVisible({
      timeout: 10_000,
    });

    // Submit a deliberately invalid code.
    await page.getByPlaceholder("000000").first().fill("000000");
    await page.getByRole("button", { name: /verify/i }).click();

    await expect(
      page.getByText(/invalid|incorrect|wrong|expired|failed|denied/i).first()
    ).toBeVisible({ timeout: 10_000 });
    expect(page.url()).not.toMatch(/\/hosts/);

    // Cleanup: disable MFA to restore the default state (only if we know the
    // secret, i.e. we enabled it in this test).
    if (secret) {
      await page.goto("/settings/security");
      await disableMFA(page, secret);
    }
  });
});
