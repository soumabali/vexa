import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function TermsPage() {
  return (
    <div className="container mx-auto max-w-3xl px-4 py-12">
      <Card>
        <CardHeader>
          <CardTitle className="text-3xl">Terms of Service</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <section className="space-y-3">
            <h2 className="text-xl font-semibold">1. Acceptance of Terms</h2>
            <p className="text-muted-foreground">
              By accessing and using vexa, you accept and agree to be bound by the terms
              and provisions of this agreement.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-xl font-semibold">2. Description of Service</h2>
            <p className="text-muted-foreground">
              vexa provides tools for managing SSH connections, keys, and configurations.
              The service is provided &quot;as is&quot; and we make no warranties regarding its availability or performance.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-xl font-semibold">3. User Accounts</h2>
            <p className="text-muted-foreground">
              You are responsible for maintaining the confidentiality of your account credentials
              and for all activities that occur under your account.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-xl font-semibold">4. Security</h2>
            <p className="text-muted-foreground">
              You agree to use strong passwords and enable two-factor authentication when available.
              We are not responsible for unauthorized access to your account due to weak credentials.
            </p>
          </section>

          <section className="space-y-3">
            <h2 className="text-xl font-semibold">5. Limitation of Liability</h2>
            <p className="text-muted-foreground">
              In no event shall vexa be liable for any indirect, incidental, special,
              consequential, or punitive damages.
            </p>
          </section>
        </CardContent>
      </Card>
    </div>
  );
}
