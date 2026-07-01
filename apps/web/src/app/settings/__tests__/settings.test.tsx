import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import type { ReactNode } from 'react';
import SettingsPage from '../page';
import SecuritySettingsPage from '../security/page';
import APIKeysPage from '../api-keys/page';
import AppearanceSettingsPage from '../appearance/page';

// @/components/ui/radio-group doesn't exist in the codebase yet
vi.mock('@/components/ui/radio-group', () => ({
  RadioGroup: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  RadioGroupItem: () => <div />,
}));

// Mock appearance page separately since it imports radio-group
vi.mock('../appearance/page', () => ({
  default: () => <div>Appearance</div>,
}));

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
  usePathname: () => '/settings/security',
}));

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

// Mock the auth API surface used by security page
vi.mock('@/lib/api/auth', () => ({
  authApi: {
    getUserProfile: vi.fn().mockResolvedValue({
      id: 'u1',
      email: 'user@example.com',
      role: 'user',
      mfa_enabled: true,
      mfa_required: false,
    }),
    regenerateBackupCodes: vi.fn(),
  },
}));

// Mock webauthn lib (security page imports it for credential listing)
vi.mock('@/lib/webauthn', () => ({
  isWebAuthnSupported: vi.fn().mockResolvedValue(false),
  listCredentials: vi.fn().mockResolvedValue({ credentials: [] }),
  deleteCredential: vi.fn(),
}));

describe('SettingsPage', () => {
  it('renders settings sections', () => {
    render(<SettingsPage />);

    expect(screen.getAllByText('Settings')[0]).toBeInTheDocument();
    expect(screen.getByText('Profile')).toBeInTheDocument();
    expect(screen.getByText('Security')).toBeInTheDocument();
    expect(screen.getByText('API Keys')).toBeInTheDocument();
    expect(screen.getByText('Appearance')).toBeInTheDocument();
    expect(screen.getByText((c) => c.includes('Notifications'))).toBeInTheDocument();
  });

  it('links to correct pages', () => {
    render(<SettingsPage />);

    expect(screen.getByText('Profile').closest('a')).toHaveAttribute('href', '/settings/profile');
    expect(screen.getByText('Security').closest('a')).toHaveAttribute('href', '/settings/security');
    expect(screen.getByText('API Keys').closest('a')).toHaveAttribute('href', '/settings/api-keys');
  });
});

describe('SecuritySettingsPage', () => {
  it('renders security settings', () => {
    render(<SecuritySettingsPage />);

    expect(screen.getByText('Security Settings')).toBeInTheDocument();
    expect(screen.getByText((c) => c.includes('Multi-Factor') && c.includes('Authentication'))).toBeInTheDocument();
    expect(screen.getByText((c) => c.includes('Passkeys') || c.includes('WebAuthn'))).toBeInTheDocument();
    expect(screen.getByText((c) => c.includes('Security') && c.includes('Audit'))).toBeInTheDocument();
  });

  it('toggles 2FA', async () => {
    render(<SecuritySettingsPage />);

    const toggle = await screen.findByRole('switch', { name: /two-factor authentication status/i });
    fireEvent.click(toggle);

    expect(screen.getByText((c) => c.includes('Multi-Factor') && c.includes('Authentication'))).toBeInTheDocument();
    expect(screen.getAllByText((c) => c.includes('Backup') && c.includes('codes'))[0]).toBeInTheDocument();
  });
});

describe('APIKeysPage', () => {
  it('renders API keys', () => {
    render(<APIKeysPage />);

    expect(screen.getByText('API Keys')).toBeInTheDocument();
    expect(screen.getByText('Production API Key')).toBeInTheDocument();
    expect(screen.getByText('CI/CD Key')).toBeInTheDocument();
  });

  it('shows create dialog', () => {
    render(<APIKeysPage />);

    const createButton = screen.getAllByText((c) => c.includes('Create') && c.includes('Key'))[0];
    fireEvent.click(createButton);

    expect(screen.getAllByText((c) => c.includes('Create') && c.includes('Key'))[0]).toBeInTheDocument();
  });

  it('copies API key', () => {
    const mockClipboard = {
      writeText: vi.fn(),
    };
    Object.assign(navigator, { clipboard: mockClipboard });

    render(<APIKeysPage />);

    const copyButtons = screen.getAllByRole('button', { name: /copy/i });
    fireEvent.click(copyButtons[0]);

    expect(mockClipboard.writeText).toHaveBeenCalled();
  });
});

describe('AppearanceSettingsPage (mocked)', () => {
  it('renders mocked appearance page', () => {
    render(<AppearanceSettingsPage />);
    expect(screen.getByText('Appearance')).toBeInTheDocument();
  });
});

import { waitFor } from '@testing-library/react';
import { authApi } from '@/lib/api/auth';

describe('SecuritySettingsPage - backup codes dialog', () => {
  beforeEach(() => {
    vi.mocked(authApi.regenerateBackupCodes).mockReset();
  });

  it('opens the request dialog when regenerate is clicked', async () => {
    render(<SecuritySettingsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('open-backup-codes')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('open-backup-codes'));
    expect(
      screen.getByRole('heading', { name: 'Regenerate backup codes' })
    ).toBeInTheDocument();
    expect(screen.getByTestId('backup-codes-totp')).toBeInTheDocument();
  });

  it('shows one-time codes after a successful TOTP submit', async () => {
    vi.mocked(authApi.regenerateBackupCodes).mockResolvedValueOnce({
      backup_codes: ['AAAA-1111', 'BBBB-2222', 'CCCC-3333'],
      message: 'codes generated',
    });

    render(<SecuritySettingsPage />);
    await waitFor(() => {
      expect(screen.getByTestId('open-backup-codes')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('open-backup-codes'));
    fireEvent.change(screen.getByTestId('backup-codes-totp'), {
      target: { value: '123456' },
    });
    fireEvent.click(screen.getByTestId('submit-backup-codes'));

    await waitFor(() => {
      expect(screen.getByTestId('backup-codes-list')).toBeInTheDocument();
    });
    expect(screen.getByText('AAAA-1111')).toBeInTheDocument();
    expect(screen.getByText('BBBB-2222')).toBeInTheDocument();
    expect(screen.getByText('CCCC-3333')).toBeInTheDocument();
  });

  it('clears codes from state when user closes the dialog', async () => {
    vi.mocked(authApi.regenerateBackupCodes).mockResolvedValueOnce({
      backup_codes: ['XXXX-9999'],
      message: 'codes generated',
    });

    render(<SecuritySettingsPage />);
    await waitFor(() => {
      expect(screen.getByTestId('open-backup-codes')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('open-backup-codes'));
    fireEvent.change(screen.getByTestId('backup-codes-totp'), {
      target: { value: '123456' },
    });
    fireEvent.click(screen.getByTestId('submit-backup-codes'));

    await waitFor(() => {
      expect(screen.getByText('XXXX-9999')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('confirm-backup-codes'));

    await waitFor(() => {
      expect(screen.queryByText('XXXX-9999')).not.toBeInTheDocument();
    });
  });
});
