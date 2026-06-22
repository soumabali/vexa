'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Download, FileText, FileCode, FileImage, FileVideo, FileAudio, FileArchive, AlertTriangle } from 'lucide-react';

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

interface FileSystem {
  name: string;
  host: string;
  port: number;
  username: string;
  path: string;
  protocol: 'sftp' | 'scp';
}

interface FilePreviewProps {
  file: FileItem;
  host: FileSystem | null;
  onDownload: (file: FileItem) => void;
}

export function FilePreview({ file, host, onDownload }: FilePreviewProps) {
  const [content, setContent] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadPreview();
  }, [file]);

  const loadPreview = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const url = host
        ? `/api/files/preview?host=${host.host}&path=${encodeURIComponent(file.path)}`
        : `/api/files/preview?path=${encodeURIComponent(file.path)}`;

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Failed to load preview');
      }

      const contentType = response.headers.get('content-type');

      if (contentType?.startsWith('text/') || contentType?.includes('json') || contentType?.includes('javascript')) {
        const text = await response.text();
        setContent(text);
      } else if (contentType?.startsWith('image/')) {
        const blob = await response.blob();
        const imageUrl = URL.createObjectURL(blob);
        setContent(imageUrl);
      } else if (contentType?.startsWith('video/')) {
        const blob = await response.blob();
        const videoUrl = URL.createObjectURL(blob);
        setContent(videoUrl);
      } else if (contentType?.startsWith('audio/')) {
        const blob = await response.blob();
        const audioUrl = URL.createObjectURL(blob);
        setContent(audioUrl);
      } else if (contentType === 'application/pdf') {
        const blob = await response.blob();
        const pdfUrl = URL.createObjectURL(blob);
        setContent(pdfUrl);
      } else {
        setContent(null);
        setError('Preview not available for this file type');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setIsLoading(false);
    }
  };

  const getFileType = () => {
    const ext = file.name.split('.').pop()?.toLowerCase();
    if (['txt', 'md', 'log', 'csv'].includes(ext || '')) return 'text';
    if (['js', 'ts', 'jsx', 'tsx', 'go', 'rs', 'py', 'java', 'cpp', 'c', 'json', 'yaml', 'yml', 'xml', 'html', 'css', 'sql'].includes(ext || '')) return 'code';
    if (['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp', 'bmp'].includes(ext || '')) return 'image';
    if (['mp4', 'avi', 'mov', 'mkv', 'webm'].includes(ext || '')) return 'video';
    if (['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'].includes(ext || '')) return 'audio';
    if (['zip', 'tar', 'gz', 'bz2', '7z', 'rar'].includes(ext || '')) return 'archive';
    return 'unknown';
  };

  const fileType = getFileType();

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <div className="flex justify-center">
          <Button onClick={() => onDownload(file)}>
            <Download className="h-4 w-4 mr-2" />
            Download Instead
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">{file.path}</p>
          <div className="flex items-center gap-2">
            <Badge variant="outline">{fileType.toUpperCase()}</Badge>
            <Badge variant="outline">{formatBytes(file.size)}</Badge>
            <Badge variant="outline">{file.permissions}</Badge>
          </div>
        </div>
        <Button onClick={() => onDownload(file)}>
          <Download className="h-4 w-4 mr-2" />
          Download
        </Button>
      </div>

      <div className="border rounded-lg overflow-hidden">
        {fileType === 'text' && content && (
          <pre className="p-4 text-sm font-mono bg-muted overflow-auto max-h-[500px] whitespace-pre-wrap">
            {content}
          </pre>
        )}

        {fileType === 'code' && content && (
          <pre className="p-4 text-sm font-mono bg-muted overflow-auto max-h-[500px]">
            <code>{content}</code>
          </pre>
        )}

        {fileType === 'image' && content && (
          <img
            src={content}
            alt={file.name}
            className="max-w-full max-h-[500px] mx-auto"
          />
        )}

        {fileType === 'video' && content && (
          <video
            src={content}
            controls
            className="max-w-full max-h-[500px] mx-auto"
          >
            Your browser does not support the video tag.
          </video>
        )}

        {fileType === 'audio' && content && (
          <div className="p-8 flex flex-col items-center">
            <FileAudio className="h-16 w-16 text-muted-foreground mb-4" />
            <audio src={content} controls className="w-full max-w-md" />
          </div>
        )}

        {fileType === 'archive' && (
          <div className="p-8 text-center">
            <FileArchive className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">
              Archive preview not available
            </p>
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => onDownload(file)}
            >
              <Download className="h-4 w-4 mr-2" />
              Download to Extract
            </Button>
          </div>
        )}

        {fileType === 'unknown' && (
          <div className="p-8 text-center">
            <FileText className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">
              Preview not available for this file type
            </p>
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => onDownload(file)}
            >
              <Download className="h-4 w-4 mr-2" />
              Download
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}
