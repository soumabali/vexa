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
import { MaterialIcon } from "@/components/ui/material-icon";

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
          errors: result.success ? [] : result.error.issues.map((e) => `${e.path.join(".")}: ${e.message}`),
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
        <div className="flex items-center gap-3 rounded-lg border border-outline-variant bg-surface-container-low p-4">
          <MaterialIcon name="description" size="xl" className="text-on-surface-variant" />
          <div>
            <p className="text-headline-sm text-on-surface">Export all hosts to JSON</p>
            <p className="text-body-md text-on-surface-variant">
              Downloads a JSON file containing all your hosts with connection details
            </p>
          </div>
        </div>
        <Button onClick={handleExportAll} disabled={isLoading} className="w-full border border-outline-variant text-on-surface">
          {isLoading ? (
            <>
              <LoadingSpinner className="h-4 w-4 mr-2" />
              Exporting...
            </>
          ) : (
            <>
              <MaterialIcon name="file_export" size="sm" className="mr-2" />
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
        <Label className="text-label-md text-on-surface">Import Method</Label>
        <div className="flex gap-3">
          <label
            className="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border-2 border-dashed border-outline-variant p-4 transition-colors hover:border-outline hover:bg-surface-container-low"
          >
            <MaterialIcon name="upload" className="text-on-surface-variant" />
            <span className="text-label-md text-on-surface">Upload File</span>
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
        <Label htmlFor="paste-json" className="text-label-md text-on-surface">Or paste JSON directly</Label>
        <textarea
          id="paste-json"
          onChange={handlePasteJSON}
          placeholder='[{"name": "Server 1", "hostType": "ssh", "host": "192.168.1.1", "port": 22}]'
          rows={5}
          className="w-full rounded-md border border-outline-variant bg-surface-container px-3 py-2 text-body-md font-mono text-on-surface placeholder:text-on-surface-variant focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary"
        />
      </div>

      {/* Parsed hosts preview */}
      {parsedHosts.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label className="text-label-md text-on-surface">
              Preview ({parsedHosts.length} host{parsedHosts.length !== 1 ? "s" : ""})
            </Label>
            <div className="flex items-center gap-2 text-label-md">
              <Badge variant="default" className="bg-primary-container text-on-primary-container">
                {parsedHosts.filter((h) => h.valid).length} valid
              </Badge>
              {parsedHosts.filter((h) => !h.valid).length > 0 && (
                <Badge variant="destructive">
                  {parsedHosts.filter((h) => !h.valid).length} invalid
                </Badge>
              )}
            </div>
          </div>

          <div className="max-h-[200px] overflow-y-auto rounded-md border border-outline-variant">
            <table className="w-full text-xs">
              <thead className="bg-surface-container-low sticky top-0">
                <tr>
                  <th className="px-2 py-1.5 text-left text-label-md text-on-surface">Status</th>
                  <th className="px-2 py-1.5 text-left text-label-md text-on-surface">Name</th>
                  <th className="px-2 py-1.5 text-left text-label-md text-on-surface">Type</th>
                  <th className="px-2 py-1.5 text-left text-label-md text-on-surface">Host</th>
                </tr>
              </thead>
              <tbody>
                {parsedHosts.map((h, i) => (
                  <tr key={i} className="border-t border-outline-variant">
                    <td className="px-2 py-1.5 text-on-surface">
                      {h.valid ? (
                        <MaterialIcon name="check" size="sm" className="text-primary" />
                      ) : (
                        <div className="flex items-center gap-1">
                          <MaterialIcon name="close" size="sm" className="text-error" />
                          <span className="text-error">{h.errors[0]}</span>
                        </div>
                      )}
                    </td>
                    <td className="px-2 py-1.5 text-on-surface">{String((h.original as Record<string, unknown>)?.name || "-")}</td>
                    <td className="px-2 py-1.5 text-on-surface">{String((h.original as Record<string, unknown>)?.hostType || "-")}</td>
                    <td className="px-2 py-1.5 text-on-surface">{String((h.original as Record<string, unknown>)?.host || "-")}</td>
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
              className="h-4 w-4 rounded border-outline-variant"
            />
            <Label htmlFor="overwrite" className="text-label-md text-on-surface">
              Overwrite existing hosts with same name
            </Label>
          </div>

          {/* Import button */}
          <Button
            onClick={handleImport}
            disabled={isLoading || parsedHosts.filter((h) => h.valid).length === 0}
            className="w-full bg-primary-container text-on-primary-container rounded-lg"
          >
            {isLoading ? (
              <>
                <LoadingSpinner className="h-4 w-4 mr-2" />
                Importing...
              </>
            ) : (
              <>
                <MaterialIcon name="file_import" size="sm" className="mr-2" />
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
        <div className="flex items-center gap-2 rounded-lg border border-outline-variant bg-surface-container p-4">
          <MaterialIcon name="check" className="text-primary" />
          <div>
            <p className="text-headline-sm text-on-surface">Import successful!</p>
            <p className="text-body-md text-on-surface-variant">
              {importResult.imported} imported
              {importResult.skipped > 0 && `, ${importResult.skipped} skipped`}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}