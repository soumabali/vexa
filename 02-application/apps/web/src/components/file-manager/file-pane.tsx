'use client';

import React, { useState } from 'react';
import { useFileManager } from '@/hooks/use-file-manager';

interface FilePaneProps {
  side: 'local' | 'remote';
  title?: string;
}

export function FilePane({ side, title }: FilePaneProps) {
  const { activePane, getFilteredItems } = useFileManager();
  const files = getFilteredItems(side);
  
  return (
    <div className="flex flex-col h-full border rounded-lg">
      <div className="p-2 border-b font-semibold">{title || side}</div>
      <div className="flex-1 overflow-auto p-2">
        {files.length === 0 && <p className="text-muted-foreground text-sm">No files</p>}
        {files.map(f => (
          <div key={f.id} className="py-1 text-sm">{f.name}</div>
        ))}
      </div>
    </div>
  );
}
