import { MaterialIcon } from "@/components/ui/material-icon";

interface AuthLayoutProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
}

const features = [
  { icon: "dns", label: "Manage SSH hosts & cloud instances" },
  { icon: "terminal", label: "Instant browser terminal access" },
  { icon: "vpn_lock", label: "Secure tunnels & key management" },
];

export function AuthLayout({ children, title, description }: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex">
      {/* Left panel */}
      <div className="hidden md:flex w-1/2 bg-surface-container flex-col justify-center p-12">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
            <MaterialIcon name="terminal" size="md" className="text-on-primary" fill />
          </div>
          <span className="text-headline-lg font-bold text-on-surface">Vexa</span>
        </div>
        <p className="text-body-lg text-on-surface-variant max-w-md mb-10">
          Secure, terminal-first infrastructure access from your browser.
        </p>
        <ul className="space-y-4 max-w-md">
          {features.map(({ icon, label }) => (
            <li key={icon} className="flex items-center gap-3">
              <MaterialIcon name={icon} size="md" className="text-primary" />
              <span className="text-body-md text-on-surface">{label}</span>
            </li>
          ))}
        </ul>
        <div className="mt-auto pt-12 text-on-surface-variant text-label-md">
          v1.0.0
        </div>
      </div>

      {/* Right panel */}
      <div className="flex-1 bg-background flex items-center justify-center p-8">
        <div className="w-full max-w-md bg-surface-container border border-outline-variant rounded-xl p-8">
          {(title || description) && (
            <div className="space-y-2 mb-6">
              {title && <h1 className="text-headline-sm font-bold text-on-surface">{title}</h1>}
              {description && <p className="text-on-surface-variant text-body-md">{description}</p>}
            </div>
          )}
          {children}
        </div>
      </div>
    </div>
  );
}