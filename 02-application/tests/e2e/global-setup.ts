import { chromium, FullConfig } from "@playwright/test";
import testUser from "./fixtures/test-user.json";

async function globalSetup(config: FullConfig) {
  const { baseURL, storageState } = config.projects[0].use;
  if (storageState && typeof storageState === "string") {
    const browser = await chromium.launch();
    const context = await browser.newContext({ baseURL });
    const page = await context.newPage();

    await page.goto("/login");
    await page.getByLabel(/email/i).fill(testUser.email);
    await page.getByLabel(/password/i).fill(testUser.password);
    await page.getByRole("button", { name: /sign in|signin|login/i }).click();
    await page.waitForURL(/\/hosts|\/dashboard|\/mfa/, { timeout: 10000 });

    await context.storageState({ path: storageState });
    await browser.close();
  }
}

export default globalSetup;
