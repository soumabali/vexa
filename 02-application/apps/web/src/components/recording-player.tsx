"use client";

import React, { useEffect, useRef, useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Download,
  Share2,
  Trash2,
  BookOpen,
  Clock,
  Maximize2,
  Minimize2,
  Volume2,
  VolumeX,
  Settings,
} from "lucide-react";

interface RecordingPlayerProps {
  recordingId: string;
  autoPlay?: boolean;
  onClose?: () => void;
}

export function RecordingPlayer({
  recordingId,
  autoPlay = false,
  onClose,
}: RecordingPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [speed, setSpeed] = useState(1.0);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [volume, setVolume] = useState(1);
  const [showSettings, setShowSettings] = useState(false);
  const [theme, setTheme] = useState("monokai");

  // Use asciinema-player library
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // Load asciinema player dynamically
    const loadPlayer = async () => {
      try {
        const AsciinemaPlayer = await import("asciinema-player");

        const player = AsciinemaPlayer.create(
          `/api/recordings/${recordingId}/play`,
          container,
          {
            autoPlay,
            speed,
            theme,
            fit: "both",
            cols: 80,
            rows: 24,
            onReady: () => {
              console.log("Player ready");
            },
            onPlay: () => setIsPlaying(true),
            onPause: () => setIsPlaying(false),
            onTimeUpdate: (time: number) => {
              setCurrentTime(time);
            },
          }
        );

        // Get duration
        player.getDuration().then((dur: number) => {
          setDuration(dur);
        });

        return () => {
          player.dispose();
        };
      } catch (error) {
        console.error("Failed to load asciinema player:", error);
      }
    };

    const cleanup = loadPlayer();
    return () => {
      cleanup.then((fn) => fn?.());
    };
  }, [recordingId, autoPlay, speed, theme]);

  const togglePlay = () => {
    // Access player instance through ref or global
    // Implementation depends on how we store the player reference
  };

  const handleSpeedChange = (value: number[]) => {
    setSpeed(value[0]);
  };

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  };

  const handleDownload = () => {
    window.open(`/api/recordings/${recordingId}/download`, "_blank");
  };

  const handleShare = () => {
    const url = `${window.location.origin}/recordings/${recordingId}`;
    navigator.clipboard.writeText(url);
    // Show toast notification
  };

  const handleDelete = () => {
    if (confirm("Are you sure you want to delete this recording?")) {
      fetch(`/api/recordings/${recordingId}`, { method: "DELETE" }).then(() => {
        onClose?.();
      });
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  const themes = [
    "monokai",
    "tango",
    "solarized-dark",
    "solarized-light",
    "asciinema",
    "dracula",
    "nord",
    "gruvbox-dark",
    "gruvbox-light",
  ];

  return (
    <Card className="overflow-hidden bg-black border-gray-800">
      {/* Player Container */}
      <div
        ref={containerRef}
        className="relative w-full bg-black"
        style={{ minHeight: "400px" }}
      >
        {/* Loading State */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
        </div>
      </div>

      {/* Controls */}
      <div className="p-4 bg-gray-900 border-t border-gray-800">
        {/* Progress Bar */}
        <div className="mb-4">
          <Slider
            value={[currentTime]}
            max={duration || 100}
            step={0.1}
            className="w-full"
            onValueChange={(value: number[]) => setCurrentTime(value[0])}
          />
          <div className="flex justify-between text-xs text-gray-400 mt-1">
            <span>{formatTime(currentTime)}</span>
            <span>{formatTime(duration)}</span>
          </div>
        </div>

        {/* Control Buttons */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setCurrentTime(Math.max(0, currentTime - 10))}
            >
              <SkipBack className="h-4 w-4" />
            </Button>

            <Button
              variant="ghost"
              size="icon"
              onClick={togglePlay}
              className="bg-white text-black hover:bg-gray-200"
            >
              {isPlaying ? (
                <Pause className="h-5 w-5" />
              ) : (
                <Play className="h-5 w-5" />
              )}
            </Button>

            <Button
              variant="ghost"
              size="icon"
              onClick={() => setCurrentTime(Math.min(duration, currentTime + 10))}
            >
              <SkipForward className="h-4 w-4" />
            </Button>

            {/* Speed Control */}
            <div className="flex items-center gap-2 ml-4">
              <span className="text-xs text-gray-400">Speed:</span>
              <select
                value={speed}
                onChange={(e) => setSpeed(parseFloat(e.target.value))}
                className="bg-gray-800 text-white text-xs rounded px-2 py-1"
              >
                {[0.5, 0.75, 1.0, 1.25, 1.5, 2.0].map((s) => (
                  <option key={s} value={s}>
                    {s}x
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={handleDownload}>
              <Download className="h-4 w-4" />
            </Button>

            <Button variant="ghost" size="icon" onClick={handleShare}>
              <Share2 className="h-4 w-4" />
            </Button>

            <Button variant="ghost" size="icon" onClick={() => setShowSettings(!showSettings)}>
              <Settings className="h-4 w-4" />
            </Button>

            <Button variant="ghost" size="icon" onClick={toggleFullscreen}>
              {isFullscreen ? (
                <Minimize2 className="h-4 w-4" />
              ) : (
                <Maximize2 className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>

        {/* Settings Panel */}
        {showSettings && (
          <div className="mt-4 p-3 bg-gray-800 rounded-lg">
            <h4 className="text-sm font-medium mb-2">Theme</h4>
            <div className="grid grid-cols-3 gap-2">
              {themes.map((t) => (
                <button
                  key={t}
                  onClick={() => setTheme(t)}
                  className={`text-xs px-2 py-1 rounded focus:outline-none focus-visible:ring-2 focus-visible:ring-primary ${
                    theme === t
                      ? "bg-blue-600 text-white"
                      : "bg-gray-700 text-gray-300"
                  }`}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </Card>
  );
}
