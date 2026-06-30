"use client";

import { Toaster as Sonner } from "sonner";

type ToasterProps = React.ComponentProps<typeof Sonner>;

export function Toaster({ ...props }: ToasterProps) {
  return (
    <Sonner
      toastOptions={{
        classNames: {
          toast:
            "border border-outline-variant bg-surface-container-highest text-on-surface",
          description: "text-on-surface-variant",
          actionButton: "bg-surface-variant text-on-surface",
          cancelButton: "bg-surface-variant text-on-surface-variant",
        },
      }}
      {...props}
    />
  );
}
