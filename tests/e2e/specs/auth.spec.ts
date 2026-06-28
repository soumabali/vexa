import { test, expect, Page, APIResponse } from "@playwright/test";
import { authenticator } from "otplib";
import testUser from "../fixtures/test-user.json";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Login (step 1) and complete MFA verification if required.
 * Returns the final URL after successful authentication.
 */
async function loginWithCredentials(page: Page): Promise<string> {
  await page.goto("/login");
  await page.getByLabel(/email/i).fill(testUser.email);
  await page.getByLabel(/password/i).fill(testUser.password);
  await page.getByRole("button", { name: /sign in|signin|login/i }).click();

  // Wait for either MFA step or hosts redirect
  try {
    await page.waitForURL(/\/hosts|\/dashboard|\/mfa/, { timeout: 10_000 });
  } catch {
    // stay on /login — probably bad credentials; surface the failure
    throw new Error("Login did not redirect after submit");
  }

  if (page.url().includes("/mfa") || (await page.locator("#totp-code").count())) {
    // MFA already enabled — fetch pending session then submit code
    const mfaToken = await page.evaluate(() => {
      const cookies = document.cookie;
      const match = cookies.split("; ").find((c) => c.startsWith("mfa_token="));
      return match ? decodeURIComponent(match.split("=")[1]) : "";
    });

    // Need the user's TOTP secret — fetch it via the API using current cookie jar
    const setupResponse = await page.request.post("/api/v1/auth/mfa/setup");
    let secret = "";
    if (setupResponse.ok()) {
      const body = await setupResponse.json();
      secret = body.secret;
    }

    const code = secret ? authenticator.generate(secret) : "000000";
    await page.locator("#totp-code").fill(code);
  }

  await page.waitForURL(/\/hosts|\/dashboard/, { timeout: 15_000 });
  return page.url();
}

async function logout(page: Page) {
  // Best-effort: clear cookies + local storage; tests rely on storage state reset
  await page.context().clearCookies();
  await page.evaluate(() => {
    try {
      localStorage.clear();
      sessionStorage.clear();
    } catch {}
  });
}

/**
 * Intercept the MFA setup API response so we can read the TOTP secret
 * without OCR'ing the QR code.
 */
async function captureMFASetupSecret(page: Page): Promise<string> {
  const responsePromise = page.waitForResponse(
    (r: APIResponse) => r.url().includes("/api/v1/auth/mfa/setup") && r.request().method() === "POST"
  );
  // Trigger the setup call by clicking the "Set up authenticator" button
  await page.getByRole("button", { name: /set up authenticator/i }).click();
  const response = await responsePromise;
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  return body.secret as string;
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

test.describe("MFA TOTP end-to-end", () => {
  test("user can enable MFA, logout, login with TOTP", async ({ page }) => {
    // 1. Reset state — make sure MFA is not already enabled
    await loginWithCredentials(page);

    // 2. Navigate to security settings
    await page.goto("/settings/security");
    await expect(page.locator('h1:has-text("Security")')).toBeVisible();

    // 3. If MFA is already on, disable it first so we have a clean slate.
    //    The disable button is rendered only when MFA is currently enabled.
    const disableButton = page.getByRole("button", { name: /disable mfa/i });
    if (await disableButton.count()) {
      await disableButton.first().click();
      // Confirm in the destructive dialog
      await page.getByRole("button", { name: /disable mfa/i }).last().click();
      await expect(page.getByText(/mfa disabled/i)).toBeVisible({ timeout: 10_000 });
    }

    // 4. Click "Set up authenticator" → dialog appears
    await page.getByRole("button", { name: /set up authenticator/i }).click();
    await expect(page.getByRole("heading", { name: /setup two-factor authentication/i })).toBeVisible();

    // 5. Wait for the setup API and capture the secret
    const secret = await captureMFASetupSecret(page);

    // 6. Verify the dialog rendered QR + secret + backup codes
    await expect(page.locator('img[alt="TOTP QR Code"]')).toBeVisible();
    await expect(page.locator("code").filter({ hasText: /[A-Z2-7]{16,}/ })).toBeVisible();
    await expect(page.getByText(/save these backup codes/i)).toBeVisible();

    // 7. Generate a TOTP code from the captured secret and submit
    const totpCode = authenticator.generate(secret);
    await page.locator('input[name="totp"], input[placeholder="000000"]').first().fill(totpCode);
    await page.getByRole("button", { name: /verify|enable|activate|submit/i }).first().click();

    // 8. Expect MFA enabled confirmation (badge / success toast)
    await expect(page.getByText(/mfa enabled|2fa enabled|two-factor enabled/i)).toBeVisible({
      timeout: 10_000,
    });

    // 9. Logout and log back in — MFA step must appear
    await logout(page);
    await page.goto("/login");
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.getByLabel(/password/i).fill(testUser.password);
    await page.getByRole("button", { name: /sign in|signin|login/i }).click();

    // 10. MFA step appears
    await expect(page.locator("#totp-code")).toBeVisible({ timeout: 10_000 });

    // 11. Submit valid code → redirect to /hosts
    const loginTotp = authenticator.generate(secret);
    await page.locator("#totp-code").fill(loginTotp);
    await page.waitForURL(/\/hosts|\/dashboard/, { timeout: 15_000 });
  });

  test("user can disable MFA from security page", async ({ page }) => {
    // 1. Login (with MFA if enabled — helper handles it)
    await loginWithCredentials(page);

    // 2. Navigate to security settings
    await page.goto("/settings/security");
    await expect(page.locator('h1:has-text("Security")')).toBeVisible();

    // 3. Click Disable
    const disableButton = page.getByRole("button", { name: /disable mfa/i }).first();
    await disableButton.click();

    // 4. Confirm in the destructive dialog
    const confirmButton = page.getByRole("button", { name: /disable mfa/i }).last();
    await confirmButton.click();

    // 5. Expect success + UI refresh (badge "Disabled" / "Not enabled")
    await expect(page.getByText(/mfa disabled|not enabled|disabled/i)).toBeVisible({
      timeout: 10_000,
    });
  });

  test("invalid TOTP code is rejected with error", async ({ page }) => {
    // 1. Login (assume MFA enabled — helper covers the code path)
    await loginWithCredentials(page);

    // 2. Force re-prompt: logout and login again
    await logout(page);
    await page.goto("/login");
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.getByLabel(/password/i).fill(testUser.password);
    await page.getByRole("button", { name: /sign in|signin|login/i }).click();

    // 3. MFA step must be visible
    await expect(page.locator("#totp-code")).toBeVisible({ timeout: 10_000 });

    // 4. Submit a deliberately invalid code
    await page.locator("#totp-code").fill("000000");

    // 5. Expect error feedback (toast / inline message) and form stays visible
    await expect(
      page.getByText(/invalid|incorrect|wrong|expired|denied/i).first()
    ).toBeVisible({ timeout: 10_000 });
    await expect(page.locator("#totp-code")).toBeVisible();
    // Still on /mfa or /login (not /hosts)
    expect(page.url()).not.toMatch(/\/hosts/);
  });
});