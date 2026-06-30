'use client';

import React, { useState, useCallback } from 'react';
import { MaterialIcon } from '@/components/ui/material-icon';

export interface TerminalToolbarProps {
  onCopy?: () => void;
  onPaste?: () => void;
  onSearch?: (query: string) => void;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  onToggleBell?: () => void;
  onSettings?: () => void;
  bellMode?: 'visual' | 'audible' | 'off';
  fontSize?: number;
  className?: string;
}

export default function TerminalToolbar({
  onCopy,
  onPaste,
  onSearch,
  onZoomIn,
  onZoomOut,
  onToggleBell,
  onSettings,
  bellMode = 'off',
  fontSize = 14,
  className,
}: TerminalToolbarProps) {
  const [searchVisible, setSearchVisible] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [wholeWord, setWholeWord] = useState(false);
  const [regexMode, setRegexMode] = useState(false);

  const handleSearch = useCallback(() => {
    if (searchQuery.trim()) {
      onSearch?.(searchQuery);
    }
  }, [searchQuery, onSearch]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleSearch();
      }
      if (e.key === 'Escape') {
        setSearchVisible(false);
        setSearchQuery('');
      }
    },
    [handleSearch]
  );

  const BellIcon = bellMode === 'off' ? 'notifications_off' : bellMode === 'audible' ? 'notifications_active' : 'notifications';

  return (
    <div className={`terminal-toolbar flex items-center gap-1 px-2 py-1 bg-surface-container border-b border-outline-variant ${className || ''}`}>
      <button
        onClick={onCopy}
        className="p-1.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors"
        title="Copy selection"
      >
        <MaterialIcon name="content_copy" size="sm" />
      </button>

      <button
        onClick={onPaste}
        className="p-1.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors"
        title="Paste"
      >
        <MaterialIcon name="content_paste" size="sm" />
      </button>

      <div className="w-px h-4 bg-outline-variant mx-1" />

      <button
        onClick={() => setSearchVisible(!searchVisible)}
        className={`p-1.5 rounded transition-colors ${
          searchVisible ? 'text-primary bg-surface-variant' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-variant'
        }`}
        title="Search"
      >
        <MaterialIcon name="search" size="sm" />
      </button>

      {searchVisible && (
        <div className="flex items-center gap-1 bg-surface-container-low rounded px-2 py-0.5">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search..."
            autoFocus
            className="bg-transparent text-on-surface text-sm outline-none w-32 placeholder:text-on-surface-variant"
          />
          <button
            onClick={() => setCaseSensitive(!caseSensitive)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              caseSensitive ? 'bg-primary text-on-primary' : 'text-on-surface-variant hover:bg-surface-variant'
            }`}
            title="Case sensitive"
          >
            Aa
          </button>
          <button
            onClick={() => setWholeWord(!wholeWord)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              wholeWord ? 'bg-primary text-on-primary' : 'text-on-surface-variant hover:bg-surface-variant'
            }`}
            title="Whole word"
          >
            \b
          </button>
          <button
            onClick={() => setRegexMode(!regexMode)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              regexMode ? 'bg-primary text-on-primary' : 'text-on-surface-variant hover:bg-surface-variant'
            }`}
            title="Regex"
          >
            .*
          </button>
        </div>
      )}

      <div className="w-px h-4 bg-outline-variant mx-1" />

      <button
        onClick={onZoomOut}
        className="p-1.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors"
        title="Zoom out"
      >
        <MaterialIcon name="zoom_out" size="sm" />
      </button>

      <span className="text-xs text-on-surface-variant min-w-[3ch] text-center">{fontSize}</span>

      <button
        onClick={onZoomIn}
        className="p-1.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors"
        title="Zoom in"
      >
        <MaterialIcon name="zoom_in" size="sm" />
      </button>

      <div className="w-px h-4 bg-outline-variant mx-1" />

      <button
        onClick={onToggleBell}
        className={`p-1.5 rounded transition-colors ${
          bellMode === 'off' ? 'text-on-surface-variant' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-variant'
        }`}
        title={`Bell: ${bellMode}`}
      >
        <MaterialIcon name={BellIcon} size="sm" />
      </button>

      <button
        onClick={onSettings}
        className="p-1.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors"
        title="Settings"
      >
        <MaterialIcon name="settings" size="sm" />
      </button>
    </div>
  );
}
