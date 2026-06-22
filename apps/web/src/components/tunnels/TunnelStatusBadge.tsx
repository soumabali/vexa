"use client"

import { Badge } from "@/components/ui/badge"

type TunnelStatus = "up" | "down" | "error" | "rotating" | "disabled"

interface TunnelStatusBadgeProps {
  status: TunnelStatus
}

export function TunnelStatusBadge({ status }: TunnelStatusBadgeProps) {
  const variants: Record<TunnelStatus, { variant: "default" | "secondary" | "destructive" | "outline"; label: string; className?: string }> = {
    up: { variant: "default", label: "● Active", className: "bg-success text-success-foreground" },
    down: { variant: "secondary", label: "● Inactive" },
    error: { variant: "destructive", label: "● Error" },
    rotating: { variant: "outline", label: "⟳ Rotating" },
    disabled: { variant: "secondary", label: "○ Disabled" },
  }

  const { variant, label, className } = variants[status]

  return (
    <Badge variant={variant} className={className}>
      {label}
    </Badge>
  )
}
