import { test, expect, Page, Response as PWResponse } from "@playwright/test";
import { generateSync } from "otplib";
import testUser from "../fixtures/test-user.json";

/**
 * Generate a TOTP code for a base32 secret (otplib v13 API).
 */
function totp(secret: string): string {
  return generateSync({ secret });
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Fill credentials and submit. Returns true if the inline MFA step appeared
 * (MFA is enabled), false if we landed straight on /hosts.
 */
async function submitCredentials(page: Page): Promise<boolean> {
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
    return true; // inline MFA step showing
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
 * Enable MFA via the security page and return the TOTP secret.
 * Requires MFA to be currently DISABLED.
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

  await page.locator("input#code").fill(totp(secret));
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

  await page.locator("input#disable-totp-code").fill(totp(secret));
  await page.getByRole("button", { name: /disable mfa/i }).last().click();

  await expect(page.getByText(/mfa disabled/i)).toBeVisible({ timeout: 10_000 });
}

// ---------------------------------------------------------------------------
// Test suite
//
// These tests mutate the shared e2e-test user's MFA state, so they run
// serially and each leaves the account with MFA DISABLED (the default state
// that every other spec file relies on).
// ---------------------------------------------------------------------------

test.describe("MFA TOTP end-to-end", () => {
  test.describe.configure({ mode: "serial" });

  test("enable MFA, login with TOTP, reject invalid code, then disable", async ({
    page,
  }) => {
    // Precondition: MFA must be off.
    const alreadyMfa = await submitCredentials(page);
    if (alreadyMfa) {
      // Cannot recover the secret to disable it — skip rather than break the
      // rest of the suite.
      test.skip(true, "MFA already enabled with unknown secret — reset manually");
    }

    // 1. Enable MFA, capturing the secret.
    const secret = await enableMFA(page);

    // 2. Logout, log back in — inline MFA step must appear.
    await logout(page);
    const needsMfa = await submitCredentials(page);
    expect(needsMfa).toBe(true);
    await expect(page.getByPlaceholder("000000").first()).toBeVisible({
      timeout: 10_000,
    });

    // 3. Reject an invalid code.
    await page.getByPlaceholder("000000").first().fill("000000");
    await page.getByRole("button", { name: /verify/i }).click();
    await expect(
      page.getByText(/invalid|incorrect|wrong|expired|failed|denied/i).first()
    ).toBeVisible({ timeout: 10_000 });
    expect(page.url()).not.toMatch(/\/hosts/);

    // 4. Submit a valid code → redirect to /hosts.
    await page.getByPlaceholder("000000").first().fill(totp(secret));
    await page.getByRole("button", { name: /verify/i }).click();
    await page.waitForURL(/\/hosts/, { timeout: 15_000 });

    // 5. Cleanup: disable MFA to restore the default state for other specs.
    await disableMFA(page, secret);
  });
});
