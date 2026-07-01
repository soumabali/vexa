import { AuthLayout } from "@/components/auth/AuthLayout";
import { ResetPasswordForm } from "@/components/auth/ResetPasswordForm";

export default function ResetPasswordPage() {
  return (
    <AuthLayout title="Reset Password" description="Choose a new password for your account">
      <ResetPasswordForm />
    </AuthLayout>
  );
}
