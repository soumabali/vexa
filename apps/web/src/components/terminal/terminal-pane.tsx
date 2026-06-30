'use client';

import React, { useState, useCallback } from 'react';
import { MaterialIcon } from '@/components/ui/material-icon';
import TerminalTabs, { Tab } from './terminal-tabs';

export interface TerminalPane {
  id: string;
  type: 'terminal' | 'split';
  direction?: 'horizontal' | 'vertical';
  children?: TerminalPane[];
  tabData?: {
    tabs?: Tab[];
    activeTabId?: string | null;
  };
  size?: number; // percentage
}

export interface TerminalPaneProps {
  pane: TerminalPane;
  defaultWsUrl: string;
  onPaneChange?: (pane: TerminalPane) => void;
}

let paneCounter = 0;

function generatePaneId(): string {
  paneCounter += 1;
  return `pane-${Date.now()}-${paneCounter}`;
}

export default function TerminalPaneComponent({
  pane,
  defaultWsUrl,
  onPaneChange,
}: TerminalPaneProps) {
  const [localPane, setLocalPane] = useState<TerminalPane>(pane);

  const updatePane = useCallback(
    (updater: (prev: TerminalPane) => TerminalPane) => {
      setLocalPane((prev) => {
        const next = updater(prev);
        onPaneChange?.(next);
        return next;
      });
    },
    [onPaneChange]
  );

  const splitPane = useCallback(
    (direction: 'horizontal' | 'vertical') => {
      updatePane((prev) => {
        if (prev.type === 'split') return prev;

        const newPane: TerminalPane = {
          id: generatePaneId(),
          type: 'terminal',
          tabData: {
            tabs: [
              {
                id: `tab-${Date.now()}`,
                title: 'Local',
                status: 'connecting',
                wsUrl: defaultWsUrl,
                active: true,
              },
            ],
            activeTabId: null,
          },
          size: 50,
        };

        return {
          ...prev,
          type: 'split',
          direction,
          children: [
            { ...prev, size: 50 },
            newPane,
          ],
        };
      });
    },
    [defaultWsUrl, updatePane]
  );

  const handleTabChange = useCallback(
    (tabs: Tab[]) => {
      updatePane((prev) => ({
        ...prev,
        tabData: {
          ...prev.tabData,
          tabs,
        },
      }));
    },
    [updatePane]
  );

  const handleActiveTabChange = useCallback(
    (tabId: string | null) => {
      updatePane((prev) => ({
        ...prev,
        tabData: {
          ...prev.tabData,
          activeTabId: tabId,
        },
      }));
    },
    [updatePane]
  );

  if (localPane.type === 'split' && localPane.children) {
    const flexDirection = localPane.direction === 'horizontal' ? 'flex-row' : 'flex-col';

    return (
      <div className={`flex ${flexDirection} h-full gap-1 bg-black`}>
        {localPane.children.map((child, index) => (
          <div
            key={child.id}
            className="relative"
            style={{ flex: child.size || 50 }}
          >
            <TerminalPaneComponent
              pane={child}
              defaultWsUrl={defaultWsUrl}
              onPaneChange={(updatedChild) => {
                updatePane((prev) => {
                  if (prev.type !== 'split' || !prev.children) return prev;
                  const newChildren = [...prev.children];
                  newChildren[index] = updatedChild;
                  return { ...prev, children: newChildren };
                });
              }}
            />
          </div>
        ))}
      </div>
    );
  }

  // Terminal pane
  return (
    <div className="h-full relative group bg-black">
      <TerminalTabs
        initialTabs={localPane.tabData?.tabs || []}
        defaultWsUrl={defaultWsUrl}
        onTabChange={handleTabChange}
        onActiveTabChange={handleActiveTabChange}
      />

      {/* Split controls */}
      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1 z-10">
        <button
          onClick={() => splitPane('horizontal')}
          className="p-1.5 bg-surface-container rounded text-on-surface-variant hover:text-on-surface hover:bg-surface-variant transition-colors"
          title="Split horizontal"
        >
          <MaterialIcon name="grid_view" size="sm" />
        </button>
        <button
          onClick={() => splitPane('vertical')}
          className="p-1.5 bg-surface-container rounded text-on-surface-variant hover:text-on-surface hover:bg-surface-variant transition-colors"
          title="Split vertical"
        >
          <MaterialIcon name="expand_more" size="sm" />
        </button>
      </div>
    </div>
  );
}

// export type { TerminalPane }; // Already exported as interface above
