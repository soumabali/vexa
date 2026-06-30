"use client";

import { Suspense, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { AuthLayout } from "@/components/auth/AuthLayout";
import { MaterialIcon } from "@/components/ui/material-icon";
import Link from "next/link";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const [isVerifying, setIsVerifying] = useState(false);
  const [isVerified, setIsVerified] = useState(false);

  const handleVerify = async () => {
    setIsVerifying(true);
    // In real app, verify token with API
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setIsVerified(true);
    setIsVerifying(false);
  };

  if (isVerified) {
    return (
      <AuthLayout title="Email Verified!" description="You can now log in to your account">
        <div className="flex flex-col items-center text-center space-y-6">
          <MaterialIcon name="mark_email_read" size="xl" className="text-primary" />
          <p className="text-on-surface-variant text-body-md">
            Your email has been successfully verified.
          </p>
          <Link href="/login" className="w-full">
            <Button className="w-full">Continue to Login</Button>
          </Link>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout title="Verify Your Email" description="Check your inbox">
      <div className="flex flex-col items-center text-center space-y-6">
        <MaterialIcon name="mail" size="xl" className="text-primary" />
        <p className="text-on-surface-variant text-body-md">
          Click the button below to verify your email address.
        </p>
        <Button
          className="w-full"
          onClick={handleVerify}
          disabled={isVerifying}
        >
          {isVerifying ? (
            <LoadingSpinner size="sm" />
          ) : (
            "Verify Email"
          )}
        </Button>
        <Link
          href="/login"
          className="text-on-surface-variant text-label-lg hover:text-on-surface transition-colors"
        >
          Back to login
        </Link>
      </div>
    </AuthLayout>
  );
}

function LoadingContent() {
  return (
    <AuthLayout title="Verify Your Email" description="Check your inbox">
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner size="md" />
      </div>
    </AuthLayout>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={<LoadingContent />}>
      <VerifyEmailContent />
    </Suspense>
  );
}