import { defineConfig, devices } from '@playwright/test';

// Production E2E config — runs the suite against the live deployment.
// Usage:
//   BASE_URL=https://vexa.nexigo.my.id \
//   API_BASE_URL=https://api-vexa.nexigo.my.id \
//   npx playwright test --config=playwright.prod.config.ts
export default defineConfig({
  testDir: './specs',
  // Specs mutate shared state (the e2e-test user's MFA flag, host list). The
  // auth.spec enables/disables MFA mid-run and must not overlap login.spec or
  // host-crud.spec, so run files sequentially (workers: 1). Tests within a
  // file that declare serial mode already order themselves.
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
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
