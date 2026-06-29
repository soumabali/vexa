import { render, screen } from '@testing-library/react';
import AdminDashboard from '../page';
import UserManagement from '../users/page';
import SessionManagement from '../sessions/page';
import HostManagement from '../hosts/page';

import { beforeEach, describe, expect, it, vi } from 'vitest';

// Mock WebSocket
global.WebSocket = vi.fn().mockImplementation(() => ({
  onopen: null,
  onclose: null,
  onmessage: null,
  close: vi.fn(),
  send: vi.fn(),
})) as unknown as typeof WebSocket;

// Mock fetch
global.fetch = vi.fn().mockImplementation((url) => {
  if (url === '/api/admin/stats') {
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          totalUsers: 100,
          activeUsers: 45,
          activeSessions: 23,
          totalHosts: 50,
          onlineHosts: 42,
          totalBandwidth: '1.2 TB',
          errorsLast24h: 5,
          avgLatency: '24 ms',
        }),
    });
  }
  if (url === '/api/admin/users') {
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          users: [
            {
              id: '1',
              name: 'Test User',
              email: 'test@example.com',
              role: 'user',
              status: 'active',
              mfaEnabled: true,
              lastLogin: '2024-01-01T00:00:00Z',
              createdAt: '2024-01-01T00:00:00Z',
            },
          ],
        }),
    });
  }
  if (url === '/api/admin/sessions') {
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          sessions: [
            {
              id: '1',
              user: 'Test User',
              host: 'prod-web-01',
              protocol: 'ssh',
              status: 'active',
              startedAt: '2024-01-01T00:00:00Z',
              duration: '2h 30m',
              clientIP: '192.168.1.1',
            },
          ],
        }),
    });
  }
  if (url === '/api/admin/hosts') {
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          hosts: [
            {
              id: '1',
              name: 'prod-web-01',
              hostname: '192.168.1.10',
              port: 22,
              protocol: 'ssh',
              status: 'online',
              os: 'Ubuntu 22.04',
              lastSeen: '2024-01-01T00:00:00Z',
              healthStatus: 'healthy',
              latency: '12 ms',
            },
          ],
        }),
    });
  }
  return Promise.resolve({ json: () => Promise.resolve({}) });
});

describe('Admin Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders dashboard with stats cards', async () => {
    render(<AdminDashboard />);
    expect(await screen.findByText('Admin Dashboard')).toBeInTheDocument();
    expect(await screen.findByText('Total Users')).toBeInTheDocument();
    expect(await screen.findByText('100')).toBeInTheDocument();
  });

  it('renders user management page', async () => {
    render(<UserManagement />);
    expect(await screen.findByText('User Management')).toBeInTheDocument();
    expect(await screen.findByText('Test User')).toBeInTheDocument();
    expect(await screen.findByText('test@example.com')).toBeInTheDocument();
  });

  it('renders session management page', async () => {
    render(<SessionManagement />);
    expect(await screen.findByText('Session Management')).toBeInTheDocument();
    expect(await screen.findByText('Test User')).toBeInTheDocument();
    expect(await screen.findByText('prod-web-01')).toBeInTheDocument();
  });

  it('renders host management page', async () => {
    render(<HostManagement />);
    expect(await screen.findByText('Host Management')).toBeInTheDocument();
    expect(await screen.findByText('prod-web-01')).toBeInTheDocument();
    expect(await screen.findByText('Ubuntu 22.04')).toBeInTheDocument();
  });

  it('displays WebSocket connection status', async () => {
    render(<AdminDashboard />);
    expect(await screen.findByText(/Live|Disconnected/)).toBeInTheDocument();
  });
});