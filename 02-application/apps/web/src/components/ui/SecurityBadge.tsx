import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface SecurityBadgeProps {
  status: "success" | "warning" | "danger" | "info" | "mfa";
  children: ReactNode;
  className?: string;
}

export function SecurityBadge({ status, children, className }: SecurityBadgeProps) {
  // Semantic colors from design system
  const colorMap = {
    success: 'bg-success/10 text-success',
    warning: 'bg-warning/10 text-warning',
    danger: 'bg-error/10 text-error',
    info: 'bg-primary/10 text-primary',
    mfa: 'bg-primary-container text-on-primary-container',
  };
  const bg = colorMap[status] || colorMap.success;
  const classes = `px-2 py-1 rounded text-xs font-medium whitespace-nowrap ${bg} ${className || ''}`;
  return <span className={classes}>{children}</span>;
}
