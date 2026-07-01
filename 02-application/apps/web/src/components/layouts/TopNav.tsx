"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { MaterialIcon } from "@/components/ui/material-icon";
import { cn } from "@/lib/utils";

interface NavItem {
  href: string;
  label: string;
  icon: string;
}

const navItems: NavItem[] = [
  { href: "/hosts", label: "Hosts", icon: "dns" },
  { href: "/terminal", label: "Terminal", icon: "terminal" },
  { href: "/tunnels", label: "Tunnels", icon: "vpn_lock" },
  { href: "/vault", label: "Vault", icon: "vpn_key" },
  { href: "/settings", label: "Settings", icon: "settings" },
];

interface TopNavProps {
  onMenuClick?: () => void;
}

export function TopNav({ onMenuClick }: TopNavProps = {}) {
  const pathname = usePathname();

  return (
    <header className="fixed top-0 left-0 right-0 h-16 bg-background border-b border-outline-variant z-50">
      <div className="flex items-center justify-between h-full px-4 md:px-6">
        {/* Left: logo + hamburger */}
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Open navigation menu"
            onClick={onMenuClick}
            className="md:hidden w-10 h-10 rounded-lg hover:bg-surface-container-highest flex items-center justify-center text-on-surface-variant transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="menu" size="md" />
          </button>
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
            <MaterialIcon name="terminal" size="sm" className="text-on-primary" fill />
          </div>
          <span className="text-headline-sm font-bold text-on-surface">Vexa</span>
        </div>

        {/* Center: nav links */}
        <nav className="hidden md:flex items-center gap-6">
          {navItems.map(({ href, label, icon }) => {
            const active = pathname?.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "flex items-center gap-2 text-sm font-medium transition-colors pb-1 border-b-2 focus:outline-none focus:ring-2 focus:ring-primary rounded",
                  active
                    ? "text-primary border-primary"
                    : "text-on-surface-variant hover:text-on-surface border-transparent",
                )}
              >
                <MaterialIcon name={icon} size="sm" fill={active} />
                <span>{label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Right: actions */}
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="Search"
            className="w-10 h-10 rounded-lg hover:bg-surface-container-highest flex items-center justify-center text-on-surface-variant transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="search" size="md" />
          </button>
          <button
            type="button"
            aria-label="Notifications"
            className="w-10 h-10 rounded-lg hover:bg-surface-container-highest flex items-center justify-center text-on-surface-variant transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="notifications" size="md" />
          </button>
          <button
            type="button"
            aria-label="Help"
            className="hidden sm:flex w-10 h-10 rounded-lg hover:bg-surface-container-highest items-center justify-center text-on-surface-variant transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="help_outline" size="md" />
          </button>
          <button
            type="button"
            aria-label="Settings"
            className="hidden sm:flex w-10 h-10 rounded-lg hover:bg-surface-container-highest items-center justify-center text-on-surface-variant transition-colors focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="settings" size="md" />
          </button>
          <button
            type="button"
            aria-label="Account"
            className="w-8 h-8 rounded-full bg-secondary-container border border-outline-variant flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <MaterialIcon name="person" size="sm" className="text-on-secondary-container" />
          </button>
        </div>
      </div>
    </header>
  );
}