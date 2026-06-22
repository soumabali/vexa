"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface FileNode {
  name: string;
  type: "file" | "directory";
  size?: number;
  children?: FileNode[];
}

interface FileManagerProps {
  type: 'local' | 'remote';
  host: Partial<FileSystem> | null;
  compact?: boolean;
  onTransferProgress?: (progress: { file: string; progress: number; speed: string; eta: string }) => void;
  onTransferComplete?: () => void;
}

export function FileManager({ type, host, compact, onTransferProgress, onTransferComplete }: FileManagerProps) {
  const [path, setPath] = useState("/home");
  const [files, setFiles] = useState<FileNode[]>([]);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Input value={path} onChange={(e) => setPath(e.target.value)} />
        <Button onClick={() => setFiles([])}>List</Button>
      </div>
      <p className="text-sm text-muted-foreground">Host: {host?.name ?? type}</p>
      <ul className="space-y-1">
        {files.map((f) => (
          <li key={f.name} className="text-sm">
            {f.type === "directory" ? "📁" : "📄"} {f.name}
          </li>
        ))}
        {files.length === 0 && (
          <p className="text-sm text-muted-foreground">No files loaded. Click List to fetch.</p>
        )}
      </ul>
    </div>
  );
}
