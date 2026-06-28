'use client';

import React from 'react';
import { FileItem } from '@/types/file-manager';
import { cn, formatFileSize, formatDate, getFileIcon } from '@/lib/utils';
import {
  Folder,
  File,
  FileText,
  Image,
  Code,
  Archive,
  Link,
  ChevronUp,
  ChevronDown,
} from 'lucide-react';

interface FileListViewProps {
  items: FileItem[];
  selected: Set<string>;
  sortBy: 'name' | 'size' | 'date' | 'permissions';
  sortOrder: 'asc' | 'desc';
  onSort: (by: 'name' | 'size' | 'date' | 'permissions') => void;
  onItemClick: (item: FileItem, e: React.MouseEvent) => void;
  onItemDoubleClick: (item: FileItem) => void;
}

const fileIconMap: Record<string, React.ReactNode> = {
  folder: <Folder className="h-5 w-5 text-blue-500" />,
  file: <File className="h-5 w-5 text-gray-500" />,
  'file-text': <FileText className="h-5 w-5 text-yellow-500" />,
  image: <Image className="h-5 w-5 text-green-500" />,
  code: <Code className="h-5 w-5 text-purple-500" />,
  archive: <Archive className="h-5 w-5 text-orange-500" />,
  link: <Link className="h-5 w-5 text-cyan-500" />,
};

type SortBy = FileListViewProps['sortBy'];
type SortOrder = FileListViewProps['sortOrder'];

function SortIcon({ column, sortBy, sortOrder }: { column: SortBy; sortBy: SortBy; sortOrder: SortOrder }) {
  if (sortBy !== column) return null;
  return sortOrder === 'asc' ? (
    <ChevronUp className="h-3 w-3 ml-1" />
  ) : (
    <ChevronDown className="h-3 w-3 ml-1" />
  );
}

export function FileListView({
  items,
  selected,
  sortBy,
  sortOrder,
  onSort,
  onItemClick,
  onItemDoubleClick,
}: FileListViewProps) {
  return (
    <div className="w-full">
      {/* Header */}
      <div className="flex items-center px-3 py-2 border-b bg-muted/50 text-sm font-medium text-muted-foreground">
        <div
          className="flex items-center flex-1 cursor-pointer hover:text-foreground transition-colors"
          onClick={() => onSort('name')}
        >
          Name
          <SortIcon column="name" sortBy={sortBy} sortOrder={sortOrder} />
        </div>
        <div
          className="flex items-center w-24 cursor-pointer hover:text-foreground transition-colors"
          onClick={() => onSort('size')}
        >
          Size
          <SortIcon column="size" sortBy={sortBy} sortOrder={sortOrder} />
        </div>
        <div
          className="flex items-center w-36 cursor-pointer hover:text-foreground transition-colors"
          onClick={() => onSort('date')}
        >
          Modified
          <SortIcon column="date" sortBy={sortBy} sortOrder={sortOrder} />
        </div>
        <div
          className="flex items-center w-20 cursor-pointer hover:text-foreground transition-colors"
          onClick={() => onSort('permissions')}
        >
          Perms
          <SortIcon column="permissions" sortBy={sortBy} sortOrder={sortOrder} />
        </div>
      </div>

      {/* Items */}
      <div className="divide-y">
        {items.map((item) => (
          <div
            key={item.id}
            className={cn(
              'flex items-center px-3 py-2 text-sm cursor-pointer hover:bg-accent/50 transition-colors',
              selected.has(item.id) && 'bg-accent'
            )}
            onClick={(e) => onItemClick(item, e)}
            onDoubleClick={() => onItemDoubleClick(item)}
          >
            <div className="flex items-center flex-1 min-w-0">
              <span className="mr-2 flex-shrink-0">
                {fileIconMap[getFileIcon(item)] || fileIconMap.file}
              </span>
              <span className="truncate">
                {item.name}
                {item.isSymlink && (
                  <span className="text-muted-foreground ml-1">→ {item.target}</span>
                )}
              </span>
            </div>
            <div className="w-24 text-muted-foreground">
              {item.type === 'directory' ? '--' : formatFileSize(item.size)}
            </div>
            <div className="w-36 text-muted-foreground">
              {formatDate(item.modified)}
            </div>
            <div className="w-20 font-mono text-xs text-muted-foreground">
              {item.permissions}
            </div>
          </div>
        ))}

        {items.length === 0 && (
          <div className="flex items-center justify-center py-8 text-muted-foreground">
            No files found
          </div>
        )}
      </div>
    </div>
  );
}
