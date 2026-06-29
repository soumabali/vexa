"use client";

import React from "react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import {
  User,
  Shield,
  Key,
  Palette,
  Bell,
  ChevronRight,
  Globe,
  Keyboard,
  Fingerprint,
  Monitor,
} from "lucide-react";

const settingsSections = [
  {
    title: "Profile",
    description: "Manage your profile information and preferences",
    icon: User,
    href: "/settings/profile",
  },
  {
    title: "Security",
    description: "Two-factor authentication, WebAuthn, and security settings",
    icon: Shield,
    href: "/settings/security",
  },
  {
    title: "Active Sessions",
    description: "Review and revoke devices signed in to your account",
    icon: Monitor,
    href: "/settings/sessions",
  },
  {
    title: "WebAuthn",
    description: "Manage passkeys and security keys",
    icon: Fingerprint,
    href: "/settings/webauthn",
  },
  {
    title: "API Keys",
    description: "Generate and manage API keys for programmatic access",
    icon: Key,
    href: "/settings/api-keys",
  },
  {
    title: "Appearance",
    description: "Theme, font, and visual preferences",
    icon: Palette,
    href: "/settings/appearance",
  },
  {
    title: "Preferences",
    description: "Terminal preferences, notifications, and behavior",
    icon: Bell,
    href: "/settings/preferences",
  },
  {
    title: "Notifications",
    description: "Email and in-app notification settings",
    icon: Bell,
    href: "/settings/notifications",
  },
  {
    title: "Keyboard",
    description: "Keyboard shortcuts and bindings",
    icon: Keyboard,
    href: "/settings/keyboard",
  },
  {
    title: "Language & Region",
    description: "Language, timezone, and date format",
    icon: Globe,
    href: "/settings/language",
  },
];

export default function SettingsPage() {
  return (
    <div className="container mx-auto py-8 px-4 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground mt-2">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="grid gap-4">
        {settingsSections.map((section) => (
          <Link key={section.href} href={section.href}>
            <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="p-3 bg-primary/10 rounded-lg">
                    <section.icon className="h-6 w-6 text-primary" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold">{section.title}</h3>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {section.description}
                    </p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
