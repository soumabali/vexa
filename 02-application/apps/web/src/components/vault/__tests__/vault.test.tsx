import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CredentialList } from '../credential-list';
import { CredentialCard } from '../credential-card';
import { CredentialForm } from '../credential-form';
import type { Credential } from '../credential-list';

const mockCredentials: Credential[] = [
  {
    id: '1',
    name: 'Test SSH Key',
    type: 'ssh-key',
    username: 'root',
    host: 'test.example.com',
    port: 22,
    tags: ['test', 'production'],
    folder: 'Test',
    isFavorite: true,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    id: '2',
    name: 'Test Password',
    type: 'password',
    username: 'admin',
    host: 'db.example.com',
    tags: ['database'],
    folder: 'Test',
    isFavorite: false,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
];

describe('CredentialList', () => {
  it('renders credentials correctly', () => {
    render(
      <CredentialList
        credentials={mockCredentials}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onDelete={vi.fn()}
        onFavorite={vi.fn()}
      />
    );

    expect(screen.getByText('Test SSH Key')).toBeInTheDocument();
    expect(screen.getByText('Test Password')).toBeInTheDocument();
  });

  it('filters by search query', () => {
    render(
      <CredentialList
        credentials={mockCredentials}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onDelete={vi.fn()}
        onFavorite={vi.fn()}
      />
    );

    const searchInput = screen.getByPlaceholderText('Search credentials...');
    fireEvent.change(searchInput, { target: { value: 'SSH' } });

    expect(screen.getByText('Test SSH Key')).toBeInTheDocument();
    expect(screen.queryByText('Test Password')).not.toBeInTheDocument();
  });

  it('adds new credential', async () => {
    const onAdd = vi.fn();
    render(
      <CredentialList
        credentials={mockCredentials}
        onAdd={onAdd}
        onUpdate={vi.fn()}
        onDelete={vi.fn()}
        onFavorite={vi.fn()}
      />
    );

    const addButton = screen.getByText('Add Credential');
    fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByText('Add New Credential')).toBeInTheDocument();
    });
  });

  it('toggles favorite', () => {
    const onFavorite = vi.fn();
    render(
      <CredentialList
        credentials={mockCredentials}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onDelete={vi.fn()}
        onFavorite={onFavorite}
      />
    );

    const favoriteButtons = screen.getAllByRole('button', { name: /favorite/i });
    fireEvent.click(favoriteButtons[0]);

    expect(onFavorite).toHaveBeenCalledWith('1');
  });
});

describe('CredentialCard', () => {
  it('renders credential card', () => {
    render(
      <CredentialCard
        credential={mockCredentials[0]}
        onUpdate={vi.fn()}
        onDelete={vi.fn()}
        onFavorite={vi.fn()}
        onClick={vi.fn()}
        viewMode="grid"
      />
    );

    expect(screen.getByText('Test SSH Key')).toBeInTheDocument();
    expect(screen.getByText('root@test.example.com:22')).toBeInTheDocument();
  });

  it('handles delete', () => {
    const onDelete = vi.fn();
    render(
      <CredentialCard
        credential={mockCredentials[0]}
        onUpdate={vi.fn()}
        onDelete={onDelete}
        onFavorite={vi.fn()}
        onClick={vi.fn()}
        viewMode="grid"
      />
    );

    const deleteButton = screen.getByRole('button', { name: /delete/i });
    fireEvent.click(deleteButton);

    expect(onDelete).toHaveBeenCalledWith('1');
  });
});

describe('CredentialForm', () => {
  it('submits form with correct data', () => {
    const onSubmit = vi.fn();
    render(
      <CredentialForm
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    );

    const nameInput = screen.getByLabelText('Name *');
    fireEvent.change(nameInput, { target: { value: 'New Credential' } });

    const submitButton = screen.getByText('Add Credential');
    fireEvent.click(submitButton);

    expect(onSubmit).toHaveBeenCalled();
  });

  it('cancels form', () => {
    const onCancel = vi.fn();
    render(
      <CredentialForm
        onSubmit={vi.fn()}
        onCancel={onCancel}
      />
    );

    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);

    expect(onCancel).toHaveBeenCalled();
  });
});
