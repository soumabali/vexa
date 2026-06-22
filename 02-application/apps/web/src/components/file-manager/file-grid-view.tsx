'use client';

import React from 'react';
interface FileItem {
  id: string;
  name: string;
  path: string;
  size: number;
  modified: string;
  permissions: string;
  owner: string;
  group: string;
  type: 'file' | 'directory';
  isSymlink?: boolean;
  target?: string;
  mimeType?: string;
}
import { getFileIcon } from '@/lib/utils';
import { cn } from '@/lib/utils';

interface FileGridViewProps {
  items: FileItem[];
  selected: Set<string>;
  onItemClick: (item: FileItem, e: React.MouseEvent) => void;
  onItemDoubleClick: (item: FileItem) => void;
}

export function FileGridView({ items, selected, onItemClick, onItemDoubleClick }: FileGridViewProps) {
  return (
    <div className="grid grid-cols-4 gap-2 p-2">
      {items.map((item) => (
        <div
          key={item.id}
          className={cn(
            'flex flex-col items-center p-3 rounded-lg cursor-pointer border transition-colors',
            selected.has(item.id) ? 'bg-accent border-primary' : 'hover:bg-accent border-transparent'
          )}
          onClick={(e) => onItemClick(item, e)}
          onDoubleClick={() => onItemDoubleClick(item)}
        >
          <span className="text-2xl mb-1">{getFileIcon(item)}</span>
          <span className="text-xs text-center truncate w-full">{item.name}</span>
        </div>
      ))}
    </div>
  );
}