'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  ChevronLeft,
  ChevronRight,
  ChevronUp,
  RefreshCw,
  Plus,
  Upload,
  Download,
  Trash2,
  Copy,
  Scissors,
  ClipboardPaste,
  Search,
  Grid3X3,
  List,
  TreePine,
  Star,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Folder,
  Home,
  Clock,
  Bookmark,
  X,
} from 'lucide-react';

interface FileToolbarProps {
  currentPath: string;
  viewMode: 'list' | 'grid' | 'tree';
  onViewModeChange: (mode: 'list' | 'grid' | 'tree') => void;
  onNavigateUp: () => void;
  onNavigateBack: () => void;
  onNavigateForward: () => void;
  canNavigateBack: boolean;
  canNavigateForward: boolean;
  selectedCount: number;
  onSelectAll: () => void;
  onDelete: () => void;
  onCopy: () => void;
  onCut: () => void;
  onPaste: () => void;
  canPaste: boolean;
  onUpload: () => void;
  onRefresh: () => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  sortBy: 'name' | 'size' | 'date' | 'type';
  onSortChange: (sortBy: 'name' | 'size' | 'date' | 'type') => void;
  sortOrder: 'asc' | 'desc';
  onSortOrderChange: () => void;
  bookmarks: string[];
  onBookmarkToggle: (path: string) => void;
  onBookmarkNavigate: (path: string) => void;
  isBookmarked: boolean;
}

export function FileToolbar({
  currentPath,
  viewMode,
  onViewModeChange,
  onNavigateUp,
  onNavigateBack,
  onNavigateForward,
  canNavigateBack,
  canNavigateForward,
  selectedCount,
  onSelectAll,
  onDelete,
  onCopy,
  onCut,
  onPaste,
  canPaste,
  onUpload,
  onRefresh,
  searchQuery,
  onSearchChange,
  sortBy,
  onSortChange,
  sortOrder,
  onSortOrderChange,
  bookmarks,
  onBookmarkToggle,
  onBookmarkNavigate,
  isBookmarked,
}: FileToolbarProps) {
  const pathParts = currentPath.split('/').filter(Boolean);

  return (
    <div className="space-y-2">
      {/* Navigation Bar */}
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={onNavigateBack}
          disabled={!canNavigateBack}
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={onNavigateForward}
          disabled={!canNavigateForward}
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="sm" onClick={onNavigateUp}>
          <ChevronUp className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="sm" onClick={onRefresh}>
          <RefreshCw className="h-4 w-4" />
        </Button>

        {/* Breadcrumb */}
        <div className="flex-1 flex items-center gap-1 px-2 py-1 bg-muted rounded-md text-sm">
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-1"
            onClick={() => onBookmarkNavigate('/')}
          >
            <Home className="h-3 w-3" />
          </Button>
          {pathParts.map((part, index) => {
            const path = '/' + pathParts.slice(0, index + 1).join('/');
            return (
              <React.Fragment key={path}>
                <ChevronRight className="h-3 w-3 text-muted-foreground" />
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-1 text-xs"
                  onClick={() => onBookmarkNavigate(path)}
                >
                  {part}
                </Button>
              </React.Fragment>
            );
          })}
        </div>

        {/* Bookmark Toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onBookmarkToggle(currentPath)}
        >
          <Star
            className={`h-4 w-4 ${
              isBookmarked ? 'text-yellow-500 fill-yellow-500' : ''
            }`}
          />
        </Button>
      </div>

      {/* Action Bar */}
      <div className="flex items-center gap-2">
        {/* Search */}
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search files..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-8"
          />
          {searchQuery && (
            <Button
              variant="ghost"
              size="sm"
              className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6 p-0"
              onClick={() => onSearchChange('')}
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>

        <div className="flex-1" />

        {/* Selection Actions */}
        {selectedCount > 0 && (
          <>
            <Badge variant="secondary">{selectedCount} selected</Badge>
            <Button variant="ghost" size="sm" onClick={onSelectAll}>
              All
            </Button>
            <Button variant="ghost" size="sm" onClick={onCopy}>
              <Copy className="h-4 w-4 mr-1" />
              Copy
            </Button>
            <Button variant="ghost" size="sm" onClick={onCut}>
              <Scissors className="h-4 w-4 mr-1" />
              Cut
            </Button>
            {canPaste && (
              <Button variant="ghost" size="sm" onClick={onPaste}>
                <ClipboardPaste className="h-4 w-4 mr-1" />
                Paste
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              className="text-red-600"
              onClick={onDelete}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              Delete
            </Button>
          </>
        )}

        <Button variant="ghost" size="sm" onClick={onUpload}>
          <Upload className="h-4 w-4 mr-1" />
          Upload
        </Button>

        {/* View Mode */}
        <div className="flex items-center border rounded-md">
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="sm"
            className="rounded-none"
            onClick={() => onViewModeChange('list')}
          >
            <List className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="sm"
            className="rounded-none"
            onClick={() => onViewModeChange('grid')}
          >
            <Grid3X3 className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'tree' ? 'secondary' : 'ghost'}
            size="sm"
            className="rounded-none"
            onClick={() => onViewModeChange('tree')}
          >
            <TreePine className="h-4 w-4" />
          </Button>
        </div>

        {/* Sort */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <ArrowUpDown className="h-4 w-4 mr-1" />
              Sort
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {(['name', 'size', 'date', 'type'] as const).map((sort) => (
              <DropdownMenuItem
                key={sort}
                onClick={() => {
                  if (sortBy === sort) {
                    onSortOrderChange();
                  } else {
                    onSortChange(sort);
                  }
                }}
              >
                <span className="capitalize">{sort}</span>
                {sortBy === sort && (
                  sortOrder === 'asc' ? (
                    <ArrowUp className="h-3 w-3 ml-2" />
                  ) : (
                    <ArrowDown className="h-3 w-3 ml-2" />
                  )
                )}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Bookmarks */}
        {bookmarks.length > 0 && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <Bookmark className="h-4 w-4 mr-1" />
                Bookmarks
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {bookmarks.map((path) => (
                <DropdownMenuItem
                  key={path}
                  onClick={() => onBookmarkNavigate(path)}
                >
                  {path}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </div>
  );
}
