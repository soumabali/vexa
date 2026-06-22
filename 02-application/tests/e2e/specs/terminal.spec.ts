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
});
