"use client";

import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { HostList } from "@/components/hosts/HostList";

export default function HostsPage() {
  return (
    <DashboardLayout>
      <HostList />
    </DashboardLayout>
  );
}
