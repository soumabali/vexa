"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import { MaterialIcon } from "@/components/ui/material-icon";
import { toast } from "sonner";

type ThemeOption = "dark" | "light" | "system";
type DensityOption = "comfortable" | "compact";

const THEMES: { id: ThemeOption; label: string; icon: string; desc: string }[] = [
  { id: "dark", label: "Dark", icon: "dark_mode", desc: "Always dark" },
  { id: "light", label: "Light", icon: "light_mode", desc: "Always light" },
  { id: "system", label: "System", icon: "desktop_windows", desc: "Follow OS" },
];

const ACCENT_COLORS: { id: string; label: string; class: string }[] = [
  { id: "indigo", label: "Indigo", class: "bg-primary" },
  { id: "blue", label: "Blue", class: "bg-blue-500" },
  { id: "teal", label: "Teal", class: "bg-teal-500" },
  { id: "green", label: "Green", class: "bg-success" },
  { id: "amber", label: "Amber", class: "bg-amber-500" },
  { id: "rose", label: "Rose", class: "bg-rose-500" },
];

export default function AppearanceSettingsPage() {
  const [theme, setTheme] = useState<ThemeOption>("dark");
  const [density, setDensity] = useState<DensityOption>("comfortable");
  const [fontSize, setFontSize] = useState<number>(14);
  const [accent, setAccent] = useState<string>("indigo");

  const handleSave = () => {
    toast.success("Appearance updated", {
      icon: <MaterialIcon name="check" size="sm" className="text-primary" />,
    });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Appearance</h1>
        <p className="text-sm text-on-surface-variant">
          Customize how Vexa looks. Changes apply to this device only.
        </p>
      </div>

      {/* Theme */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="palette" size="sm" className="text-on-surface-variant" />
            Theme
          </CardTitle>
          <CardDescription>Choose a color scheme for the interface.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            {THEMES.map((option) => {
              const active = theme === option.id;
              return (
                <button
                  key={option.id}
                  type="button"
                  onClick={() => setTheme(option.id)}
                  className={`relative flex flex-col items-start gap-2 rounded-lg p-4 text-left transition-colors ${
                    active
                      ? "border-2 border-primary bg-primary/5"
                      : "border border-outline-variant hover:border-outline"
                  }`}
                >
                  <div
                    className={`flex h-10 w-10 items-center justify-center rounded-lg ${
                      active ? "bg-primary-container text-on-primary-container" : "bg-surface-container-high text-on-surface-variant"
                    }`}
                  >
                    <MaterialIcon name={option.icon} size="md" />
                  </div>
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-on-surface">{option.label}</p>
                    <p className="text-xs text-on-surface-variant">{option.desc}</p>
                  </div>
                  {active && (
                    <span className="absolute right-3 top-3 flex h-5 w-5 items-center justify-center rounded-full bg-primary text-on-primary">
                      <MaterialIcon name="check" size="sm" />
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Density */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="density_medium" size="sm" className="text-on-surface-variant" />
            Density
          </CardTitle>
          <CardDescription>Control the spacing between elements.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {(
              [
                { id: "comfortable", label: "Comfortable", desc: "Relaxed spacing", icon: "format_indent_increase" },
                { id: "compact", label: "Compact", desc: "Tighter rows", icon: "compress" },
              ] as { id: DensityOption; label: string; desc: string; icon: string }[]
            ).map((option) => {
              const selected = density === option.id;
              return (
                <button
                  key={option.id}
                  type="button"
                  onClick={() => setDensity(option.id)}
                  className={`flex items-center gap-3 rounded-lg p-4 text-left transition-colors ${
                    selected
                      ? "bg-secondary-container text-on-secondary-container"
                      : "border border-outline-variant hover:border-outline text-on-surface"
                  }`}
                >
                  <div
                    className={`flex h-9 w-9 items-center justify-center rounded-full border-2 ${
                      selected ? "border-on-secondary-container" : "border-outline-variant"
                    }`}
                  >
                    {selected && (
                      <span className="h-4 w-4 rounded-full bg-on-secondary-container" />
                    )}
                  </div>
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium">{option.label}</p>
                    <p className={`text-xs ${selected ? "text-on-secondary-container/80" : "text-on-surface-variant"}`}>
                      {option.desc}
                    </p>
                  </div>
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Font size */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="text_fields" size="sm" className="text-on-surface-variant" />
            Font Size
          </CardTitle>
          <CardDescription>Adjust the base text size for the interface.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <MaterialIcon name="format_size" size="sm" className="text-on-surface-variant" />
            <Slider
              value={[fontSize]}
              min={12}
              max={18}
              step={1}
              onValueChange={(value) => setFontSize(value[0])}
              className="flex-1"
            />
            <span className="w-12 text-right text-sm font-medium tabular-nums text-on-surface">
              {fontSize}px
            </span>
          </div>
          <div className="rounded-lg border border-outline-variant bg-surface-container-lowest p-4">
            <p
              className="text-on-surface"
              style={{ fontSize: `${fontSize}px`, lineHeight: 1.5 }}
            >
              The quick brown fox jumps over the lazy dog.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Accent color */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="colorize" size="sm" className="text-on-surface-variant" />
            Accent Color
          </CardTitle>
          <CardDescription>Pick the primary accent used for buttons and highlights.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            {ACCENT_COLORS.map((color) => {
              const active = accent === color.id;
              return (
                <button
                  key={color.id}
                  type="button"
                  onClick={() => setAccent(color.id)}
                  aria-label={color.label}
                  aria-pressed={active}
                  className={`flex h-9 w-9 items-center justify-center rounded-full ring-2 ring-offset-2 ring-offset-surface-container transition-all ${
                    color.class
                  } ${active ? "ring-on-surface" : "ring-transparent hover:ring-outline-variant"}`}
                >
                  {active && <MaterialIcon name="check" size="sm" className="text-on-primary" />}
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={handleSave}>Save Changes</Button>
      </div>
    </div>
  );
}