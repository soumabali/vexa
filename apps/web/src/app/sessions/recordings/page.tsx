"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Play,
  Download,
  Trash2,
  Search,
  Clock,
  Monitor,
  FileText,
  Calendar,
  X,
} from "lucide-react";
import { RecordingPlayer } from "@/components/recording-player";

interface Recording {
  id: string;
  session_id: string;
  host_id: string;
  status: string;
  started_at: string;
  ended_at?: string;
  duration: number;
  file_size_bytes: number;
  terminal_type: string;
  shell: string;
  command_history: string[];
  created_at: string;
}

export default function RecordingsPage() {
  const router = useRouter();
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRecording, setSelectedRecording] = useState<Recording | null>(
    null
  );
  const [isPlayerOpen, setIsPlayerOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const fetchRecordings = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/recordings?limit=50&offset=${(currentPage - 1) * 50}`);
      const data = await response.json();
      setRecordings(data.recordings || []);
      setTotalPages(Math.ceil((data.total || 0) / 50));
    } catch (error) {
      console.error("Failed to fetch recordings:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRecordings();
  }, [currentPage]);

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      fetchRecordings();
      return;
    }

    try {
      setLoading(true);
      const response = await fetch(`/api/recordings/search?q=${encodeURIComponent(searchQuery)}`);
      const data = await response.json();
      setRecordings(data.recordings || []);
    } catch (error) {
      console.error("Search failed:", error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlay = (recording: Recording) => {
    setSelectedRecording(recording);
    setIsPlayerOpen(true);
  };

  const handleDownload = (recordingId: string) => {
    window.open(`/api/recordings/${recordingId}/download`, "_blank");
  };

  const handleDelete = async (recordingId: string) => {
    if (!confirm("Are you sure you want to delete this recording?")) return;

    try {
      await fetch(`/api/recordings/${recordingId}`, { method: "DELETE" });
      setRecordings(recordings.filter((r) => r.id !== recordingId));
    } catch (error) {
      console.error("Failed to delete recording:", error);
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}m ${secs}s`;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, string> = {
      recording: "bg-red-500",
      completed: "bg-green-500",
      failed: "bg-red-700",
      deleted: "bg-gray-500",
    };

    return (
      <Badge className={`${variants[status] || "bg-gray-500"} text-white`}>
        {status}
      </Badge>
    );
  };

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Session Recordings</h1>
        <p className="text-gray-400">
          Browse, search, and replay your terminal session recordings
        </p>
      </div>

      {/* Search Bar */}
      <div className="flex gap-2 mb-6">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search recordings by content, commands, or host..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            className="pl-10"
          />
        </div>
        <Button onClick={handleSearch}>Search</Button>
        {searchQuery && (
          <Button variant="outline" onClick={() => { setSearchQuery(""); fetchRecordings(); }}>
            Clear
          </Button>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <div className="text-2xl font-bold">{recordings.length}</div>
          <div className="text-sm text-gray-400">Total Recordings</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold">
            {formatDuration(
              recordings.reduce((acc, r) => acc + (r.duration || 0), 0)
            )}
          </div>
          <div className="text-sm text-gray-400">Total Duration</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold">
            {formatFileSize(
              recordings.reduce((acc, r) => acc + (r.file_size_bytes || 0), 0)
            )}
          </div>
          <div className="text-sm text-gray-400">Total Size</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold">
            {
              recordings.filter((r) => r.status === "recording").length
            }
          </div>
          <div className="text-sm text-gray-400">Currently Recording</div>
        </Card>
      </div>

      {/* Recordings Table */}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Session</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead>Size</TableHead>
              <TableHead>Terminal</TableHead>
              <TableHead>Recorded</TableHead>
              <TableHead>Commands</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white mx-auto"></div>
                </TableCell>
              </TableRow>
            ) : recordings.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-gray-400">
                  No recordings found. Start a session to record your terminal.
                </TableCell>
              </TableRow>
            ) : (
              recordings.map((recording) => (
                <TableRow key={recording.id}>
                  <TableCell>{getStatusBadge(recording.status)}</TableCell>
                  <TableCell>
                    <div className="font-mono text-sm">
                      {recording.session_id.slice(0, 8)}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3 text-gray-400" />
                      {formatDuration(recording.duration)}
                    </div>
                  </TableCell>
                  <TableCell>{formatFileSize(recording.file_size_bytes)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Monitor className="h-3 w-3 text-gray-400" />
                      {recording.terminal_type}
                    </div>
                  </TableCell>
                  <TableCell>{formatDate(recording.started_at)}</TableCell>
                  <TableCell>
                    <div className="text-sm text-gray-400">
                      {recording.command_history?.length || 0} commands
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handlePlay(recording)}
                        disabled={recording.status !== "completed"}
                      >
                        <Play className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDownload(recording.id)}
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(recording.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-6">
          <Button
            variant="outline"
            onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
            disabled={currentPage === 1}
          >
            Previous
          </Button>
          <span className="flex items-center px-4">
            Page {currentPage} of {totalPages}
          </span>
          <Button
            variant="outline"
            onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
            disabled={currentPage === totalPages}
          >
            Next
          </Button>
        </div>
      )}

      {/* Player Dialog */}
      <Dialog open={isPlayerOpen} onOpenChange={setIsPlayerOpen}>
        <DialogContent className="max-w-6xl w-full bg-black border-gray-800">
          <DialogHeader>
            <DialogTitle className="flex items-center justify-between">
              <span>Session Recording</span>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsPlayerOpen(false)}
              >
                <X className="h-4 w-4" />
              </Button>
            </DialogTitle>
          </DialogHeader>
          {selectedRecording && (
            <RecordingPlayer
              recordingId={selectedRecording.id}
              onClose={() => setIsPlayerOpen(false)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
