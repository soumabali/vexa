'use client';

import React, { useState, useCallback } from 'react';
import { Copy, ClipboardPaste, Search, ZoomIn, ZoomOut, Bell, BellOff, BellRing, Settings } from 'lucide-react';

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

  const BellIcon = bellMode === 'off' ? BellOff : bellMode === 'audible' ? BellRing : Bell;

  return (
    <div className={`terminal-toolbar flex items-center gap-1 px-2 py-1 bg-[#2d2d2d] border-b border-[#3c3c3c] ${className || ''}`}>
      <button
        onClick={onCopy}
        className="p-1.5 rounded hover:bg-[#505050] text-gray-300 hover:text-white transition-colors"
        title="Copy selection"
      >
        <Copy className="w-4 h-4" />
      </button>

      <button
        onClick={onPaste}
        className="p-1.5 rounded hover:bg-[#505050] text-gray-300 hover:text-white transition-colors"
        title="Paste"
      >
        <ClipboardPaste className="w-4 h-4" />
      </button>

      <div className="w-px h-4 bg-[#3c3c3c] mx-1" />

      <button
        onClick={() => setSearchVisible(!searchVisible)}
        className={`p-1.5 rounded hover:bg-[#505050] transition-colors ${
          searchVisible ? 'text-blue-400 bg-[#505050]' : 'text-gray-300 hover:text-white'
        }`}
        title="Search"
      >
        <Search className="w-4 h-4" />
      </button>

      {searchVisible && (
        <div className="flex items-center gap-1 bg-[#1e1e1e] rounded px-2 py-0.5">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search..."
            autoFocus
            className="bg-transparent text-white text-sm outline-none w-32 placeholder-gray-500"
          />
          <button
            onClick={() => setCaseSensitive(!caseSensitive)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              caseSensitive ? 'bg-blue-500 text-white' : 'text-gray-400 hover:bg-[#505050]'
            }`}
            title="Case sensitive"
          >
            Aa
          </button>
          <button
            onClick={() => setWholeWord(!wholeWord)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              wholeWord ? 'bg-blue-500 text-white' : 'text-gray-400 hover:bg-[#505050]'
            }`}
            title="Whole word"
          >
            \b
          </button>
          <button
            onClick={() => setRegexMode(!regexMode)}
            className={`text-xs px-1.5 py-0.5 rounded ${
              regexMode ? 'bg-blue-500 text-white' : 'text-gray-400 hover:bg-[#505050]'
            }`}
            title="Regex"
          >
            .*
          </button>
        </div>
      )}

      <div className="w-px h-4 bg-[#3c3c3c] mx-1" />

      <button
        onClick={onZoomOut}
        className="p-1.5 rounded hover:bg-[#505050] text-gray-300 hover:text-white transition-colors"
        title="Zoom out"
      >
        <ZoomOut className="w-4 h-4" />
      </button>

      <span className="text-xs text-gray-400 min-w-[3ch] text-center">{fontSize}</span>

      <button
        onClick={onZoomIn}
        className="p-1.5 rounded hover:bg-[#505050] text-gray-300 hover:text-white transition-colors"
        title="Zoom in"
      >
        <ZoomIn className="w-4 h-4" />
      </button>

      <div className="w-px h-4 bg-[#3c3c3c] mx-1" />

      <button
        onClick={onToggleBell}
        className={`p-1.5 rounded hover:bg-[#505050] transition-colors ${
          bellMode === 'off' ? 'text-gray-500' : 'text-gray-300 hover:text-white'
        }`}
        title={`Bell: ${bellMode}`}
      >
        <BellIcon className="w-4 h-4" />
      </button>

      <button
        onClick={onSettings}
        className="p-1.5 rounded hover:bg-[#505050] text-gray-300 hover:text-white transition-colors"
        title="Settings"
      >
        <Settings className="w-4 h-4" />
      </button>
    </div>
  );
}
