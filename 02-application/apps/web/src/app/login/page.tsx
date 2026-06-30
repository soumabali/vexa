import { AuthLayout } from "@/components/auth/AuthLayout";
import { LoginForm } from "@/components/auth/LoginForm";

export default function LoginPage() {
  return (
    <AuthLayout title="Welcome Back" description="Sign in to your vexa account">
      <LoginForm />
    </AuthLayout>
  );
}
