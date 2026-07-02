"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { MaterialIcon } from "@/components/ui/material-icon";
import { cn } from "@/lib/utils";

interface SidebarProps {
  activeItem?: "hosts" | "terminal" | "tunnels" | "vault" | "settings" | "sessions";
  mobileOpen?: boolean;
  onClose?: () => void;
}

interface NavItem {
  href: string;
  label: string;
  icon: string;
  key: NonNullable<SidebarProps["activeItem"]>;
  primary?: boolean;
}

const navItems: NavItem[] = [
  { href: "/hosts", label: "All Hosts", icon: "dns", key: "hosts", primary: true },
  { href: "/groups", label: "Groups", icon: "folder", key: "hosts" },
  { href: "/instances", label: "Cloud Instances", icon: "cloud", key: "hosts" },
  { href: "/sessions", label: "Recent Sessions", icon: "history", key: "sessions", primary: true },
  { href: "/vault", label: "Key Management", icon: "vpn_key", key: "vault", primary: true },
];

const settingsNavItems: { href: string; label: string; icon: string }[] = [
  { href: "/settings/profile", label: "Profile", icon: "person" },
  { href: "/settings/security", label: "Security", icon: "shield" },
  { href: "/settings/sessions", label: "Active Sessions", icon: "devices" },
  { href: "/settings/webauthn", label: "WebAuthn", icon: "fingerprint" },
  { href: "/settings/api-keys", label: "API Keys", icon: "vpn_key" },
  { href: "/settings/appearance", label: "Appearance", icon: "palette" },
  { href: "/settings/notifications", label: "Notifications", icon: "notifications" },
];

export function Sidebar({ activeItem = "hosts", mobileOpen = false, onClose }: SidebarProps) {
  const pathname = usePathname();
  const inSettings = pathname?.startsWith("/settings") ?? false;

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden md:flex flex-col w-[260px] fixed left-0 top-16 bottom-0 bg-surface-container border-r border-outline-variant z-40">
        <SidebarContent activeItem={activeItem} inSettings={inSettings} />
      </aside>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="md:hidden fixed inset-0 z-[90]">
          {/* Scrim */}
          <button
            type="button"
            aria-label="Close navigation"
            onClick={onClose}
            className="absolute inset-0 bg-black/50"
          />
          {/* Drawer panel */}
          <aside className="absolute left-0 top-0 bottom-0 w-[280px] bg-surface-container border-r border-outline-variant z-50 flex flex-col animate-[slideIn_0.2s_ease-out]">
            <div className="flex items-center justify-between px-4 py-4 border-b border-outline-variant">
              <span className="text-on-surface font-semibold">Navigation</span>
              <button
                type="button"
                aria-label="Close navigation"
                onClick={onClose}
                className="w-10 h-10 rounded-lg hover:bg-surface-container-highest flex items-center justify-center text-on-surface-variant focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <MaterialIcon name="close" size="md" />
              </button>
            </div>
            <div onClick={onClose} className="flex-1 flex flex-col">
              <SidebarContent activeItem={activeItem} inSettings={inSettings} />
            </div>
          </aside>
        </div>
      )}
    </>
  );
}

function SidebarContent({
  activeItem,
  inSettings,
}: {
  activeItem: NonNullable<SidebarProps["activeItem"]>;
  inSettings: boolean;
}) {
  const pathname = usePathname();
  return (
    <>
      {/* User context */}
      <div className="flex items-center gap-3 px-4 py-4 border-b border-outline-variant">
        <div className="w-9 h-9 rounded-full bg-secondary-container flex items-center justify-center">
          <MaterialIcon name="person" size="sm" className="text-on-secondary-container" />
        </div>
        <div className="flex flex-col min-w-0">
          <span className="text-on-surface text-sm font-semibold truncate">demo-user</span>
          <span className="text-on-surface-variant text-xs truncate">local-cluster</span>
        </div>
      </div>

      {/* New connection button */}
      <div className="px-4 py-3">
        <button
          type="button"
          className="w-full flex items-center justify-center gap-2 bg-primary-container text-on-primary-container rounded-lg px-3 py-2 text-sm font-medium hover:opacity-90 transition-opacity focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <MaterialIcon name="add" size="sm" />
          <span>New Connection</span>
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-2 space-y-1">
        {inSettings
          ? settingsNavItems.map(({ href, label, icon }) => {
              const active = pathname === href;
              return (
                <Link
                  key={href}
                  href={href}
                  className={cn(
                    "flex items-center gap-3 px-3 py-2 font-label-md text-label-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary rounded-lg",
                    active
                      ? "text-primary bg-secondary-container border-l-2 border-primary rounded-lg"
                      : "text-on-surface-variant hover:bg-surface-variant rounded-lg",
                  )}
                >
                  <MaterialIcon name={icon} size="sm" fill={active} />
                  <span>{label}</span>
                </Link>
              );
            })
          : navItems.map(({ href, label, icon, key, primary }) => {
              const active = activeItem === key && primary === true;
              return (
                <Link
                  key={`${href}-${label}`}
                  href={href}
                  className={cn(
                    "flex items-center gap-3 px-3 py-2 font-label-md text-label-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary rounded-lg",
                    active
                      ? "text-primary bg-secondary-container border-l-2 border-primary rounded-lg"
                      : "text-on-surface-variant hover:bg-surface-variant rounded-lg",
                  )}
                >
                  <MaterialIcon name={icon} size="sm" fill={active} />
                  <span>{label}</span>
                </Link>
              );
            })}
      </nav>

      {/* Footer */}
      <div className="border-t border-outline-variant pt-4 px-3 pb-4 space-y-1">
        <Link
          href="/help"
          className="flex items-center gap-3 px-3 py-2 text-on-surface-variant hover:bg-surface-variant rounded-lg font-label-md text-label-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <MaterialIcon name="help" size="sm" />
          <span>Help Docs</span>
        </Link>
        {!inSettings && (
          <Link
            href="/settings"
            className="flex items-center gap-3 px-3 py-2 text-on-surface-variant hover:bg-surface-variant rounded-lg font-label-md text-label-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="settings" size="sm" />
            <span>Settings</span>
          </Link>
        )}
        <Link
          href="/login"
          className="flex items-center gap-3 px-3 py-2 text-error hover:bg-surface-variant rounded-lg font-label-md text-label-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <MaterialIcon name="logout" size="sm" />
          <span>Log out</span>
        </Link>
      </div>
    </>
  );
}