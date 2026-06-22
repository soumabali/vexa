'use client'

import { useState } from 'react'
import { X, Monitor, Keyboard, Palette, Shield, Bell } from 'lucide-react'

interface SettingsPanelProps {
  onClose: () => void
}

type SettingsTab = 'general' | 'appearance' | 'keyboard' | 'security' | 'notifications'

export function SettingsPanel({ onClose }: SettingsPanelProps) {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general')

  const tabs: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
    { id: 'general', label: 'General', icon: <Monitor className="w-4 h-4" /> },
    { id: 'appearance', label: 'Appearance', icon: <Palette className="w-4 h-4" /> },
    { id: 'keyboard', label: 'Keyboard', icon: <Keyboard className="w-4 h-4" /> },
    { id: 'security', label: 'Security', icon: <Shield className="w-4 h-4" /> },
    { id: 'notifications', label: 'Notifications', icon: <Bell className="w-4 h-4" /> },
  ]

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border border-border rounded-lg w-[600px] max-w-[90vw] max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold">Settings</h2>
          <button
            onClick={onClose}
            className="p-1.5 hover:bg-secondary rounded-md transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex flex-1 overflow-hidden">
          {/* Sidebar */}
          <div className="w-48 border-r border-border p-2 space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 w-full px-3 py-2 text-sm rounded-md transition-colors ${
                  activeTab === tab.id
                    ? 'bg-secondary text-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-secondary/50'
                }`}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 p-4 overflow-y-auto">
            {activeTab === 'general' && <GeneralSettings />}
            {activeTab === 'appearance' && <AppearanceSettings />}
            {activeTab === 'keyboard' && <KeyboardSettings />}
            {activeTab === 'security' && <SecuritySettings />}
            {activeTab === 'notifications' && <NotificationSettings />}
          </div>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 p-4 border-t border-border">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              // TODO: Save settings
              onClose()
            }}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  )
}

function GeneralSettings() {
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-2">Startup</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" />
            <span className="text-sm">Start on system startup</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" />
            <span className="text-sm">Restore last session on startup</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Minimize to system tray on close</span>
          </label>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium mb-2">Behavior</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Confirm before closing active connections</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Auto-reconnect on connection drop</span>
          </label>
        </div>
      </div>
    </div>
  )
}

function AppearanceSettings() {
  const [theme, setTheme] = useState('dark')
  const [fontSize, setFontSize] = useState(14)

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-2">Theme</h3>
        <div className="flex gap-2">
          {['light', 'dark', 'system'].map((t) => (
            <button
              key={t}
              onClick={() => setTheme(t)}
              className={`px-4 py-2 text-sm rounded-md border transition-colors ${
                theme === t
                  ? 'border-primary bg-primary/10'
                  : 'border-border hover:bg-secondary'
              }`}
            >
              {t.charAt(0).toUpperCase() + t.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium mb-2">Font Size</h3>
        <div className="flex items-center gap-4">
          <input
            type="range"
            min={10}
            max={24}
            value={fontSize}
            onChange={(e) => setFontSize(Number(e.target.value))}
            className="flex-1"
          />
          <span className="text-sm w-12">{fontSize}px</span>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium mb-2">Terminal Colors</h3>
        <div className="grid grid-cols-8 gap-2">
          {['#1a1a1a', '#ff5f56', '#27c93f', '#ffbd2e', '#007aff', '#ff2d55', '#5ac8fa', '#e0e0e0',
            '#3a3a3a', '#ff6b6b', '#32d74b', '#ffd60a', '#0a84ff', '#ff375f', '#64d2ff', '#ffffff'].map((color) => (
            <div
              key={color}
              className="w-8 h-8 rounded border border-border cursor-pointer hover:ring-2 ring-primary"
              style={{ backgroundColor: color }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

function KeyboardSettings() {
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-2">Global Shortcuts</h3>
        <div className="space-y-2">
          {[
            { label: 'Quick Connect', shortcut: 'Ctrl+Shift+T' },
            { label: 'Toggle Window', shortcut: 'Ctrl+Shift+H' },
            { label: 'New Tab', shortcut: 'Ctrl+T' },
            { label: 'Close Tab', shortcut: 'Ctrl+W' },
            { label: 'Copy', shortcut: 'Ctrl+Shift+C' },
            { label: 'Paste', shortcut: 'Ctrl+Shift+V' },
            { label: 'Find', shortcut: 'Ctrl+F' },
          ].map((item) => (
            <div
              key={item.label}
              className="flex items-center justify-between py-2 border-b border-border last:border-0"
            >
              <span className="text-sm">{item.label}</span>
              <kbd className="px-2 py-1 text-xs bg-secondary border border-border rounded">
                {item.shortcut}
              </kbd>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function SecuritySettings() {
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-2">Clipboard</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Auto-clear clipboard after copy (passwords)</span>
          </label>
          <div className="flex items-center gap-2 ml-6">
            <span className="text-sm text-muted-foreground">Clear after:</span>
            <select className="bg-secondary border border-border rounded text-sm px-2 py-1">
              <option>10 seconds</option>
              <option>30 seconds</option>
              <option>1 minute</option>
              <option>5 minutes</option>
            </select>
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium mb-2">Session</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Encrypt window state at rest</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Lock app after inactivity</span>
          </label>
        </div>
      </div>
    </div>
  )
}

function NotificationSettings() {
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-medium mb-2">Notifications</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Show connection status notifications</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" defaultChecked />
            <span className="text-sm">Show error notifications</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="rounded" />
            <span className="text-sm">Show update notifications</span>
          </label>
        </div>
      </div>
    </div>
  )
}
