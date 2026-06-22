export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

export function formatDate(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

export function formatSpeed(bytesPerSecond: number): string {
  return `${formatFileSize(bytesPerSecond)}/s`;
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.ceil(seconds)}s`;
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.ceil(seconds % 60);
    return `${mins}m ${secs}s`;
  }
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${mins}m`;
}

export function getFileIcon(file: { name: string; type: string }): string {
  if (file.type === 'directory') return 'folder';
  if (file.type === 'symlink') return 'link';
  
  const ext = file.name.split('.').pop()?.toLowerCase() || '';
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'];
  const codeExts = ['js', 'ts', 'jsx', 'tsx', 'html', 'css', 'py', 'java', 'cpp', 'c', 'go', 'rs', 'php'];
  const docExts = ['pdf', 'doc', 'docx', 'txt', 'md', 'rtf'];
  const archiveExts = ['zip', 'tar', 'gz', 'rar', '7z'];
  
  if (imageExts.includes(ext)) return 'image';
  if (codeExts.includes(ext)) return 'code';
  if (docExts.includes(ext)) return 'file-text';
  if (archiveExts.includes(ext)) return 'archive';
  return 'file';
}

export function isTextFile(fileName: string): boolean {
  const textExts = ['txt', 'md', 'js', 'ts', 'jsx', 'tsx', 'html', 'css', 'json', 'xml', 'yaml', 'yml', 'py', 'java', 'cpp', 'c', 'go', 'rs', 'php', 'rb', 'sh', 'bash', 'conf', 'ini', 'log'];
  const ext = fileName.split('.').pop()?.toLowerCase() || '';
  return textExts.includes(ext);
}

export function isImageFile(fileName: string): boolean {
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'];
  const ext = fileName.split('.').pop()?.toLowerCase() || '';
  return imageExts.includes(ext);
}

export function getParentPath(path: string): string {
  const parts = path.split('/').filter(Boolean);
  parts.pop();
  return '/' + parts.join('/');
}

export function joinPaths(...parts: string[]): string {
  return parts
    .map(p => p.replace(/^\/+|\/+$/g, ''))
    .filter(Boolean)
    .join('/');
}

export function sortFiles(
  files: any[],
  sortBy: 'name' | 'size' | 'date' | 'permissions',
  sortOrder: 'asc' | 'desc'
): any[] {
  const sorted = [...files].sort((a, b) => {
    // Directories first
    if (a.type === 'directory' && b.type !== 'directory') return -1;
    if (b.type === 'directory' && a.type !== 'directory') return 1;
    
    let comparison = 0;
    switch (sortBy) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'size':
        comparison = (a.size || 0) - (b.size || 0);
        break;
      case 'date':
        comparison = new Date(a.modified).getTime() - new Date(b.modified).getTime();
        break;
      case 'permissions':
        comparison = (a.permissions || '').localeCompare(b.permissions || '');
        break;
    }
    return comparison;
  });
  
  return sortOrder === 'desc' ? sorted.reverse() : sorted;
}

export function filterFiles(files: any[], filterText: string): any[] {
  if (!filterText) return files;
  const lowerFilter = filterText.toLowerCase();
  return files.filter(file =>
    file.name.toLowerCase().includes(lowerFilter)
  );
}
