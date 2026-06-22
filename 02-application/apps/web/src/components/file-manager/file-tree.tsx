'use client';

import React, { useState } from 'react';
import { ChevronRight, ChevronDown, Folder, File } from 'lucide-react';

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

interface FileTreeProps {
  files: FileItem[];
  selectedFiles: Set<string>;
  onSelect: (path: string, multi: boolean) => void;
  onFileClick: (file: FileItem) => void;
  currentPath: string;
  onNavigate: (path: string) => void;
}

interface TreeNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  children: TreeNode[];
  isOpen: boolean;
}

export function FileTree({
  files,
  selectedFiles,
  onSelect,
  onFileClick,
  currentPath,
  onNavigate,
}: FileTreeProps) {
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set(['/']));

  const toggleExpanded = (path: string) => {
    setExpandedPaths((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const buildTree = (fileList: FileItem[]): TreeNode[] => {
    const root: TreeNode[] = [];
    const map = new Map<string, TreeNode>();

    // Sort by path depth
    const sorted = [...fileList].sort((a, b) => a.path.localeCompare(b.path));

    sorted.forEach((file) => {
      const parts = file.path.split('/').filter(Boolean);
      let currentPath = '';
      let currentLevel = root;

      parts.forEach((part, index) => {
        currentPath = currentPath + '/' + part;
        const isLast = index === parts.length - 1;
        const existing = currentLevel.find((n) => n.path === currentPath);

        if (existing) {
          if (isLast) {
            existing.type = file.type;
          }
          currentLevel = existing.children;
        } else {
          const newNode: TreeNode = {
            name: part,
            path: currentPath,
            type: isLast ? file.type : 'directory',
            children: [],
            isOpen: expandedPaths.has(currentPath),
          };
          currentLevel.push(newNode);
          currentLevel = newNode.children;
        }
      });
    });

    return root;
  };

  const tree = buildTree(files);

  const renderNode = (node: TreeNode, depth: number = 0) => {
    const isSelected = selectedFiles.has(node.path);
    const isExpanded = expandedPaths.has(node.path);

    return (
      <div key={node.path}>
        <div
          className={`flex items-center py-1 px-2 cursor-pointer hover:bg-accent ${
            isSelected ? 'bg-accent' : ''
          } ${currentPath === node.path ? 'font-medium' : ''}`}
          style={{ paddingLeft: `${depth * 20 + 8}px` }}
          onClick={(e) => {
            if (e.ctrlKey || e.metaKey) {
              onSelect(node.path, true);
            } else {
              onSelect(node.path, false);
              if (node.type === 'directory') {
                toggleExpanded(node.path);
                onNavigate(node.path);
              } else {
                const file = files.find((f) => f.path === node.path);
                if (file) onFileClick(file);
              }
            }
          }}
        >
          {node.type === 'directory' ? (
            <span
              className="mr-1 cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpanded(node.path);
              }}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4 inline" />
              ) : (
                <ChevronRight className="h-4 w-4 inline" />
              )}
            </span>
          ) : (
            <span className="w-4 mr-1 inline-block" />
          )}
          {node.type === 'directory' ? (
            <Folder className="h-4 w-4 mr-2 text-blue-500 inline" />
          ) : (
            <File className="h-4 w-4 mr-2 text-gray-400 inline" />
          )}
          <span className="text-sm">{node.name}</span>
        </div>
        {node.type === 'directory' && isExpanded && node.children.length > 0 && (
          <div>{node.children.map((child) => renderNode(child, depth + 1))}</div>
        )}
      </div>
    );
  };

  return (
    <div className="py-2">
      {tree.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">No files</div>
      ) : (
        tree.map((node) => renderNode(node))
      )}
    </div>
  );
}
