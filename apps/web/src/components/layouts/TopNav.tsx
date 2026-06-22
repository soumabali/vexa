"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Server, Globe, Shield, Settings, TerminalSquare } from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
  { href: "/hosts", label: "Hosts", icon: Server },
  { href: "/terminal", label: "Terminal", icon: TerminalSquare },
  { href: "/tunnels", label: "Tunnels", icon: Globe },
  { href: "/vault", label: "Vault", icon: Shield },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function TopNav() {
  const pathname = usePathname();

  return (
    <header className="border-b border-border bg-surface-container">
      <div className="container flex items-center justify-between h-14">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
            <svg className="w-5 h-5 text-primary-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <span className="font-bold text-lg tracking-tight">vexa</span>
        </div>
        <nav className="flex items-center gap-6">
          {navItems.map(({ href, label, icon: Icon }) => (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-2 text-sm font-medium transition-colors",
                pathname?.startsWith(href)
                  ? "text-primary hover:text-primary/80"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <Icon className="h-4 w-4" />
              {label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
