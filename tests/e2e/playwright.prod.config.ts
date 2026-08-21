import { defineConfig, devices } from '@playwright/test';

// Production E2E config — runs the suite against the live deployment.
// Usage:
//   BASE_URL=https://vexa.nexigo.my.id \
//   API_BASE_URL=https://api-vexa.nexigo.my.id \
//   npx playwright test --config=playwright.prod.config.ts
export default defineConfig({
  testDir: './specs',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 2,
  reporter: [['list'], ['html', { open: 'never' }]],
  timeout: 45000,
  expect: {
    timeout: 15000,
  },
  globalSetup: require.resolve('./global-setup'),
  use: {
    baseURL: process.env.BASE_URL || 'https://vexa.nexigo.my.id',
    storageState: './playwright/.auth/user.json',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'off',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
      },
    },
  ],
  // No webServer — we test the live deployment, not a local dev server.
});
