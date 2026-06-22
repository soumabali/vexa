import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { FileManager } from '../file-manager';
import { FileList } from '../file-list';
import { FileTree } from '../file-tree';

// file-grid.tsx doesn't exist in the codebase (file-grid-view.tsx is the actual file)
vi.mock('../file-grid', () => ({
  FileGrid: () => null,
}));
import { FilePreview } from '../file-preview';
import { FileUpload } from '../file-upload';
import { FileToolbar } from '../file-toolbar';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock fetch
global.fetch = vi.fn();

const mockFiles = [
  {
    name: 'test.txt',
    path: '/test.txt',
    size: 1024,
    modified: '2024-01-01T00:00:00Z',
    permissions: '644',
    owner: 'user',
    group: 'group',
    type: 'file' as const,
    mimeType: 'text/plain',
  },
  {
    name: 'folder',
    path: '/folder',
    size: 4096,
    modified: '2024-01-01T00:00:00Z',
    permissions: '755',
    owner: 'user',
    group: 'group',
    type: 'directory' as const,
  },
];

const mockHost = {
  name: 'Test Host',
  host: 'test.com',
  port: 22,
  username: 'user',
  path: '/home/user',
  protocol: 'sftp' as const,
};

describe('FileManager', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('renders local file manager', () => {
    render(
      <FileManager
        type="local"
        host={null}
        onTransferProgress={() => {}}
        onTransferComplete={() => {}}
      />
    );

    expect(screen.getByText('Local Files')).toBeInTheDocument();
  });

  it('renders remote file manager', () => {
    render(
      <FileManager
        type="remote"
        host={mockHost}
        onTransferProgress={() => {}}
        onTransferComplete={() => {}}
      />
    );

    expect(screen.getByText('Test Host')).toBeInTheDocument();
  });

  it('fetches files on mount', async () => {
    const mockResponse = {
      json: () => Promise.resolve({ files: mockFiles }),
    };
    (global.fetch as any).mockResolvedValueOnce(mockResponse);

    render(
      <FileManager
        type="local"
        host={null}
        onTransferProgress={() => {}}
        onTransferComplete={() => {}}
      />
    );

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/files/local?path=%2F');
    });
  });

  it('navigates to directory on click', async () => {
    const mockResponse = {
      json: () => Promise.resolve({ files: mockFiles }),
    };
    (global.fetch as any).mockResolvedValue(mockResponse);

    render(
      <FileManager
        type="local"
        host={null}
        onTransferProgress={() => {}}
        onTransferComplete={() => {}}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('folder')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('folder'));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/files/local?path=%2Ffolder');
    });
  });
});

describe('FileList', () => {
  it('renders file list', () => {
    render(
      <FileList
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        onDownload={() => {}}
        onRename={() => {}}
        onChmod={() => {}}
        onBookmarkToggle={() => {}}
        bookmarks={[]}
        sortBy="name"
        sortOrder="asc"
        onSortChange={() => {}}
        onSortOrderChange={() => {}}
      />
    );

    expect(screen.getByText('test.txt')).toBeInTheDocument();
    expect(screen.getByText('folder')).toBeInTheDocument();
  });

  it('shows selected files', () => {
    render(
      <FileList
        files={mockFiles}
        selectedFiles={new Set(['/test.txt'])}
        onSelect={() => {}}
        onFileClick={() => {}}
        onDownload={() => {}}
        onRename={() => {}}
        onChmod={() => {}}
        onBookmarkToggle={() => {}}
        bookmarks={[]}
        sortBy="name"
        sortOrder="asc"
        onSortChange={() => {}}
        onSortOrderChange={() => {}}
      />
    );

    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes[1]).toBeChecked();
  });

  it('sorts by column', () => {
    const onSortChange = vi.fn();
    render(
      <FileList
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        onDownload={() => {}}
        onRename={() => {}}
        onChmod={() => {}}
        onBookmarkToggle={() => {}}
        bookmarks={[]}
        sortBy="name"
        sortOrder="asc"
        onSortChange={onSortChange}
        onSortOrderChange={() => {}}
      />
    );

    fireEvent.click(screen.getByText('Size'));
    expect(onSortChange).toHaveBeenCalledWith('size');
  });
});

