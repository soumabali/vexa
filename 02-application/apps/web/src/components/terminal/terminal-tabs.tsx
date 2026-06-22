'use client';

import React, { useState, useCallback, useRef } from 'react';
import { X, Plus, GripVertical } from 'lucide-react';
import Terminal, { TerminalProps, TerminalStatus } from './terminal';

export interface Tab {
  id: string;
  title: string;
  status: TerminalStatus;
  wsUrl: string;
  active: boolean;
}

export interface TerminalTabsProps {
  initialTabs?: Tab[];
  defaultWsUrl: string;
  onTabChange?: (tabs: Tab[]) => void;
  onActiveTabChange?: (tabId: string | null) => void;
}

let tabCounter = 0;

function generateTabId(): string {
  tabCounter += 1;
  return `tab-${Date.now()}-${tabCounter}`;
}

export default function TerminalTabs({
  initialTabs = [],
  defaultWsUrl,
  onTabChange,
  onActiveTabChange,
}: TerminalTabsProps) {
  const [tabs, setTabs] = useState<Tab[]>(
    initialTabs.length > 0
      ? initialTabs
      : [
          {
            id: generateTabId(),
            title: 'Local',
            status: 'connecting',
            wsUrl: defaultWsUrl,
            active: true,
          },
        ]
  );
  const [draggedTab, setDraggedTab] = useState<string | null>(null);
  const dragOverTab = useRef<string | null>(null);

  const updateTabs = useCallback(
    (updater: (prev: Tab[]) => Tab[]) => {
      setTabs((prev) => {
        const next = updater(prev);
        onTabChange?.(next);
        return next;
      });
    },
    [onTabChange]
  );

  const addTab = useCallback(() => {
    const newTab: Tab = {
      id: generateTabId(),
      title: `Session ${tabs.length + 1}`,
      status: 'connecting',
      wsUrl: defaultWsUrl,
      active: true,
    };
    updateTabs((prev) => [
      ...prev.map((t) => ({ ...t, active: false })),
      newTab,
    ]);
    onActiveTabChange?.(newTab.id);
  }, [tabs.length, defaultWsUrl, updateTabs, onActiveTabChange]);

  const closeTab = useCallback(
    (tabId: string) => {
      updateTabs((prev) => {
        const index = prev.findIndex((t) => t.id === tabId);
        const filtered = prev.filter((t) => t.id !== tabId);

        if (filtered.length === 0) {
          return [
            {
              id: generateTabId(),
              title: 'Local',
              status: 'connecting',
              wsUrl: defaultWsUrl,
              active: true,
            },
          ];
        }

        if (prev[index]?.active) {
          const newActive = filtered[Math.min(index, filtered.length - 1)];
          onActiveTabChange?.(newActive.id);
          return filtered.map((t) => (t.id === newActive.id ? { ...t, active: true } : t));
        }

        return filtered;
      });
    },
    [defaultWsUrl, updateTabs, onActiveTabChange]
  );

  const activateTab = useCallback(
    (tabId: string) => {
      updateTabs((prev) =>
        prev.map((t) => ({ ...t, active: t.id === tabId }))
      );
      onActiveTabChange?.(tabId);
    },
    [updateTabs, onActiveTabChange]
  );

  const handleDragStart = useCallback((e: React.DragEvent, tabId: string) => {
    setDraggedTab(tabId);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', tabId);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent, tabId: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    dragOverTab.current = tabId;
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const droppedId = e.dataTransfer.getData('text/plain');
      const targetId = dragOverTab.current;

      if (!droppedId || !targetId || droppedId === targetId) {
        setDraggedTab(null);
        return;
      }

      updateTabs((prev) => {
        const fromIndex = prev.findIndex((t) => t.id === droppedId);
        const toIndex = prev.findIndex((t) => t.id === targetId);

        if (fromIndex === -1 || toIndex === -1) return prev;

        const newTabs = [...prev];
        const [moved] = newTabs.splice(fromIndex, 1);
        newTabs.splice(toIndex, 0, moved);
        return newTabs;
      });

      setDraggedTab(null);
      dragOverTab.current = null;
    },
    [updateTabs]
  );

  const handleTitleChange = useCallback(
    (tabId: string, title: string) => {
      updateTabs((prev) =>
        prev.map((t) => (t.id === tabId ? { ...t, title: title || t.title } : t))
      );
    },
    [updateTabs]
  );

  const handleStatusChange = useCallback(
    (tabId: string, status: TerminalStatus) => {
      updateTabs((prev) =>
        prev.map((t) => (t.id === tabId ? { ...t, status } : t))
      );
    },
    [updateTabs]
  );

  const activeTab = tabs.find((t) => t.active);

  return (
    <div className="terminal-tabs flex flex-col h-full bg-[#1e1e1e] rounded-lg overflow-hidden">
      {/* Tab bar */}
      <div className="tab-bar flex items-center bg-[#2d2d2d] border-b border-[#3c3c3c] overflow-x-auto">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            draggable
            onDragStart={(e) => handleDragStart(e, tab.id)}
            onDragOver={(e) => handleDragOver(e, tab.id)}
            onDrop={handleDrop}
            onClick={() => activateTab(tab.id)}
            className={`tab flex items-center gap-1.5 px-3 py-2 min-w-[120px] max-w-[200px] cursor-pointer select-none text-sm transition-colors ${
              tab.active
                ? 'bg-[#1e1e1e] text-white'
                : 'bg-[#2d2d2d] text-gray-400 hover:bg-[#3c3c3c] hover:text-gray-200'
            } ${draggedTab === tab.id ? 'opacity-50' : ''}`}
          >
            <GripVertical className="w-3 h-3 text-gray-500 cursor-grab" />
            
            {/* Status indicator */}
            <span
              className={`w-2 h-2 rounded-full flex-shrink-0 ${
                tab.status === 'connected'
                  ? 'bg-green-500'
                  : tab.status === 'connecting'
                  ? 'bg-yellow-500 animate-pulse'
                  : tab.status === 'error'
                  ? 'bg-red-500'
                  : 'bg-gray-500'
              }`}
            />

            <span className="truncate flex-1">{tab.title}</span>

            <button
              onClick={(e) => {
                e.stopPropagation();
                closeTab(tab.id);
              }}
              className="p-0.5 rounded hover:bg-[#505050] text-gray-400 hover:text-white transition-colors"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ))}

        <button
          onClick={addTab}
          className="p-2 text-gray-400 hover:text-white hover:bg-[#3c3c3c] transition-colors"
          title="New tab"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>

      {/* Terminal content */}
      <div className="flex-1 relative">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={`absolute inset-0 ${tab.active ? 'block' : 'hidden'}`}
          >
            <Terminal
              id={tab.id}
              wsUrl={tab.wsUrl}
              onTitleChange={(title) => handleTitleChange(tab.id, title)}
              onConnectionChange={(status) => handleStatusChange(tab.id, status)}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
