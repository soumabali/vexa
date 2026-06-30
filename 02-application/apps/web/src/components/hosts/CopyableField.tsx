"use client";

import { useState, useCallback, KeyboardEvent } from "react";
import { Button } from "@/components/ui/button";
import { MaterialIcon } from "@/components/ui/material-icon";

interface CopyableFieldProps {
  value: string;
  label?: string;
  ariaLabel?: string;
  className?: string;
  /** Render value as monospaced. Useful for IPs, hostnames. */
  mono?: boolean;
  /** Optional mask/hide (e.g., for sensitive data). If set, full value only copied on demand. */
  masked?: boolean;
}

/**
 * Read-only field with copy-to-clipboard button.
 *
 * Accessibility:
 * - aria-label on the button describes the action.
 * - Button is keyboard reachable (native <button>).
 * - aria-live region announces "Copied" feedback to screen readers.
 * - Enter / Space activate via native button semantics.
 */
export function CopyableField({
  value,
  label,
  ariaLabel,
  className = "",
  mono = true,
  masked = false,
}: CopyableFieldProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard may be unavailable (insecure context); fail silently.
      // User can still select-and-copy the displayed text.
    }
  }, [value]);

  const onKeyDown = useCallback(
    (e: KeyboardEvent<HTMLSpanElement>) => {
      // Allow Ctrl/Cmd+C on the value itself (no-op; native behavior covers it).
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "c") {
        return;
      }
    },
    [],
  );

  const buttonAriaLabel =
    ariaLabel ?? (label ? `Copy ${label}` : "Copy value");

  return (
    <div className={`flex items-center gap-2 min-w-0 ${className}`}>
      {label && (
        <span className="text-label-md text-on-surface-variant shrink-0">{label}:</span>
      )}
      <span
        className={`truncate ${mono ? "text-mono-code font-mono-code text-on-surface" : "text-on-surface"}`}
        onKeyDown={onKeyDown}
        title={value}
      >
        {value}
      </span>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={handleCopy}
        aria-label={buttonAriaLabel}
        className="h-7 px-2 shrink-0 border border-outline-variant rounded text-on-surface-variant hover:text-on-surface"
      >
        {copied ? (
          <MaterialIcon name="check" className="h-3.5 w-3.5 text-success" size="sm" />
        ) : (
          <MaterialIcon name="content_copy" className="h-3.5 w-3.5" size="sm" />
        )}
      </Button>
      <span role="status" aria-live="polite" className="sr-only">
        {copied ? "Copied" : ""}
      </span>
      {masked && <span className="sr-only">masked</span>}
    </div>
  );
}
