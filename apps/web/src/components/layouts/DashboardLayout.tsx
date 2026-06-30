"use client";

import { usePathname } from "next/navigation";
import { TopNav } from "./TopNav";
import { Sidebar } from "./Sidebar";
import { MaterialIcon } from "@/components/ui/material-icon";

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const pathname = usePathname();

  const activeItem = pathname?.startsWith("/hosts") ? "hosts" :
    pathname?.startsWith("/terminal") ? "terminal" :
    pathname?.startsWith("/tunnels") ? "tunnels" :
    pathname?.startsWith("/vault") ? "vault" :
    pathname?.startsWith("/settings") ? "settings" :
    pathname?.startsWith("/sessions") ? "sessions" : "hosts";

  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <div className="flex pt-16 h-screen">
        <Sidebar activeItem={activeItem} />
        <main className="flex-1 md:ml-[260px] p-8 overflow-y-auto bg-background h-full">
          {children}
        </main>
      </div>
      {/* Mobile FAB */}
      <button
        type="button"
        aria-label="New connection"
        className="md:hidden fixed bottom-6 right-6 w-14 h-14 bg-primary text-on-primary rounded-full shadow-xl z-[100] flex items-center justify-center"
      >
        <MaterialIcon name="add" size="lg" />
      </button>
    </div>
  );
}