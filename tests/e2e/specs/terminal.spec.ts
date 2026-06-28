import { test, expect } from "@playwright/test";

test.describe("Terminal Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/terminal");
  });

  test("renders terminal component", async ({ page }) => {
    const terminal = page.locator(".xterm-viewport, .xterm");
    await expect(terminal.first()).toBeVisible({ timeout: 15000 });
  });

  test("shows connection status indicators", async ({ page }) => {
    const statusIndicator = page.locator(
      ".bg-gray-500, .bg-yellow-500, .bg-green-500, .bg-red-500"
    ).first();
    await expect(statusIndicator).toBeVisible({ timeout: 10000 });
  });

  test("renders terminal tabs", async ({ page }) => {
    const tabBar = page.locator(".tab-bar").first();
    await expect(tabBar).toBeVisible({ timeout: 10000 });
    await expect(page.locator(".tab-bar").getByText("Local")).toBeVisible();
  });

  test("adds a new terminal tab", async ({ page }) => {
    const addButton = page.locator('button[title="New tab"], button:has(svg.lucide-plus)');
    await addButton.click();
    await expect(page.locator(".tab-bar").getByText("Session 1")).toBeVisible({ timeout: 5000 });
  });

  test("switches between terminal tabs", async ({ page }) => {
    const addButton = page.locator('button[title="New tab"], button:has(svg.lucide-plus)');
    await addButton.click();
    await expect(page.locator(".tab-bar").getByText("Session 1")).toBeVisible();

    await page.locator(".tab-bar").getByText("Local").click();
    await expect(page.locator(".tab-bar").getByText("Local").first()).toBeVisible();
  });

  // Real command-input smoke test. Requires the API + mock SSH backend
  // running (docker-compose up). Skipped automatically when the desktop
  // shell backend isn't reachable — see tests/e2e/README.md.
  test("user dapat mengetik command dan lihat output", async ({ page, request }) => {
    // Optional pre-flight: confirm backend is up. Skip gracefully if not.
    try {
      const res = await request.get("http://127.0.0.1:18080/health", { timeout: 2000 });
      if (!res.ok()) test.skip(true, "API not healthy on :18080");
    } catch {
      test.skip(true, "API not reachable on :18080 — start docker-compose first");
    }

    const terminal = page.locator(".xterm").first();
    await expect(terminal).toBeVisible({ timeout: 15000 });

    // Focus the terminal canvas before typing — xterm.js needs a focused
    // surface to receive keyboard events.
    await terminal.click();

    const marker = `vexa-e2e-${Date.now()}`;
    await page.keyboard.type(`echo ${marker}`);
    await page.keyboard.press("Enter");

    // xterm renders to a canvas; we assert on a screen-line DOM node
    // that mirrors the terminal buffer for assertions.
    await expect(
      page.locator(".xterm").getByText(marker, { exact: false })
    ).toBeVisible({ timeout: 10000 });
  });
});