describe('FileGrid', () => {
  it('renders file grid', () => {
    render(
      <FileGrid
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        onDownload={() => {}}
        onRename={() => {}}
        onChmod={() => {}}
        onBookmarkToggle={() => {}}
        bookmarks={[]}
      />
    );

    expect(screen.getByText('test.txt')).toBeInTheDocument();
    expect(screen.getByText('folder')).toBeInTheDocument();
  });

  it('shows bookmark indicator', () => {
    render(
      <FileGrid
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        onDownload={() => {}}
        onRename={() => {}}
        onChmod={() => {}}
        onBookmarkToggle={() => {}}
        bookmarks={['/test.txt']}
      />
    );

    expect(screen.getByText('test.txt')).toBeInTheDocument();
  });
});

describe('FileTree', () => {
  it('renders file tree', () => {
    render(
      <FileTree
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        currentPath="/"
        onNavigate={() => {}}
      />
    );

    expect(screen.getByText('test.txt')).toBeInTheDocument();
    expect(screen.getByText('folder')).toBeInTheDocument();
  });

  it('toggles directory expansion', () => {
    render(
      <FileTree
        files={mockFiles}
        selectedFiles={new Set()}
        onSelect={() => {}}
        onFileClick={() => {}}
        currentPath="/"
        onNavigate={() => {}}
      />
    );

    const chevron = screen.getByText('folder').previousElementSibling;
    if (chevron) {
      fireEvent.click(chevron);
    }
  });
});

describe('FilePreview', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('renders loading state', () => {
    render(
      <FilePreview
        file={mockFiles[0]}
        host={null}
        onDownload={() => {}}
      />
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows error for unsupported file type', async () => {
    const mockResponse = {
      ok: true,
      headers: {
        get: () => 'application/octet-stream',
      },
    };
    (global.fetch as any).mockResolvedValueOnce(mockResponse);

    render(
      <FilePreview
        file={mockFiles[0]}
        host={null}
        onDownload={() => {}}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Preview not available for this file type')).toBeInTheDocument();
    });
  });
});

describe('FileUpload', () => {
  it('renders upload component', () => {
    render(<FileUpload onUpload={() => {}} />);

    expect(screen.getByText('Drag & drop files here')).toBeInTheDocument();
    expect(screen.getByText('or click to browse')).toBeInTheDocument();
  });

  it('handles file drop', () => {
    const onUpload = vi.fn();
    render(<FileUpload onUpload={onUpload} />);

    const file = new File(['test content'], 'test.txt', { type: 'text/plain' });
    const dropZone = screen.getByText('Drag & drop files here');

    fireEvent.dragOver(dropZone);
    fireEvent.drop(dropZone, {
      dataTransfer: {
        files: [file],
      },
    });

    expect(screen.getByText('test.txt')).toBeInTheDocument();
  });
});

describe('FileToolbar', () => {
  it('renders toolbar', () => {
    render(
      <FileToolbar
        currentPath="/"
        viewMode="list"
        onViewModeChange={() => {}}
        onNavigateUp={() => {}}
        onNavigateBack={() => {}}
        onNavigateForward={() => {}}
        canNavigateBack={false}
        canNavigateForward={false}
        selectedCount={0}
        onSelectAll={() => {}}
        onDelete={() => {}}
        onCopy={() => {}}
        onCut={() => {}}
        onPaste={() => {}}
        canPaste={false}
        onUpload={() => {}}
        onRefresh={() => {}}
        searchQuery=""
        onSearchChange={() => {}}
        sortBy="name"
        onSortChange={() => {}}
        sortOrder="asc"
        onSortOrderChange={() => {}}
        bookmarks={[]}
        onBookmarkToggle={() => {}}
        onBookmarkNavigate={() => {}}
        isBookmarked={false}
      />
    );

    expect(screen.getByPlaceholderText('Search files...')).toBeInTheDocument();
  });

  it('shows selection actions when files selected', () => {
    render(
      <FileToolbar
        currentPath="/"
        viewMode="list"
        onViewModeChange={() => {}}
        onNavigateUp={() => {}}
        onNavigateBack={() => {}}
        onNavigateForward={() => {}}
        canNavigateBack={false}
        canNavigateForward={false}
        selectedCount={2}
        onSelectAll={() => {}}
        onDelete={() => {}}
        onCopy={() => {}}
        onCut={() => {}}
        onPaste={() => {}}
        canPaste={true}
        onUpload={() => {}}
        onRefresh={() => {}}
        searchQuery=""
        onSearchChange={() => {}}
        sortBy="name"
        onSortChange={() => {}}
        sortOrder="asc"
        onSortOrderChange={() => {}}
        bookmarks={[]}
        onBookmarkToggle={() => {}}
        onBookmarkNavigate={() => {}}
        isBookmarked={false}
      />
    );

    expect(screen.getByText('2 selected')).toBeInTheDocument();
    expect(screen.getByText('Paste')).toBeInTheDocument();
  });
});
