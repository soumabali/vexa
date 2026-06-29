import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';

// Mock the auth API module used by the sessions page.
const listSessions = vi.fn();
const revokeSession = vi.fn();
vi.mock('@/lib/api/auth', () => ({
  authApi: {
    listSessions: (...args: unknown[]) => listSessions(...args),
    revokeSession: (...args: unknown[]) => revokeSession(...args),
  },
}));

// Provide a no-op DashboardLayout so the page renders in isolation.
vi.mock('@/components/layouts/DashboardLayout', () => ({
  DashboardLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

import SessionsPage from '../sessions/page';

const sampleSessions = [
  {
    session_id: 'sess-current',
    ip_address: '203.0.113.10',
    user_agent: 'Mozilla/5.0 (Macintosh) Chrome/120 Safari/537',
    created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
    last_active_at: new Date().toISOString(),
    is_current: true,
  },
  {
    session_id: 'sess-other',
    ip_address: '198.51.100.7',
    user_agent: 'Mozilla/5.0 (Windows NT 10.0) Firefox/118',
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    last_active_at: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
    is_current: false,
  },
];

beforeEach(() => {
  listSessions.mockReset();
  revokeSession.mockReset();
});

describe('SessionsPage', () => {
  it('renders active sessions list', async () => {
    listSessions.mockResolvedValueOnce(sampleSessions);
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getAllByTestId('session-row')).toHaveLength(2);
    });
    expect(screen.getByText('Chrome on macOS')).toBeInTheDocument();
    expect(screen.getByText('Firefox on Windows')).toBeInTheDocument();
  });

  it('marks the current session with a badge', async () => {
    listSessions.mockResolvedValueOnce(sampleSessions);
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getByText('This device')).toBeInTheDocument();
    });
  });

  it('does not show a revoke button for the current session', async () => {
    listSessions.mockResolvedValueOnce(sampleSessions);
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getAllByTestId('session-row')).toHaveLength(2);
    });
    const buttons = screen.getAllByTestId('revoke-button');
    expect(buttons).toHaveLength(1);
    expect(buttons[0]).toHaveAttribute('aria-label', 'Revoke Firefox session');
  });

  it('opens confirmation dialog and revokes a non-current session', async () => {
    listSessions.mockResolvedValueOnce(sampleSessions);
    revokeSession.mockResolvedValueOnce({ message: 'session revoked' });
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getAllByTestId('session-row')).toHaveLength(2);
    });

    fireEvent.click(screen.getByTestId('revoke-button'));
    await waitFor(() => {
      expect(screen.getByText('Revoke this session?')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('confirm-revoke'));

    await waitFor(() => {
      expect(revokeSession).toHaveBeenCalledWith('sess-other');
    });
  });

  it('shows empty state when no sessions exist', async () => {
    listSessions.mockResolvedValueOnce([]);
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getByText('No active sessions.')).toBeInTheDocument();
    });
  });

  it('renders an error state when listing fails', async () => {
    listSessions.mockRejectedValueOnce(new Error('boom'));
    render(<SessionsPage />);

    await waitFor(() => {
      expect(screen.getByText('boom')).toBeInTheDocument();
    });
  });
});