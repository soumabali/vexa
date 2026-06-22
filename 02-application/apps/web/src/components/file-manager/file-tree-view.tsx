'use client';

import React, { useState } from 'react';
import { FileItem } from '@/types/file-manager';
import { getFileIcon, formatDate } from '@/lib/utils';
import { cn } from '@/lib/utils';
import { ChevronRight, ChevronDown } from 'lucide-react';

interface FileTreeViewProps {
  items: FileItem[];
  selected: Set<string>;
  onItemClick: (item: FileItem, e: React.MouseEvent) => void;
  onItemDoubleClick: (item: FileItem) => void;
}

export function FileTreeView({ items, selected, onItemClick, onItemDoubleClick }: FileTreeViewProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  return (
    <div className="p-2">
      {items.map((item) => (
        <div
          key={item.id}
          className="flex items-center gap-2 py-1 px-2 rounded cursor-pointer hover:bg-accent"
          onClick={(e) => onItemClick(item, e)}
          onDoubleClick={() => onItemDoubleClick(item)}
        >
          {item.type === 'directory' && (
            <button
              className="p-0.5 hover:bg-accent rounded"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(item.id);
              }}
            >
              {expanded.has(item.id) ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </button>
          )}
          <span className="text-sm">{getFileIcon(item)}</span>
          <span className={cn('text-sm', selected.has(item.id) && 'font-medium')}>{item.name}</span>
        </div>
      ))}
    </div>
  );
}