import { test, expect } from '@playwright/test';
import testUser from '../fixtures/test-user.json';

async function apiLogin() {
  const apiBase = process.env.API_BASE_URL || process.env.BASE_URL?.replace('3000', '8080') || 'http://localhost:8080';
  const res = await fetch(`${apiBase}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Origin': process.env.BASE_URL || 'http://localhost:3000' },
    body: JSON.stringify({ email: testUser.email, password: testUser.password }),
  });
  if (!res.ok) {
    throw new Error(`apiLogin failed: ${res.status} ${await res.text()}`);
  }
  const data = await res.json();
  return data.access_token as string;
}

async function cleanupHosts(token: string) {
  const apiBase = process.env.API_BASE_URL || process.env.BASE_URL?.replace('3000', '8080') || 'http://localhost:8080';
  const res = await fetch(`${apiBase}/api/v1/hosts`, {
    headers: { Authorization: `Bearer ${token}`, Origin: process.env.BASE_URL || 'http://localhost:3000' },
  });
  if (!res.ok) return;
  const data = await res.json();
  const hosts = data.hosts || [];
  for (const h of hosts) {
    if (h.name?.startsWith('E2E') || h.name?.startsWith('Debug')) {
      await fetch(`${apiBase}/api/v1/hosts/${h.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}`, Origin: process.env.BASE_URL || 'http://localhost:3000' },
      });
    }
  }
}

async function deleteAllHosts(token: string) {
  cleanupHosts(token);
}

test.describe("Host CRUD Operations", () => {
  test.beforeEach(async ({ page, context }) => {
    // Clean up previous test hosts via API
    const token = await apiLogin();
    await cleanupHosts(token);

    // Navigate to hosts page
    await page.goto('/hosts');
    await expect(page.getByRole('heading', { name: 'Hosts' }).first()).toBeVisible();
  });

  test('renders hosts list page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /hosts|servers/i }).first()).toBeVisible();
  });

  test('creates a new host', async ({ page }) => {
    await page.getByRole('button', { name: /add host/i }).nth(1).click();

    await page.getByLabel('Name *').fill('E2E Test Host');
    await page.getByLabel('Host / IP *').fill('test.example.com');
    await page.getByLabel('Port *').fill('22');
    await page.getByLabel('Username').fill('root');

    await page.getByRole('button', { name: /create host/i }).click();
    await expect(page.locator('[data-testid^="host-row-"][data-host-name="E2E Test Host"]')).toBeVisible({ timeout: 10000 });
  });

  test('edits an existing host', async ({ page, context }) => {
    await context.tracing.start({ screenshots: true, snapshots: true });
    const consoleLogs: string[] = [];
    const networkLogs: string[] = [];
    page.on('console', msg => consoleLogs.push(`[${msg.type()}] ${msg.text()}`));
    page.on('request', req => networkLogs.push(`REQ ${req.method()} ${req.url()}`));
    page.on('response', res => networkLogs.push(`RESP ${res.status()} ${res.url()}`));

    const apiBase = process.env.API_BASE_URL || process.env.BASE_URL?.replace('3000', '8080') || 'http://localhost:8080';
    const origin = process.env.BASE_URL || 'http://localhost:3000';

    // Pre-create host via API to ensure edit target exists
    const token = await apiLogin();
    const createRes = await fetch(`${apiBase}/api/v1/hosts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        Origin: origin,
      },
      body: JSON.stringify({
        name: 'E2E Edit Host',
        address: 'edit.example.com',
        protocol: 'ssh',
        port: 22,
        tags: [],
        description: '',
        group_path: '',
      }),
    });
    expect(createRes.ok).toBeTruthy();
    const created = await createRes.json();
    const hostId = created.host?.id || created.id;

    await page.goto('/hosts');
    await page.reload();
    await expect(page.locator(`[data-testid="host-row-${hostId}"]`)).toBeVisible();

    // Hover row to reveal actions, then click menu
    await page.locator(`[data-testid="host-row-${hostId}"]`).hover();
    await page.locator(`[data-testid="host-actions-menu-${hostId}"]`).click();
    await page.getByRole('menuitem', { name: /edit/i }).click();

    await page.getByLabel('Name *').fill('E2E Edit Host Updated');
    await page.getByRole('button', { name: /update host/i }).click();

    await expect(page.locator('[data-testid^="host-row-"][data-host-name="E2E Edit Host Updated"]')).toBeVisible({ timeout: 10000 });
    await context.tracing.stop({ path: 'trace-edit.zip' });
  });

  test('deletes a host', async ({ page }) => {
    const apiBase = process.env.API_BASE_URL || process.env.BASE_URL?.replace('3000', '8080') || 'http://localhost:8080';
    const origin = process.env.BASE_URL || 'http://localhost:3000';

    // Pre-create host via API
    const token = await apiLogin();
    const createRes = await fetch(`${apiBase}/api/v1/hosts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        Origin: origin,
      },
      body: JSON.stringify({
        name: 'E2E Delete Host',
        address: 'delete.example.com',
        protocol: 'ssh',
        port: 22,
        tags: [],
        description: '',
        group_path: '',
      }),
    });
    expect(createRes.ok).toBeTruthy();
    const created = await createRes.json();
    const hostId = created.host?.id || created.id;

    await page.goto('/hosts');
    await page.reload();
    await expect(page.locator(`[data-testid="host-row-${hostId}"]`)).toBeVisible();

    await page.locator(`[data-testid="host-row-${hostId}"]`).hover();
    await page.locator(`[data-testid="host-actions-menu-${hostId}"]`).click();
    await page.getByRole('menuitem', { name: /delete/i }).click();

    // Confirm delete in dialog
    await page.getByRole('button', { name: /delete/i }).filter({ hasNotText: /cancel/i }).click();

    await expect(page.locator(`[data-testid="host-row-${hostId}"]`)).toHaveCount(0, { timeout: 10000 });
  });
});
