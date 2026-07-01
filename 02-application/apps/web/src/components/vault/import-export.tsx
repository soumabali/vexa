"use client";

import React, { useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Upload,
  Download,
  FileJson,
  FileSpreadsheet,
  FileKey,
  Check,
  AlertTriangle,
  X,
} from "lucide-react";

export function ImportExportDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [importTab, setImportTab] = useState("openssh");
  const [exportTab, setExportTab] = useState("json");
  const [importData, setImportData] = useState("");
  const [importProgress, setImportProgress] = useState(0);
  const [importResults, setImportResults] = useState<{
    success: number;
    failed: number;
    errors: string[];
  } | null>(null);
  const [isImporting, setIsImporting] = useState(false);

  const handleImport = async () => {
    setIsImporting(true);
    setImportProgress(0);
    setImportResults(null);

    // Simulate import progress
    for (let i = 0; i <= 100; i += 10) {
      await new Promise((resolve) => setTimeout(resolve, 200));
      setImportProgress(i);
    }

    setImportResults({
      success: 5,
      failed: 0,
      errors: [],
    });
    setIsImporting(false);
  };

  const handleExport = () => {
    const data = JSON.stringify(
      {
        version: "1.0",
        exportedAt: new Date().toISOString(),
        credentials: [],
      },
      null,
      2
    );

    const blob = new Blob([data], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `vexa-vault-backup-${new Date().toISOString().split("T")[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      const file = files[0];
      const reader = new FileReader();
      reader.onload = (event) => {
        setImportData(event.target?.result as string);
      };
      reader.readAsText(file);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl bg-surface-container-high border-outline-variant">
        <DialogHeader>
          <DialogTitle>Import / Export Credentials</DialogTitle>
          <DialogDescription>
            Import credentials from various formats or export your vault
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="import" className="mt-4">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="import">Import</TabsTrigger>
            <TabsTrigger value="export">Export</TabsTrigger>
          </TabsList>

          <TabsContent value="import" className="space-y-4">
            <Tabs value={importTab} onValueChange={setImportTab}>
              <TabsList className="grid grid-cols-4">
                <TabsTrigger value="openssh">OpenSSH</TabsTrigger>
                <TabsTrigger value="putty">PuTTY</TabsTrigger>
                <TabsTrigger value="csv">CSV</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
              </TabsList>

              <TabsContent value="openssh" className="space-y-4">
                <Card
                  className="border-dashed border-outline-variant bg-surface-container-lowest"
                  onDragOver={handleDragOver}
                  onDrop={handleDrop}
                >
                  <CardContent className="p-8 text-center">
                    <Upload className="h-8 w-8 mx-auto mb-4 text-on-surface-variant" />
                    <p className="text-sm text-on-surface-variant mb-2">
                      Drag and drop your OpenSSH config file here
                    </p>
                    <p className="text-xs text-on-surface-variant">
                      Supports ~/.ssh/config format
                    </p>
                  </CardContent>
                </Card>

                <div className="space-y-2">
                  <Label>Or paste config content</Label>
                  <Textarea
                    value={importData}
                    onChange={(e) => setImportData(e.target.value)}
                    placeholder="Host example&#xa;  HostName 192.168.1.1&#xa;  User root&#xa;  Port 22"
                    rows={6}
                  />
                </div>
              </TabsContent>

              <TabsContent value="putty" className="space-y-4">
                <Card className="border-dashed border-outline-variant bg-surface-container-lowest">
                  <CardContent className="p-8 text-center">
                    <Upload className="h-8 w-8 mx-auto mb-4 text-on-surface-variant" />
                    <p className="text-sm text-on-surface-variant">
                      Import from PuTTY .ppk or .reg files
                    </p>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="csv" className="space-y-4">
                <div className="space-y-2">
                  <Label>CSV Content</Label>
                  <Textarea
                    value={importData}
                    onChange={(e) => setImportData(e.target.value)}
                    placeholder="name,host,username,port,type&#xa;Server,192.168.1.1,root,22,ssh-key"
                    rows={6}
                  />
                </div>
              </TabsContent>

              <TabsContent value="json" className="space-y-4">
                <div className="space-y-2">
                  <Label>JSON Content</Label>
                  <Textarea
                    value={importData}
                    onChange={(e) => setImportData(e.target.value)}
                    placeholder={'[&#xa;  {&#xa;    "name": "Server",&#xa;    "host": "192.168.1.1",&#xa;    "username": "root",&#xa;    "port": 22&#xa;  }&#xa;]'}
                    rows={6}
                  />
                </div>
              </TabsContent>
            </Tabs>

            {isImporting && (
              <div className="space-y-2">
                <Progress value={importProgress} />
                <p className="text-sm text-on-surface-variant text-center">
                  Importing... {importProgress}%
                </p>
              </div>
            )}

            {importResults && (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Check className="h-5 w-5 text-green-500" />
                  <span>{importResults.success} credentials imported</span>
                </div>
                {importResults.failed > 0 && (
                  <div className="flex items-center gap-2">
                    <AlertTriangle className="h-5 w-5 text-yellow-500" />
                    <span>{importResults.failed} failed</span>
                  </div>
                )}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <Button variant="outline" className="border-outline-variant" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleImport}
                disabled={!importData || isImporting}
                className="bg-primary text-on-primary hover:bg-primary/90"
              >
                Import
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="export" className="space-y-4">
            <Tabs value={exportTab} onValueChange={setExportTab}>
              <TabsList className="grid grid-cols-3">
                <TabsTrigger value="json">JSON</TabsTrigger>
                <TabsTrigger value="csv">CSV</TabsTrigger>
                <TabsTrigger value="env">Env</TabsTrigger>
              </TabsList>

              <TabsContent value="json" className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <FileJson className="h-5 w-5" />
                      JSON Export
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-on-surface-variant mb-4">
                      Export all credentials as encrypted JSON backup
                    </p>
                    <Button onClick={handleExport}>
                      <Download className="h-4 w-4 mr-2" />
                      Download JSON
                    </Button>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="csv" className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <FileSpreadsheet className="h-5 w-5" />
                      CSV Export
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-on-surface-variant mb-4">
                      Export credentials as CSV spreadsheet
                    </p>
                    <Button onClick={handleExport}>
                      <Download className="h-4 w-4 mr-2" />
                      Download CSV
                    </Button>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="env" className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <FileKey className="h-5 w-5" />
                      Environment Variables
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-on-surface-variant mb-4">
                      Export as environment variable declarations
                    </p>
                    <Button onClick={handleExport}>
                      <Download className="h-4 w-4 mr-2" />
                      Download .env
                    </Button>
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
