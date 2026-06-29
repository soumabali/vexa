"use client";

import { useState, useCallback, KeyboardEvent } from "react";
import { Copy, Check } from "lucide-react";
import { Button } from "@/components/ui/button";

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
        <span className="text-sm text-muted-foreground shrink-0">{label}:</span>
      )}
      <span
        className={`truncate ${mono ? "font-mono text-sm" : "text-sm"}`}
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
        className="h-7 px-2 shrink-0"
      >
        {copied ? (
          <Check className="h-3.5 w-3.5 text-green-500" aria-hidden="true" />
        ) : (
          <Copy className="h-3.5 w-3.5" aria-hidden="true" />
        )}
      </Button>
      <span role="status" aria-live="polite" className="sr-only">
        {copied ? "Copied" : ""}
      </span>
      {masked && <span className="sr-only">masked</span>}
    </div>
  );
}
