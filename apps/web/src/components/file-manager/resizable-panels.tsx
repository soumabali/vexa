'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { PaneSide, FileItem } from '@/types/file-manager';
import { useFileManager } from '@/hooks/use-file-manager';
import { cn } from '@/lib/utils';

interface ResizablePanelsProps {
  leftPanel: React.ReactNode;
  rightPanel: React.ReactNode;
  bottomPanel?: React.ReactNode;
  initialLeftWidth?: number;
  minLeftWidth?: number;
  maxLeftWidth?: number;
}

export function ResizablePanels({
  leftPanel,
  rightPanel,
  bottomPanel,
  initialLeftWidth = 50,
  minLeftWidth = 20,
  maxLeftWidth = 80,
}: ResizablePanelsProps) {
  const [leftWidth, setLeftWidth] = useState(initialLeftWidth);
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const startXRef = useRef(0);
  const startWidthRef = useRef(leftWidth);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    setIsDragging(true);
    startXRef.current = e.clientX;
    startWidthRef.current = leftWidth;
    e.preventDefault();
  }, [leftWidth]);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging || !containerRef.current) return;
      
      const containerWidth = containerRef.current.offsetWidth;
      const deltaX = e.clientX - startXRef.current;
      const deltaPercent = (deltaX / containerWidth) * 100;
      const newWidth = Math.max(minLeftWidth, Math.min(maxLeftWidth, startWidthRef.current + deltaPercent));
      
      setLeftWidth(newWidth);
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };
  }, [isDragging, minLeftWidth, maxLeftWidth]);

  return (
    <div ref={containerRef} className="flex flex-col h-full w-full">
      <div className="flex flex-1 min-h-0">
        {/* Left Panel */}
        <div
          className="flex flex-col min-w-0 border-r border-border"
          style={{ width: `${leftWidth}%` }}
        >
          {leftPanel}
        </div>

        {/* Resizer */}
        <div
          className={cn(
            'w-1 bg-border hover:bg-primary cursor-col-resize transition-colors z-10 flex items-center justify-center',
            isDragging && 'bg-primary'
          )}
          onMouseDown={handleMouseDown}
        >
          <div className="w-0.5 h-8 bg-muted-foreground/50 rounded-full" />
        </div>

        {/* Right Panel */}
        <div
          className="flex flex-col min-w-0"
          style={{ width: `${100 - leftWidth}%` }}
        >
          {rightPanel}
        </div>
      </div>

      {/* Bottom Panel (Transfer Queue) */}
      {bottomPanel && (
        <>
          <div className="h-px bg-border" />
          <div className="flex-shrink-0">
            {bottomPanel}
          </div>
        </>
      )}
    </div>
  );
}
