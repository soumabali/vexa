"use client";

import React, { useState, useCallback } from "react";
import { hostsApi } from "@/lib/api/hosts";
import { CreateHostInput, createHostSchema } from "@/lib/validations/hosts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import {
  UploadIcon,
  DownloadIcon,
  FileJsonIcon,
  AlertTriangleIcon,
  CheckIcon,
  XIcon,
  ClipboardIcon,
} from "lucide-react";

interface HostImportExportProps {
  mode: "import" | "export";
  onImportComplete?: (count: number) => void;
  onExportComplete?: (count: number) => void;
}

interface ParsedHost {
  data: CreateHostInput;
  valid: boolean;
  errors: string[];
  original: unknown;
}

export function HostImportExport({ mode, onImportComplete, onExportComplete }: HostImportExportProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [overwrite, setOverwrite] = useState(false);
  const [parsedHosts, setParsedHosts] = useState<ParsedHost[]>([]);
  const [importResult, setImportResult] = useState<{ imported: number; skipped: number } | null>(null);

  const parseJSON = useCallback((text: string): ParsedHost[] => {
    try {
      const json = JSON.parse(text);
      const hosts = Array.isArray(json) ? json : json.hosts || [];
      return hosts.map((h: unknown) => {
        const result = createHostSchema.safeParse(h);
        return {
          data: result.success ? (result.data as CreateHostInput) : ({} as CreateHostInput),
          valid: result.success,
          errors: result.success ? [] : (result.error as any).issues.map((e: any) => `${e.path.join(".")}: ${e.message}`),
          original: h,
        };
      });
    } catch (e) {
      return [];
    }
  }, []);

  const handleFileUpload = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;
      setError("");
      setIsLoading(true);
      try {
        const text = await file.text();
        const parsed = parseJSON(text);
        setParsedHosts(parsed);
      } catch (err) {
        setError("Failed to parse file. Please ensure it's valid JSON.");
      } finally {
        setIsLoading(false);
      }
    },
    [parseJSON]
  );

  const handlePasteJSON = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const text = e.target.value;
      if (!text.trim()) {
        setParsedHosts([]);
        return;
      }
      const parsed = parseJSON(text);
      setParsedHosts(parsed);
    },
    [parseJSON]
  );

  const handleImport = async () => {
    if (parsedHosts.length === 0) return;
    setIsLoading(true);
    setError("");
    try {
      const validHosts = parsedHosts.filter((h) => h.valid).map((h) => h.data);
      const result = await hostsApi.import(validHosts, overwrite);
      setImportResult(result);
      onImportComplete?.(result.imported);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Import failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleExportAll = async () => {
    setIsLoading(true);
    setError("");
    try {
      const hosts = await hostsApi.export();
      const json = JSON.stringify(hosts, null, 2);
      const blob = new Blob([json], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `hosts-export-${new Date().toISOString().split("T")[0]}.json`;
      a.click();
      onExportComplete?.(hosts.length);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Export failed");
    } finally {
      setIsLoading(false);
    }
  };

  if (mode === "export") {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-3 rounded-lg border bg-accent/50 p-4">
          <FileJsonIcon className="h-8 w-8 text-muted-foreground" />
          <div>
            <p className="font-medium">Export all hosts to JSON</p>
            <p className="text-sm text-muted-foreground">
              Downloads a JSON file containing all your hosts with connection details
            </p>
          </div>
        </div>
        <Button onClick={handleExportAll} disabled={isLoading} className="w-full">
          {isLoading ? (
            <>
              <LoadingSpinner className="h-4 w-4 mr-2" />
              Exporting...
            </>
          ) : (
            <>
              <DownloadIcon className="h-4 w-4 mr-2" />
              Export All Hosts
            </>
          )}
        </Button>
        {error && <ErrorDisplay message={error} />}
      </div>
    );
  }

  // Import mode
  return (
    <div className="space-y-4">
      {/* Upload method */}
      <div className="space-y-3">
        <Label>Import Method</Label>
        <div className="flex gap-3">
          <label
            className="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border-2 border-dashed p-4 transition-colors hover:border-primary hover:bg-primary/5"
          >
            <UploadIcon className="h-5 w-5 text-muted-foreground" />
            <span className="text-sm font-medium">Upload File</span>
            <input
              type="file"
              accept=".json"
              onChange={handleFileUpload}
              className="hidden"
            />
          </label>
        </div>
      </div>

      {/* Or paste JSON */}
      <div className="space-y-2">
        <Label htmlFor="paste-json">Or paste JSON directly</Label>
        <textarea
          id="paste-json"
          onChange={handlePasteJSON}
          placeholder='[{"name": "Server 1", "hostType": "ssh", "host": "192.168.1.1", "port": 22}]'
          rows={5}
          className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>

      {/* Parsed hosts preview */}
      {parsedHosts.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label>
              Preview ({parsedHosts.length} host{parsedHosts.length !== 1 ? "s" : ""})
            </Label>
            <div className="flex items-center gap-2 text-sm">
              <Badge variant="default" className="bg-emerald-500">
                {parsedHosts.filter((h) => h.valid).length} valid
              </Badge>
              {parsedHosts.filter((h) => !h.valid).length > 0 && (
                <Badge variant="destructive">
                  {parsedHosts.filter((h) => !h.valid).length} invalid
                </Badge>
              )}
            </div>
          </div>

          <div className="max-h-[200px] overflow-y-auto rounded-md border">
            <table className="w-full text-xs">
              <thead className="bg-muted sticky top-0">
                <tr>
                  <th className="px-2 py-1.5 text-left font-medium">Status</th>
                  <th className="px-2 py-1.5 text-left font-medium">Name</th>
                  <th className="px-2 py-1.5 text-left font-medium">Type</th>
                  <th className="px-2 py-1.5 text-left font-medium">Host</th>
                </tr>
              </thead>
              <tbody>
                {parsedHosts.map((h, i) => (
                  <tr key={i} className="border-t">
                    <td className="px-2 py-1.5">
                      {h.valid ? (
                        <CheckIcon className="h-4 w-4 text-emerald-500" />
                      ) : (
                        <div className="flex items-center gap-1">
                          <XIcon className="h-4 w-4 text-red-500" />
                          <span className="text-red-500">{h.errors[0]}</span>
                        </div>
                      )}
                    </td>
                    <td className="px-2 py-1.5">{String((h.original as Record<string, unknown>)?.name || "-")}</td>
                    <td className="px-2 py-1.5">{String((h.original as Record<string, unknown>)?.hostType || "-")}</td>
                    <td className="px-2 py-1.5">{String((h.original as Record<string, unknown>)?.host || "-")}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Overwrite option */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="overwrite"
              checked={overwrite}
              onChange={(e) => setOverwrite(e.target.checked)}
              className="h-4 w-4 rounded border-input"
            />
            <Label htmlFor="overwrite" className="text-sm">
              Overwrite existing hosts with same name
            </Label>
          </div>

          {/* Import button */}
          <Button
            onClick={handleImport}
            disabled={isLoading || parsedHosts.filter((h) => h.valid).length === 0}
            className="w-full"
          >
            {isLoading ? (
              <>
                <LoadingSpinner className="h-4 w-4 mr-2" />
                Importing...
              </>
            ) : (
              <>
                <UploadIcon className="h-4 w-4 mr-2" />
                Import {parsedHosts.filter((h) => h.valid).length} Host
                {parsedHosts.filter((h) => h.valid).length !== 1 ? "s" : ""}
              </>
            )}
          </Button>
        </div>
      )}

      {error && <ErrorDisplay message={error} />}

      {/* Import result */}
      {importResult && (
        <div className="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 p-4">
          <CheckIcon className="h-5 w-5 text-emerald-600" />
          <div>
            <p className="font-medium text-emerald-800">Import successful!</p>
            <p className="text-sm text-emerald-700">
              {importResult.imported} imported
              {importResult.skipped > 0 && `, ${importResult.skipped} skipped`}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}