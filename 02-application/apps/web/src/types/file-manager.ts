export interface FileItem {
  id: string;
  name: string;
  type: 'file' | 'directory';
  size: number;
  modified: string;
  permissions: string;
  owner: string;
  group: string;
  path: string;
  isSymlink?: boolean;
  target?: string;
  thumbnail?: string;
  mimeType?: string;
}

export interface FileSystemState {
  currentPath: string;
  items: FileItem[];
  selectedItems: Set<string>;
  sortBy: 'name' | 'size' | 'date' | 'permissions';
  sortOrder: 'asc' | 'desc';
  viewMode: 'list' | 'grid' | 'tree';
  filterText: string;
  isLoading: boolean;
  error: string | null;
}

export interface TransferItem {
  id: string;
  fileName: string;
  sourcePath: string;
  destPath: string;
  type: 'upload' | 'download';
  status: 'queued' | 'transferring' | 'paused' | 'completed' | 'error' | 'cancelled';
  progress: number;
  bytesTransferred: number;
  totalBytes: number;
  speed: number; // bytes per second
  eta: number; // seconds
  error?: string;
  startTime?: Date;
  endTime?: Date;
}

export interface TransferQueue {
  items: TransferItem[];
  activeCount: number;
  completedCount: number;
  totalCount: number;
  totalBytesTransferred: number;
  totalBytes: number;
}

export interface ConnectionConfig {
  host: string;
  port: number;
  username: string;
  password?: string;
  privateKey?: string;
  passphrase?: string;
}

export interface Bookmark {
  id: string;
  name: string;
  path: string;
  side: 'local' | 'remote';
}

export interface FileFilter {
  text: string;
  types?: ('file' | 'directory' | 'symlink')[];
  dateFrom?: Date;
  dateTo?: Date;
  sizeFrom?: number;
  sizeTo?: number;
}

export type FileOperation = 'copy' | 'cut' | 'delete' | 'rename' | 'chmod' | 'preview' | 'edit';

export interface ClipboardItem {
  items: FileItem[];
  operation: 'copy' | 'cut';
  sourceSide: 'local' | 'remote';
  sourcePath: string;
}

export interface DragItem {
  items: FileItem[];
  sourceSide: 'local' | 'remote';
  sourcePath: string;
}

export type PaneSide = 'local' | 'remote';
