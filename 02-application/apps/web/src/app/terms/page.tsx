import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { MaterialIcon } from "@/components/ui/material-icon";

export default function TermsPage() {
  return (
    <DashboardLayout>
      <div className="container mx-auto max-w-3xl px-4 py-12">
        <div className="mb-8 flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-primary-container flex items-center justify-center">
            <MaterialIcon name="description" size="md" className="text-on-primary-container" />
          </div>
          <h1 className="text-headline-lg font-bold text-on-surface">Terms of Service</h1>
        </div>

        <div className="bg-surface-container rounded-xl border border-outline-variant p-8 space-y-6">
          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">1. Acceptance of Terms</h2>
            <p className="text-body-md text-on-surface-variant">
              By accessing and using vexa, you accept and agree to be bound by the terms and
              provisions of this agreement.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">2. Description of Service</h2>
            <p className="text-body-md text-on-surface-variant">
              vexa provides tools for managing SSH connections, keys, and configurations.
              The service is provided &quot;as is&quot; without any warranties regarding availability
              or performance.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">3. User Accounts</h2>
            <p className="text-body-md text-on-surface-variant">
              You are responsible for maintaining the confidentiality of your account credentials
              and for all activities that occur under your account.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">4. Security</h2>
            <p className="text-body-md text-on-surface-variant">
              You agree to use strong passwords and enable two-factor authentication when available.
              We are not responsible for unauthorized access to your account due to weak credentials.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">5. Acceptable Use</h2>
            <p className="text-body-md text-on-surface-variant">
              You agree not to use vexa for any illegal or unauthorized purpose. You must comply with
              all applicable laws and regulations when using this service.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">6. Limitation of Liability</h2>
            <p className="text-body-md text-on-surface-variant">
              vexa and its contributors shall not be liable for any indirect, incidental, or
              consequential damages arising from the use of this service.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-title-lg font-semibold text-on-surface">7. Changes to Terms</h2>
            <p className="text-body-md text-on-surface-variant">
              We reserve the right to modify these terms at any time. Continued use of vexa after
              changes constitutes acceptance of the updated terms.
            </p>
          </section>
        </div>
      </div>
    </DashboardLayout>
  );
}