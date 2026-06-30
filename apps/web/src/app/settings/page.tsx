"use client";

import React from "react";
import Link from "next/link";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { MaterialIcon } from "@/components/ui/material-icon";

interface SettingsSection {
  title: string;
  description: string;
  icon: string;
  href: string;
}

interface SettingsGroup {
  label: string;
  sections: SettingsSection[];
}

const settingsGroups: SettingsGroup[] = [
  {
    label: "Core Account",
    sections: [
      {
        title: "Profile",
        description: "Update identity, email, and public-facing details.",
        icon: "person",
        href: "/settings/profile",
      },
      {
        title: "Security",
        description: "Multi-factor authentication, login history, password audits.",
        icon: "security",
        href: "/settings/security",
      },
    ],
  },
  {
    label: "Developer Access",
    sections: [
      {
        title: "API Keys",
        description: "Manage programmatic access and third-party integrations.",
        icon: "key",
        href: "/settings/api-keys",
      },
      {
        title: "WebAuthn",
        description: "Register security keys and passwordless authentication.",
        icon: "verified_user",
        href: "/settings/webauthn",
      },
    ],
  },
  {
    label: "System",
    sections: [
      {
        title: "Appearance",
        description: "Customize theme, density, and typography preferences.",
        icon: "palette",
        href: "/settings/appearance",
      },
      {
        title: "Notifications",
        description: "Configure alerts, channels, and event triggers.",
        icon: "notifications",
        href: "/settings/notifications",
      },
      {
        title: "Sessions",
        description: "View and manage active and past sessions.",
        icon: "history",
        href: "/settings/sessions",
      },
    ],
  },
];

export default function SettingsPage() {
  return (
    <DashboardLayout>
      <div className="container mx-auto py-8 px-4 max-w-4xl">
        <div className="mb-10">
          <h1 className="text-headline-lg text-on-surface mb-2">Settings</h1>
          <p className="text-body-lg text-on-surface-variant">
            Manage account configurations, security preferences, and system behavior.
          </p>
        </div>

        <div className="space-y-10">
          {settingsGroups.map((group) => (
            <section key={group.label}>
              <h2 className="text-label-md text-primary mb-3 uppercase tracking-[0.2em]">
                {group.label}
              </h2>
              <div className="space-y-4">
                {group.sections.map((section) => (
                  <Link
                    key={section.href}
                    href={section.href}
                    className="group flex items-center justify-between p-5 rounded-xl bg-surface-container border border-outline-variant hover:border-primary hover:-translate-y-px transition-all duration-300"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 flex items-center justify-center rounded-xl bg-surface-container-high text-primary group-hover:bg-primary group-hover:text-on-primary transition-colors">
                        <MaterialIcon name={section.icon} className="text-[28px]" />
                      </div>
                      <div>
                        <h3 className="text-body-lg font-bold text-on-surface">
                          {section.title}
                        </h3>
                        <p className="text-body-md text-on-surface-variant">
                          {section.description}
                        </p>
                      </div>
                    </div>
                    <MaterialIcon
                      name="chevron_right"
                      className="text-on-surface-variant group-hover:translate-x-1 transition-transform"
                    />
                  </Link>
                ))}
              </div>
            </section>
          ))}
        </div>
      </div>
    </DashboardLayout>
  );
}