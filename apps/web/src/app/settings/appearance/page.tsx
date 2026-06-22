"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Separator } from "@/components/ui/separator";
import { Slider } from "@/components/ui/slider";
import {
  Palette,
  Monitor,
  Moon,
  Sun,
  Type,
  Layout,
  Grid3X3,
  List,
} from "lucide-react";

export default function AppearanceSettingsPage() {
  const [theme, setTheme] = useState<"light" | "dark" | "system">("system");
  const [fontSize, setFontSize] = useState(14);
  const [fontFamily, setFontFamily] = useState("monospace");
  const [showLineNumbers, setShowLineNumbers] = useState(true);
  const [wordWrap, setWordWrap] = useState(true);
  const [compactMode, setCompactMode] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [showGitBranch, setShowGitBranch] = useState(true);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  const themes = [
    { value: "light", label: "Light", icon: Sun, description: "Clean and bright" },
    { value: "dark", label: "Dark", icon: Moon, description: "Easy on the eyes" },
    { value: "system", label: "System", icon: Monitor, description: "Follow system preference" },
  ];

  const fontFamilies = [
    { value: "monospace", label: "Monospace" },
    { value: "jetbrains-mono", label: "JetBrains Mono" },
    { value: "fira-code", label: "Fira Code" },
    { value: "source-code-pro", label: "Source Code Pro" },
    { value: "cascadia-code", label: "Cascadia Code" },
  ];

  return (
    <div className="container mx-auto py-8 px-4 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Appearance</h1>
        <p className="text-muted-foreground mt-2">
          Customize the look and feel of the application
        </p>
      </div>

      {/* Theme */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Palette className="h-5 w-5" />
            Theme
          </CardTitle>
          <CardDescription>Choose your preferred color scheme</CardDescription>
        </CardHeader>
        <CardContent>
          <RadioGroup
            value={theme}
            onValueChange={(v) => setTheme(v as "light" | "dark" | "system")}
            className="grid grid-cols-3 gap-4"
          >
            {themes.map((t) => (
              <div key={t.value}>
                <RadioGroupItem
                  value={t.value}
                  id={t.value}
                  className="peer sr-only"
                />
                <label
                  htmlFor={t.value}
                  className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
                >
                  <t.icon className="h-6 w-6 mb-2" />
                  <span className="font-medium">{t.label}</span>
                  <span className="text-xs text-muted-foreground">
                    {t.description}
                  </span>
                </label>
              </div>
            ))}
          </RadioGroup>
        </CardContent>
      </Card>

      {/* Font */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Type className="h-5 w-5" />
            Font
          </CardTitle>
          <CardDescription>Configure font settings for the terminal and UI</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label>Font Family</Label>
            <div className="grid grid-cols-2 gap-2">
              {fontFamilies.map((font) => (
                <Button
                  key={font.value}
                  variant={fontFamily === font.value ? "default" : "outline"}
                  className="justify-start"
                  onClick={() => setFontFamily(font.value)}
                >
                  <span style={{ fontFamily: font.value }}>Aa</span>
                  <span className="ml-2">{font.label}</span>
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>Font Size</Label>
              <span className="text-sm text-muted-foreground">{fontSize}px</span>
            </div>
            <Slider
              value={[fontSize]}
              onValueChange={(v) => {
                const val = Array.isArray(v) ? v[0] : v;
                setFontSize(typeof val === 'number' ? val : parseInt(val as string));
              }}
              min={10}
              max={20}
              step={1}
            />
          </div>
        </CardContent>
      </Card>

      {/* Terminal */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Layout className="h-5 w-5" />
            Terminal
          </CardTitle>
          <CardDescription>Terminal display preferences</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Show Line Numbers</Label>
              <p className="text-sm text-muted-foreground">
                Display line numbers in the terminal
              </p>
            </div>
            <Switch
              checked={showLineNumbers}
              onCheckedChange={setShowLineNumbers}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Word Wrap</Label>
              <p className="text-sm text-muted-foreground">
                Wrap long lines to fit the terminal width
              </p>
            </div>
            <Switch
              checked={wordWrap}
              onCheckedChange={setWordWrap}
            />
          </div>
        </CardContent>
      </Card>

      {/* Layout */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Layout className="h-5 w-5" />
            Layout
          </CardTitle>
          <CardDescription>Configure the application layout</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Compact Mode</Label>
              <p className="text-sm text-muted-foreground">
                Reduce padding and spacing for a denser UI
              </p>
            </div>
            <Switch
              checked={compactMode}
              onCheckedChange={setCompactMode}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Collapsed Sidebar</Label>
              <p className="text-sm text-muted-foreground">
                Start with the sidebar collapsed
              </p>
            </div>
            <Switch
              checked={sidebarCollapsed}
              onCheckedChange={setSidebarCollapsed}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Show Git Branch</Label>
              <p className="text-sm text-muted-foreground">
                Display current git branch in the status bar
              </p>
            </div>
            <Switch
              checked={showGitBranch}
              onCheckedChange={setShowGitBranch}
            />
          </div>
        </CardContent>
      </Card>

      {/* Default View Mode */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Default View Mode</CardTitle>
          <CardDescription>Choose your preferred default view</CardDescription>
        </CardHeader>
        <CardContent>
          <RadioGroup
            value={viewMode}
            onValueChange={(v) => setViewMode(v as "grid" | "list")}
            className="grid grid-cols-2 gap-4"
          >
            <div>
              <RadioGroupItem
                value="grid"
                id="grid"
                className="peer sr-only"
              />
              <label
                htmlFor="grid"
                className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
              >
                <Grid3X3 className="h-6 w-6 mb-2" />
                <span className="font-medium">Grid</span>
                <span className="text-xs text-muted-foreground">Card-based layout</span>
              </label>
            </div>
            <div>
              <RadioGroupItem
                value="list"
                id="list"
                className="peer sr-only"
              />
              <label
                htmlFor="list"
                className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
              >
                <List className="h-6 w-6 mb-2" />
                <span className="font-medium">List</span>
                <span className="text-xs text-muted-foreground">Compact rows</span>
              </label>
            </div>
          </RadioGroup>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button>Save Changes</Button>
      </div>
    </div>
  );
}
