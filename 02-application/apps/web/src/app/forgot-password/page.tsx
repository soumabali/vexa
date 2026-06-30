import { AuthLayout } from "@/components/auth/AuthLayout";
import { ForgotPasswordForm } from "@/components/auth/ForgotPasswordForm";

export default function ForgotPasswordPage() {
  return (
    <AuthLayout title="Forgot Password?" description="We&apos;ll send you a reset link">
      <ForgotPasswordForm />
    </AuthLayout>
  );
}
